package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	inkblotStoryCollection   = "stories"
	inkblotCommentCollection = "comments"
)

func (m *MongoContext) getStoryCommentsForId(storyId string) (ApiComments, error) {
	mongoSession := m.session.Clone()
	defer mongoSession.Close()

	c := mongoSession.DB(Configuration.MongoDbName).C(inkblotCommentCollection)

	var comments []ApiComment
	err := c.Find(bson.M{"storyid": storyId}).All(&comments)
	if err != nil {
		Error.Printf("Mongo query from %s/%s returned '%s' for storyid = %s", Configuration.MongoDbName,
			inkblotCommentCollection, err.Error(), storyId)
		return ApiComments{}, err
	}

	apiComments := ApiComments{}
	apiComments.Comments = comments
	apiComments.StoryId = storyId
	return apiComments, nil
}

func (m *MongoContext) insertStory(story *Story) (id string, err error) {
	mongoSession := m.session.Clone()
	defer mongoSession.Close()

	c := mongoSession.DB(Configuration.MongoDbName).C(inkblotStoryCollection)

	id = GetNewId()
	story.StoryId = id

	err = c.Insert(story)
	if mgo.IsDup(err) {
		// retry insert with new id
		m.insertStory(story)
	} else if err != nil {
		Error.Printf("Mongo insert to %s/%s returned '%s'", Configuration.MongoDbName,
			inkblotStoryCollection, err.Error())
	}
	return
}

func (m *MongoContext) insertComment(comment *Comment) (id string, err error) {
	mongoSession := m.session.Clone()
	defer mongoSession.Close()
	c := mongoSession.DB(Configuration.MongoDbName).C(inkblotCommentCollection)

	id = GetNewId()
	comment.CommentId = id

	err = c.Insert(comment)
	if mgo.IsDup(err) {
		// retry insert with new id
		m.insertComment(comment)
	} else if err != nil {
		Error.Printf("Mongo insert to %s/%s returned '%s'", Configuration.MongoDbName,
			inkblotStoryCollection, err.Error())
	}
	return
}
