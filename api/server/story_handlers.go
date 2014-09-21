package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/niilo/golib/http/handlers"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	inkblotStoryCollection   = "stories"
	inkblotCommentCollection = "comments"
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
		Error.Printf("Mongo query from %s/%s returned '%s' for id = %s", Configuration.MongoDbName,
			inkblotStoryCollection, err.Error(), id)
		http.NotFound(w, req)
		return
	}
	story.writeToResponse(w)
}

func (a *AppContext) getStoryComments(w http.ResponseWriter, req *http.Request, p httprouter.Params) {

	id, err := getIdValidateAndSendError(w, &p)
	if err != nil {
		return
	}

	mongoSession := a.mongoSession.Clone()
	defer mongoSession.Close()

	c := mongoSession.DB(Configuration.MongoDbName).C(inkblotCommentCollection)

	var comments []ApiComment
	err = c.Find(bson.M{"storyid": id}).All(&comments)
	if err != nil {
		Error.Printf("Mongo query from %s/%s returned '%s' for storyid = %s", Configuration.MongoDbName,
			inkblotCommentCollection, err.Error(), id)
		http.NotFound(w, req)
		return
	}

	apiComments := ApiComments{}
	apiComments.Comments = comments
	apiComments.StoryId = id

	buf, err := json.Marshal(&apiComments)
	if err != nil {
		Error.Printf(err.Error())
		http.Error(w, "json marshalling failed.", http.StatusInternalServerError)
		return
	}
	writeJson(w, &buf)
}

func (a *AppContext) createStory(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	apiComment := ApiComment{}
	err := json.NewDecoder(req.Body).Decode(&apiComment)
	if err != nil {
		Error.Print(err.Error())
		http.Error(w, "Request decoding failed.", http.StatusInternalServerError)
		return
	}

	story := Story{}
	story.Created = time.Now()
	story.CommentsCount = 1
	story.NewestComment = time.Now()
	story.SubjectId = apiComment.SubjectId
	story.SubjectUrl = apiComment.SubjectUrl
	sid, err := story.insertToMongo(a)
	if err != nil {
		http.Error(w, "Story creation failed.", http.StatusInternalServerError)
	}

	comment := NewComment(sid, &apiComment, req)
	_, err = comment.insertToMongo(a)
	if err != nil {
		http.Error(w, "Comment creation failed.", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(sid))
}

func (a *AppContext) createStoryComment(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	id, err := getIdValidateAndSendError(w, &p)
	if err != nil {
		return
	}

	apiComment := ApiComment{}
	err = json.NewDecoder(req.Body).Decode(&apiComment)
	if err != nil {
		Error.Print(err.Error())
		http.Error(w, "Request decoding failed.", http.StatusInternalServerError)
		return
	}

	comment := NewComment(id, &apiComment, req)
	_, err = comment.insertToMongo(a)
	if err != nil {
		http.Error(w, "Comment creation failed.", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(id))
}

func (a *AppContext) likeComment(w http.ResponseWriter, req *http.Request, p httprouter.Params) {

	id, err := getIdValidateAndSendError(w, &p)
	if err != nil {
		return
	}
	userId := "test"

	mongoSession := a.mongoSession.Clone()
	defer mongoSession.Close()
	c := mongoSession.DB(Configuration.MongoDbName).C(inkblotCommentCollection)

	count, _ := c.Find(bson.M{"_id": id, "LikedBy": userId}).Count()
	// Update only if user has not liked earlier
	if count == 0 {
		// don't really care if updates fail
		_ = c.UpdateId(id, bson.M{"$addToSet": bson.M{"LikedBy": userId}})
		_ = c.UpdateId(id, bson.M{"$inc": bson.M{"likes": 1}})
	}
	w.WriteHeader(http.StatusOK)
}

func (a *AppContext) hateComment(w http.ResponseWriter, req *http.Request, p httprouter.Params) {

	id, err := getIdValidateAndSendError(w, &p)
	if err != nil {
		return
	}
	userId := "test"

	mongoSession := a.mongoSession.Clone()
	defer mongoSession.Close()
	c := mongoSession.DB(Configuration.MongoDbName).C(inkblotCommentCollection)

	count, _ := c.Find(bson.M{"_id": id, "HatedBy": userId}).Count()
	// Update only if user has not liked earlier
	if count == 0 {
		// don't really care if updates fail
		_ = c.UpdateId(id, bson.M{"$addToSet": bson.M{"HatedBy": userId}})
		_ = c.UpdateId(id, bson.M{"$inc": bson.M{"likes": -1}})
	}
	w.WriteHeader(http.StatusOK)
}

func NewComment(storyId string, apiComment *ApiComment, req *http.Request) Comment {
	now := time.Now()
	comment := Comment{}
	comment.StoryId = storyId
	comment.ReplyTo = apiComment.ReplyTo
	comment.Text = apiComment.Text
	comment.Author = apiComment.Author
	comment.Published = now
	comment.Public = true

	audit := AuditTrail{}
	audit.Ip = handlers.GetOriginalSourceIP(req)
	audit.Time = now
	audit.UserId = handlers.GetRemoteUser(req)
	comment.Audit = audit
	return comment
}

//func (comment *Comment) createFrom(story *Story, apiComment *ApiComment) {}

func (comment *Comment) insertToMongo(a *AppContext) (id string, err error) {
	mongoSession := a.mongoSession.Clone()
	defer mongoSession.Close()
	c := mongoSession.DB(Configuration.MongoDbName).C(inkblotCommentCollection)

	id = GetNewId()
	comment.CommentId = id

	err = c.Insert(comment)
	if mgo.IsDup(err) {
		// retry insert with new id
		comment.insertToMongo(a)
	} else if err != nil {
		Error.Printf("Mongo insert to %s/%s returned '%s'", Configuration.MongoDbName,
			inkblotStoryCollection, err.Error())
	}
	return
}

func (story *Story) insertToMongo(a *AppContext) (id string, err error) {
	mongoSession := a.mongoSession.Clone()
	defer mongoSession.Close()

	c := mongoSession.DB(Configuration.MongoDbName).C(inkblotStoryCollection)

	id = GetNewId()
	story.StoryId = id

	err = c.Insert(story)
	if mgo.IsDup(err) {
		// retry insert with new id
		story.insertToMongo(a)
	} else if err != nil {
		Error.Printf("Mongo insert to %s/%s returned '%s'", Configuration.MongoDbName,
			inkblotStoryCollection, err.Error())
	}
	return
}

func getIdValidateAndSendError(w http.ResponseWriter, p *httprouter.Params) (id string, err error) {
	id = p.ByName("id")
	if len(id) != 5 {
		err = errors.New("Id failed validity check")
		Error.Printf("%s : id = %s", err.Error(), id)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	return
}

func (story *Story) writeToResponse(w http.ResponseWriter) {
	buf, err := json.Marshal(&story)
	if err != nil {
		Error.Printf(err.Error())
		http.Error(w, "json marshalling failed.", http.StatusInternalServerError)
		return
	}
	writeJson(w, &buf)
}

func writeJson(w http.ResponseWriter, buf *[]byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(*buf)
}
