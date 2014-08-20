package main

import (
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

const (
	inkblotStoryCollection = "stories"
)

func (a *AppContext) getStory(w http.ResponseWriter, req *http.Request, p httprouter.Params) {

	id, err := getIdValidateAndSendError(w, &p)
	if err != nil {
		return
	}

	mongoSession := a.mongoSession.Clone()
	defer mongoSession.Close()

	c := mongoSession.DB(Configuration.MongoDbName).C(inkblotStoryCollection)

	story := Story{}
	err = c.Find(bson.M{"_id": id}).One(&story)
	if err != nil {
		logError("Mongo query from %s/%s returned '%s' for id = %s", Configuration.MongoDbName,
			inkblotStoryCollection, err.Error(), id)
		http.NotFound(w, req)
		return
	}
	story.writeToResponse(w)
}

func getIdValidateAndSendError(w http.ResponseWriter, p *httprouter.Params) (id string, err error) {
	id = p.ByName("id")
	if len(id) < 2 || len(id) > 10 {
		err = errors.New("Story id failed validity check")
		logError("%s : id = %s", err.Error(), id)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	return
}

func (story *Story) writeToResponse(w http.ResponseWriter) {
	buf, err := json.Marshal(&story)
	if err != nil {
		logFatal(err.Error())
		http.Error(w, "json marshalling failed.", http.StatusInternalServerError)
		return
	}
	writeJson(w, &buf)
}

func writeJson(w http.ResponseWriter, buf *[]byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(*buf)
}
