package structures

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Link struct {
	Location string `json:"location" bson:"location"`
	Text     string `json:"text" bson:"text"`
}

type SideBar struct {
	//Id    primitive.ObjectID `json:"-" bson:"_id"`
	Title string
	Links []Link
	Key   string
}

type LinkLogon struct {
	Id       primitive.ObjectID `json:"-" bson:"_id"`
	HexId    string             `json:"id" bson:"-"`
	Link     Link               `json:link`
	Username string             `json:"username"`
	Password string             `json:"password"`
}
