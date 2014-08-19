package main

import (
	"gopkg.in/mgo.v2/bson"
	"html"
	"strings"
	"time"
)

type BaseStory struct {
	Id       string        `json:"id"`
	Created  time.Time     `json:"created"`
	Likes    int           `json:"likes"`
	Comments []BaseComment `json:"comments"`
}

type Story struct {
	Id       string    `bson:"_id" json:"id"`
	Created  time.Time `json:"created" bson:",omitempty"`
	Body     `bson:",inline"`
	Social   `bson:",inline"`
	Abuse    bool `json:"abuse" bson:",omitempty"`
	Ban      `json:",omitempty" bson:",inline"`
	Comments []Comment `json:"comments" bson:",omitempty"`
}

type Body struct {
	Title      string            `json:"title"`
	Titles     map[string]string `json:"titles" bson:",omitempty"`
	Subtitle   string            `json:"subtitle" bson:",omitempty"`
	Images     []Image           `json:"img" bson:",omitempty"`
	Slug       string            `json:"slug" bson:",omitempty"`
	Contents   `bson:",inline"`
	Categories []string  `json:"categories" bson:",omitempty"`
	Tags       []string  `json:"tags" bson:",omitempty"`
	Approved   bool      `json:"approved"`
	Modified   time.Time `json:"modified" bson:",omitempty"`
	Embargo    bool      `json:"embargo"`
}

type Image struct {
	Url          string `json:"url"`
	Caption      string `json:"caption" bson:",omitempty"`
	Photographer string `json:"photographer" bson:",omitempty"`
}

type Contents struct {
	Text      string     `json:"text"`
	Author    string     `json:"author" bson:",omitempty"`
	Copyright string     `json:"copyright" bson:",omitempty"`
	Public    bool       `json:"public"`
	Published time.Time  `json:"published" bson:",omitempty"`
	Audit     AuditTrail `json:"audit"`
}

type BaseComment struct {
	CommentId string    `json:"commentId"`
	Text      string    `json:"text"`
	Author    string    `json:"author"`
	Published time.Time `json:"published"`
	Likes     int       `json:"likes"`
}

type Comment struct {
	CommentId string `json:"commentId"`
	SubjectId string `json:"subjectId" bson:",omitempty"`
	ReplyTo   string `json:"replyTo" bson:",omitempty"`
	Contents  `bson:",inline"`
	Social    `bson:",inline"`
	Abuse     bool `json:"abuse" bson:",omitempty"`
	Ban       `json:",omitempty" bson:",inline"`
}

type Social struct {
	Likes int      `json:"likes"`
	Liked []string `json:"liked" bson:",omitempty"`
	Hated []string `json:"hated" bson:",omitempty"`
}

type Ban struct {
	Banned   bool   `json:"banned" bson:",omitempty"`
	BannedBy string `json:"bannedby,omitempty" bson:",omitempty"`
}

type AuditTrail struct {
	UserId string    `json:"userId"`
	Ip     string    `json:"ip"`
	Time   time.Time `json:"time"`
}

type bodyHistory struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id"`
	StoryId string        `json:"storyId"`
	Body    `bson:",inline"`
}

func (s *Story) ConvertToBaseStory() (b BaseStory) {
	b.Id = s.Id
	b.Likes = s.Likes
	b.Created = s.Created
	comments := []BaseComment{}
	for _, c := range s.Comments {
		if !c.Banned {
			bc := BaseComment{}
			bc.Author = c.Author
			bc.CommentId = c.CommentId
			bc.Likes = c.Likes
			bc.Published = c.Published
			bc.Text = c.Text
			comments = append(comments, bc)
		}
	}
	b.Comments = comments
	return
}

func (c *Comment) FormatCommentQuote() {
	text := html.EscapeString(c.Text)
	if strings.Count(text, "]") != strings.Count(text, "[") {
		return
	}
	tag := "quote"
	cssClassBlockquote := "ib-blockquote"
	cssClassCite := "ib-cite"
	for strings.Contains(text, "["+tag) {
		tagStart := strings.Index(text, "["+tag)
		tagEnd := strings.Index(text, "]") + 1
		// i := strings.Index(c, "]") + 1
		if tagEnd != -1 {
			j := strings.Index(text, "[/"+tag+"]")
			if j != -1 {
				block := text[tagEnd:j]
				// fmt.Printf("\n::xml::\n%s\n", str)
				str := "<blockquote class=\"" + cssClassBlockquote + "\">"
				cite := extractCitateFromQuoteTag(text[tagStart:tagEnd])
				if len(cite) > 0 {
					str += "<cite class=\"" + cssClassCite + "\">" + cite + "</cite>"
				}
				str += block + "</blockquote>"
				e := j + 3 + len(tag)
				text = strings.TrimSpace(text[:tagStart] + str + text[e:])
			} else {
				text = strings.Replace(text, text[tagStart:tagEnd], "", 1)
			}
		}
	}
	c.Text = replaceNewLineWithHtmlBreak(text)
}

func replaceNewLineWithHtmlBreak(text string) string {
	out := strings.Replace(text, "\r\n", "<br>", -1)
	out = strings.Replace(out, "\n", "<br>", -1)
	return out
}

func extractCitateFromQuoteTag(tag string) (replyTo string) {
	tag = strings.Replace(tag, "[", "", -1)
	tag = strings.Replace(tag, "]", "", -1)
	if strings.Contains(tag, "=") {
		replyTo = strings.Split(tag, "=")[1]
	}
	return
}

type commentHistory struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Comment `bson:",inline"`
}
