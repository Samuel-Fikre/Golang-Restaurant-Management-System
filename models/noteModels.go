package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	ID         primitive.ObjectID `bson:"_id"`                          // MongoDB ObjectID
	Text       string             `bson:"text" json:"text"`             // Note text
	Title      string             `bson:"title" json:"title"`           // Note title
	Created_at time.Time          `bson:"created_at" json:"created_at"` // Time of note creation
	Updated_at time.Time          `bson:"updated_at" json:"updated_at"` // Time of last note update
	Note_id    string             `bson:"note_id" json:"note_id"`       // Custom note identifier
}
