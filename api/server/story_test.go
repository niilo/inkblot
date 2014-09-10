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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/niilo/golib/test/dockertest"
	"gopkg.in/mgo.v2"
)

func TestHandleIndexReturnsWithStatusOK(t *testing.T) {
	// SetupMongoContainer may skip or fatal the test if docker isn't found or something goes
	// wrong when setting up the container. Thus, no error is returned
	containerID, ip := dockertest.SetupMongoContainer(t)
	fmt.Printf("Mongo docker container %v running on %v ip", containerID, ip)
	defer containerID.KillRemove(t)

	app := AppContext{}
	mongoSession, err := mgo.Dial(ip)
	fmt.Print("connecting to mongodb")
	if err != nil {
		Error.Printf("MongoDB connection failed, with address '%s'.", Configuration.MongoUrl)
	}
	defer mongoSession.Close()

	mongoSession.SetMode(mgo.Monotonic, true)
	app.mongoSession = mongoSession

	req, _ := http.NewRequest("GET", "/story/a8s7df87s", nil)
	w := httptest.NewRecorder()
	prm := httprouter.Param{
		Key:   "id",
		Value: "a8s7df87s",
	}
	p := make(httprouter.Params, 1, 1)
	p[0] = prm

	app.getStory(w, req, p)

	if w.Code != http.StatusOK {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", "200", w.Code)
	}
}
