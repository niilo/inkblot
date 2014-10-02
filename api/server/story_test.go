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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	app.mongo.session = mongoSession

	ts := httptest.NewServer(app.createRoutes())
	defer ts.Close()

	storyId := testCreate(ts, t)
	testGet(ts, storyId, t)
}

var applicationJSON string = "application/json"

func testCreate(ts *httptest.Server, t *testing.T) string {

	postData := strings.NewReader("{\"text\":\"tekstiä\",\"subjectId\":\"k2j34\",\"subjectUrl\":\"www.fi/k2j34\"}")

	res, err := http.Post(ts.URL+"/story", applicationJSON, postData)
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}

	id := string(data)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Non-expected status code: %v\n\tbody: %v, data:%s\n", http.StatusCreated, res.StatusCode, id)
	}
	if res.ContentLength != 5 {
		t.Fatalf("Non-expected content length: %v != %v\n", res.ContentLength, 5)
	}
	return id
}

func testGet(ts *httptest.Server, storyId string, t *testing.T) {

	res, err := http.Get(ts.URL + "/story/" + storyId)
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error(err)
	}

	body := string(data)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Non-expected status code: %v\n\tbody: %v, data:%s\n", http.StatusCreated, res.StatusCode, body)
	}
	if !strings.Contains(body, "{\"storyId\":\""+storyId+"\",") {
		t.Fatalf("Non-expected body content: %v", body)
	}
	if res.ContentLength < 163 && res.ContentLength > 165 {
		t.Fatalf("Non-expected content length: %v < %v, content:\n%v\n", res.ContentLength, 160, body)
	}

}
