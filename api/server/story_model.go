package main

import (
	"strings"
	"time"
)

type Story struct {
	StoryId       string    `bson:"_id" json:"storyId"`
	Created       time.Time `json:"created"`
	SubjectId     string    `json:"subjectId" bson:",omitempty"`
	SubjectUrl    string    `json:"subjectUrl"`
	CommentsCount int       `json:"commentsCount"`
	NewestComment time.Time `json:"newestComment"`
}

type ApiComment struct {
	CommentId  string    `json:"commentId"`
	ReplyTo    string    `json:"replyTo" bson:",omitempty"`
	Text       string    `json:"text"`
	Author     string    `json:"author"`
	Published  time.Time `json:"published"`
	SubjectId  string    `json:"subjectId" bson:",omitempty"`
	SubjectUrl string    `json:"subjectUrl"`
	Likes      int       `json:"likes"`
}

type Comment struct {
	CommentId     string `bson:"_id" json:"commentId"`
	StoryId       string `json:"storyId"`
	ReplyTo       string `json:"replyTo" bson:",omitempty"`
	AbuseReported bool   `json:"abuseReported" bson:",omitempty"`

	Contents `bson:",inline"`
	Social   `bson:",inline"`
	Ban      `json:",omitempty" bson:",inline"`
}

type Contents struct {
	Text      string     `json:"text"`
	Author    string     `json:"author" bson:",omitempty"`
	Public    bool       `json:"public"`
	Published time.Time  `json:"published" bson:",omitempty"`
	Updated   time.Time  `json:"updated" bson:",omitempty"`
	Audit     AuditTrail `json:"audit"`
}

type AuditTrail struct {
	UserId string    `json:"userId"`
	Ip     string    `json:"ip"`
	Time   time.Time `json:"time"`
}

type Social struct {
	Likes   int      `json:"likes"`
	LikedBy []string `json:"likedBy" bson:",omitempty"`
	HatedBy []string `json:"hatedBy" bson:",omitempty"`
}

type Ban struct {
	Banned    bool   `json:"banned" bson:",omitempty"`
	BannedBy  string `json:"bannedBy,omitempty" bson:",omitempty"`
	BanReason string `json:"banReason,omitempty" bson:",omitempty"`
}

func replaceNewLineWithHtmlBreak(text string) string {
	out := strings.Replace(text, "\r\n", "<br>", -1)
	out = strings.Replace(out, "\r", "<br>", -1)
	out = strings.Replace(out, "\n", "<br>", -1)
	return out
}
