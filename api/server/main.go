// The server program issues Google search requests and demonstrates the use of
// the go.net Context API. It serves on port 8080.
//
// The /search endpoint accepts these query params:
//   q=the Google search query
//   timeout=a timeout for the request, in time.Duration format
//
// For example, http://localhost:8080/search?q=golang&timeout=1s serves the
// first few Google search results for "golang" or a "deadline exceeded" error
// if the timeout expires.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/dmotylev/nutrition"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/niilo/golib/context/google"
	"github.com/niilo/golib/context/userip"
	"github.com/niilo/golib/http/handlers"
	nio "github.com/niilo/golib/io"
	"github.com/niilo/golib/smtp"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// AppContext contains our local context; our database pool, session store, template
// registry and anything else our handlers need to access. We'll create an instance of it
// in our main() function and then explicitly pass a reference to it for our handlers to access.
type AppContext struct {
	mongoSession *mgo.Session
	smtpServer   *smtp.SmtpServer
}

var Configuration struct {
	RequestLog                 string
	AppLog                     string
	ServerAddr                 string
	ReadTimeout                time.Duration
	WriteTimeout               time.Duration
	HandlerTimeout             time.Duration
	CorsAllowedOrigin          string
	MongoUrl                   string
	MongoDbName                string
	SmtpHost                   string
	SmtpHostPort               int
	SmtpUser                   string
	SmtpUserPwd                string
	smtpEmailAddressInMessages string
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func timeoutHandler(h http.Handler) http.Handler {
	return http.TimeoutHandler(h, Configuration.HandlerTimeout, "request processing timed out")
}

func corsHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if origin := req.Header.Get("Origin"); origin == Configuration.CorsAllowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers",
				"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}
		// Stop here if its Preflighted OPTIONS request
		if req.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}

// Recoverer is a middleware that recovers from panics, logs the panic (and a
// backtrace), and returns a HTTP 500 (Internal Server Error) status if
// possible.
//
func recoverHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				logInfo("recover")
				debug.PrintStack()
				http.Error(w, http.StatusText(500), 500)
				return
			}
		}()
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

var logger *log.Logger = log.New(os.Stdout, "", 0)

func requestLogHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		rollingWriter, _ := nio.NewRollingFileWriterTime("./inkblot.log", nio.RollingArchiveNone, "", 2, "2006-01-02", nio.RollingIntervalDaily)
		bufferedRollingWriter, _ := nio.NewBufferedWriter(rollingWriter, 10240, 0)
		logHandler := handlers.NewExtendedLogHandler(h, bufferedRollingWriter)
		logHandler.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	confFile := flag.String("conf", "inkblot.cfg", "Full path to configuration file")
	logInfo("loading configuration.")

	err := nutrition.Env("INKBLOT_").File(*confFile).Feed(&Configuration)

	if err != nil {
		log.Fatalf("[inkblot] Unable to read properties:%v\n", err)
	}
}

func main() {

	appContext := AppContext{}

	appContext.smtpServer = &smtp.SmtpServer{
		Host:     Configuration.SmtpHost,
		Port:     Configuration.SmtpHostPort,
		Username: Configuration.SmtpUser,
		Passwd:   Configuration.SmtpUserPwd,
	}

	mongoSession, err := mgo.Dial(Configuration.MongoUrl)
	if err != nil {
		log.Fatalf("MongoDB connection failed, with address '%s'.", Configuration.MongoUrl)
	}

	defer mongoSession.Close()

	mongoSession.SetSocketTimeout(Configuration.HandlerTimeout)
	// Switch the session to a monotonic behavior.
	mongoSession.SetMode(mgo.Monotonic, true)
	appContext.mongoSession = mongoSession

	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/json", appContext.handleJSON)
	router.GET("/hello/:name", Hello)
	router.GET("/search", handleSearch)
	router.POST("/story", appContext.createStory)
	router.GET("/story/:id", appContext.getStory)
	router.POST("/user", appContext.CreateUser)
	router.GET("/user/:id", appContext.GetUser)

	chain := alice.New(requestLogHandler, timeoutHandler, recoverHandler, corsHandler).Then(router)

	//loggerHandler := handlers.NewNCSALoggingHandler(chain, os.Stdout)

	logInfo("Listening on port 8080")
	s := &http.Server{
		Addr:           Configuration.ServerAddr,
		Handler:        chain,
		ReadTimeout:    Configuration.ReadTimeout,
		WriteTimeout:   Configuration.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())

}

// handleSearch handles URLs like /search?q=golang&timeout=1s by forwarding the
// query to google.Search. If the query param includes timeout, the search is
// canceled after that duration elapses.
func handleSearch(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	// ctx is the Context for this handler. Calling cancel closes the
	// ctx.Done channel, which is the cancellation signal for requests
	// started by this handler.
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	timeout, err := time.ParseDuration(req.FormValue("timeout"))
	if err == nil {
		// The request has a timeout, so create a context that is
		// canceled automatically when the timeout expires.
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel() // Cancel ctx as soon as handleSearch returns.

	// Check the search query.
	query := req.FormValue("q")
	if query == "" {
		http.Error(w, "no query", http.StatusBadRequest)
		return
	}

	// Store the user IP in ctx for use by code in other packages.
	userIP, err := userip.FromRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx = userip.NewContext(ctx, userIP)

	// Run the Google search and print the results.
	start := time.Now()
	results, err := google.Search(ctx, query)
	elapsed := time.Since(start)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := resultsTemplate.Execute(w, struct {
		Results          google.Results
		Timeout, Elapsed time.Duration
	}{
		Results: results,
		Timeout: timeout,
		Elapsed: elapsed,
	}); err != nil {
		log.Print(err)
		return
	}
}

type Person struct {
	Name string
	Age  string
}

func (a *AppContext) handleJSON(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	mongoSession := a.mongoSession.Clone()
	defer mongoSession.Close()

	c := mongoSession.DB("test").C("people")
	result := Person{}
	err := c.Find(bson.M{"name": "Ale"}).One(&result)
	if err != nil {
		logInfo(err.Error())
		panic(err)
	}
	buf, _ := json.Marshal(&result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)

}

func logDebug(template string, values ...interface{}) {
	log.Printf("[inkblot] "+template+"\n", values...)
}

func logInfo(template string, values ...interface{}) {
	log.Printf("[inkblot] "+template+"\n", values...)
}

func logError(template string, values ...interface{}) {
	log.Printf("[inkblot] "+template+"\n", values...)
}

func logFatal(template string, values ...interface{}) {
	log.Printf("[inkblot] "+template+"\n", values...)
}

var resultsTemplate = template.Must(template.New("results").Parse(`
<html>
<head/>
<body>
  <ol>
  {{range .Results}}
    <li>{{.Title}} - <a href="{{.URL}}">{{.URL}}</a></li>
  {{end}}
  </ol>
  <p>{{len .Results}} results in {{.Elapsed}}; timeout {{.Timeout}}</p>
</body>
</html>
`))
