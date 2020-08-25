package structures

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	Id        primitive.ObjectID `json:"-" bson:"_id"`
	HexId     string             `json:"id" bson:"-"`
	GroupId   primitive.ObjectID `json:"-" bson:"group_id"`
	Creation  time.Time          `json:"creation_date" bson:"creation_date"`
	FirstName string             `json:"first_name" bson:"first_name"`
	LastName  string             `json:"last_name" bson:"last_name"`
}

type Login struct {
	UserId primitive.ObjectID `json:"-" bson:"userid"`
	Email  string             `json:"email_address" bson:"email_address"`
	Hash   string             `json:"-" bson:"hash"`
}

type NewUser struct {
	Email     string `json:"email_address,omitempty"`
	Password  string `json:"password,omitempty"`
	OldPass   string `json:"password_old,omitempty"` //Only for acc changes
	OldEmail  string `json:"email_address_old"`      //Only for acc changes
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	GroupId   string `json:"group_id,omitempty"`
}

type Group struct {
	Id         primitive.ObjectID `json:"-" bson:"_id"`
	HexId      string             `json:"id" bson:"-"`
	Name       string             `json:"group_name" bson:"group_name"`
	WriteRes   bool               `json:"write_resources" bson:"write_resources"`
	WritePages bool               `json:"write_pages" bson:"write_pages"`
	UserAdmin  bool               `json:"user_admin" bson:"user_admin"`
	SiteAdmin  bool               `json:"site_admin" bson:"site_admin"`
	Sudo       bool               `json:"sudo" bson:"sudo"`
}

type LoginDetails struct {
	Email    string `json:"email_address" form:"email" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}
