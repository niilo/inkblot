/*
Copy from: https://github.com/bradfitz/camlistore/blob/master/pkg/blobserver/mongo/mongo_test.go

Copyright 2014 The Camlistore Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/niilo/golib/test/dockertest"
	"gopkg.in/mgo.v2"
)

func TestStoryCreateAndGet(t *testing.T) {
	// SetupMongoContainer may skip or fatal the test if docker isn't found or something goes
	// wrong when setting up the container. Thus, no error is returned
	containerID, ip := dockertest.SetupMongoContainer(t)
	defer containerID.KillRemove(t)

	app := AppContext{}
	mongoSession, err := mgo.Dial(ip)
	if err != nil {
		Error.Printf("MongoDB connection failed, with address '%s'.", Configuration.MongoUrl)
	}
	defer mongoSession.Close()

	mongoSession.SetMode(mgo.Monotonic, true)
	app.mongoSession = mongoSession

	storyId := testCreate(&app, t)
	testGet(&app, storyId, t)
}

func testCreate(app *AppContext, t *testing.T) string {

	postData := strings.NewReader("{\"text\":\"tekstiä\",\"subjectId\":\"k2j34\",\"subjectUrl\":\"www.fi/k2j34\"}")
	req, _ := http.NewRequest("POST", "/story", postData)
	w := httptest.NewRecorder()
	app.createStory(w, req, nil)

	if w.Code != http.StatusCreated {
		t.Fatalf("Non-expected status code:%v\n\tbody: %v, data:%s", http.StatusCreated, w.Code, w.Body.String())
	}
	return w.Body.String()
}

func testGet(app *AppContext, storyId string, t *testing.T) {
	req, _ := http.NewRequest("GET", "/story/"+storyId, nil)
	w := httptest.NewRecorder()
	p := httprouter.Params{
		httprouter.Param{"id", storyId},
	}

	app.getStory(w, req, p)
	if w.Code != http.StatusOK {
		t.Fatalf("Non-expected status code: %v\n\tbody: %v, data:%s", http.StatusOK, w.Code, w.Body.String())
	}
}
