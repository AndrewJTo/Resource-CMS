package structures

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Event struct {
	Id    primitive.ObjectID `json:"-" bson:"_id"`
	HexId string             `json:"id" bson:"-"`
	Title string             `json:"title" bson:"title"`
	Note  string             `json:"note" bson:"note"`
	Date  time.Time          `json:"date" bson:"date"`
}
