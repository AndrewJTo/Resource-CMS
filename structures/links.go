package structures

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Link struct {
	Location string `json:"location" bson:"location"`
	Title    string `json:"text" bson:"text"`
}

type SideBar struct {
	Id    primitive.ObjectID `json:"-" bson:"_id"`
	Title string
	Links []Link
}

type SiteLogin struct {
	Id       primitive.ObjectID `json:"id" bson:"_id"`
	Link     Link
	Username string
	Password string
}
