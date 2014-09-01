package main

import (
	"encoding/json"
	"net/http"
	"time"

	"code.google.com/p/go.crypto/bcrypt"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	ID             string `bson:"_id,omitempty"`
	Created        time.Time
	Username       string
	Password       string
	Email          string
	ExternalUserID string
	Posts          int
}

//HashPassword takes a plaintext password and hashes it with bcrypt and sets the
//password field to the hash.
func (u *User) HashPassword(password string) {
	hpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err) //this is a panic because bcrypt errors on invalid costs
	}
	u.Password = string(hpass)
}

func (u *User) CompareHash(password string) (err error) {
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return
}

const (
	userCollection = "users"
)

type userRepo struct {
	Session *mgo.Session
}

func (r *userRepo) getUserWithID(id string) (user User, err error) {
	s := r.Session.Clone()
	defer s.Close()

	c := s.DB(Configuration.MongoDbName).C(userCollection)

	err = c.Find(bson.M{"_id": id}).One(&user)
	return

}

func (r *userRepo) create(u *User) (id string, err error) {
	s := r.Session.Clone()
	defer s.Close()

	c := s.DB(Configuration.MongoDbName).C(userCollection)

	u.ID = GetNewId()
	u.Created = time.Now()
	if err = c.Insert(u); mgo.IsDup(err) {
		r.create(u)
	}
	id = u.ID
	return
}

func (a *AppContext) GetUser(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	if len(id) != 5 {
		http.Error(w, "Invalid id.", http.StatusInternalServerError)
		return
	}

	r := userRepo{Session: a.mongoSession}
	user, err := r.getUserWithID(id)
	if err != nil {
		Error.Printf(err.Error())
		http.Error(w, "Request decoding failed.", http.StatusInternalServerError)
		return
	}
	buf, err := json.Marshal(&user)
	if err != nil {
		Error.Printf(err.Error())
		http.Error(w, "json marshalling failed.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func (a *AppContext) CreateUser(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	user := User{}
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		Error.Printf(err.Error())
		http.Error(w, "Request decoding failed.", http.StatusInternalServerError)
		return
	}
	user.HashPassword(user.Password)
	r := userRepo{Session: a.mongoSession}
	id, err := r.create(&user)
	if err != nil {
		http.Error(w, "User creation failed.", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(id))
}
