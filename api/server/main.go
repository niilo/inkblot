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
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/dmotylev/nutrition"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/niilo/golib/http/handlers"
	nio "github.com/niilo/golib/io"
	"github.com/niilo/golib/smtp"
	"gopkg.in/mgo.v2"
)

// AppContext contains our local context; our database pool, session store, template
// registry and anything else our handlers need to access. We'll create an instance of it
// in our main() function and then explicitly pass a reference to it for our handlers to access.
type AppContext struct {
	//mongoSession *mgo.Session
	mongo      MongoContext
	smtpServer *smtp.SmtpServer
}

type MongoContext struct {
	session *mgo.Session
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
func recoverHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				Error.Printf("Recovering from error '%s'", e)
				Trace.Printf(string(debug.Stack()))
				http.Error(w, http.StatusText(500), 500)
				return
			}
		}()
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func requestLogHandler(h http.Handler) http.Handler {
	rollingWriter, err := nio.NewRollingFileWriterTime(Configuration.RequestLog, nio.RollingArchiveNone, "", 2, "2006-01-02", nio.RollingIntervalDaily)
	if err != nil {
		fmt.Errorf("Request logger creation failed for %s", err.Error())
	}
	logHandler := handlers.NewExtendedLogHandler(h, rollingWriter)

	fn := func(w http.ResponseWriter, req *http.Request) {
		logHandler.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	confFile := flag.String("conf", "inkblot.cfg", "Full path to configuration file")

	err := nutrition.Env("INKBLOT_").File(*confFile).Feed(&Configuration)
	if err != nil {
		log.Fatalf("[inkblot] Unable to read properties:%v\n", err)
	}

	CreateRollingApplicationLoggers(Configuration.AppLog)
}

func (appContext *AppContext) createRoutes() *httprouter.Router {
	router := httprouter.New()
	router.POST("/story", appContext.createStory)
	router.GET("/story/:id", appContext.getStory)
	router.GET("/story/:id/comments", appContext.getStoryComments)
	router.POST("/story/:id/comment", appContext.createStoryComment)
	router.PUT("/comment/:id/like", appContext.likeComment)
	router.PUT("/comment/:id/hate", appContext.hateComment)
	router.PUT("/comment/:id/abuse", appContext.getStory)
	router.POST("/user", appContext.CreateUser)
	router.GET("/user/:id", appContext.GetUser)
	return router
}

func main() {

	Info.Print("Initializin server")

	appContext := AppContext{}

	appContext.smtpServer = &smtp.SmtpServer{
		Host:     Configuration.SmtpHost,
		Port:     Configuration.SmtpHostPort,
		Username: Configuration.SmtpUser,
		Passwd:   Configuration.SmtpUserPwd,
	}

	mongoSession, err := mgo.Dial(Configuration.MongoUrl)
	if err != nil {
		Error.Printf("MongoDB connection failed, with address '%s'.", Configuration.MongoUrl)
	}

	defer mongoSession.Close()

	mongoSession.SetSocketTimeout(Configuration.HandlerTimeout)
	mongoSession.SetMode(mgo.Monotonic, true)
	appContext.mongo.session = mongoSession

	chain := alice.New(requestLogHandler, timeoutHandler, recoverHandler, corsHandler).Then(appContext.createRoutes())

	Info.Printf("Listening on %s", Configuration.ServerAddr)
	s := &http.Server{
		Addr:           Configuration.ServerAddr,
		Handler:        chain,
		ReadTimeout:    Configuration.ReadTimeout,
		WriteTimeout:   Configuration.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())

}
