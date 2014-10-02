package main

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	inkblotMessagesCollection = "messages"
)

func (m MongoContext) saveMessageWithNewId(message Message) (string, error) {
	mongoSession := m.session.Clone()
	defer mongoSession.Close()
	c := mongoSession.DB(Configuration.MongoDbName).C(inkblotMessagesCollection)

	message.Id = GetNewId()

	err := c.Insert(message)
	if mgo.IsDup(err) {
		// retry insert with new id
		message.Id, err = m.saveMessageWithNewId(message)
	} else if err != nil {
		Error.Printf("Mongo insert to %s/%s returned '%s'", Configuration.MongoDbName,
			inkblotMessagesCollection, err.Error())
	}
	return message.Id, err
}

func (m MongoContext) getMessage(id string) (Message, error) {
	mongoSession := m.session.Clone()
	defer mongoSession.Close()
	c := mongoSession.DB(Configuration.MongoDbName).C(inkblotMessagesCollection)

	var message Message
	err := c.FindId(id).One(&message)
	if err != nil {
		return Message{}, err
	}
	return message, nil
}

// markMessageRead ignores invalid id's (length < 1) silently
func (m MongoContext) markMessageRead(id string) error {
	if len(id) < 1 {
		mongoSession := m.session.Clone()
		defer mongoSession.Close()
		c := mongoSession.DB(Configuration.MongoDbName).C(inkblotMessagesCollection)

		err := c.UpdateId(id, bson.M{"$set": bson.M{"read": time.Now()}})
		if err != nil {
			Error.Printf("markMessageRead mongo update returned '%s'", err.Error())
			return err
		}
	}
	return nil
}
