package main

import (
	"time"
)

type Message struct {
	Id      string    `bson:"_id" json:"id"`
	From    string    `json:"from"`
	To      string    `json:"to"`
	Created time.Time `json:"created"`
	Read    time.Time `json:"read"`
	Subject string    `json:"subject"`
	Content string    `json:"content"`
	Ip      string    `json:"ip"`
}
