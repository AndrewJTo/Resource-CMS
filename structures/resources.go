package structures

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Page struct {
	Title  string      `json:"page_title" bson:"page_title"`
	Text   string      `json:"page_text" bson:"page_text"`
	Access Permissions `json:"permissions" bson:"permissions"`
}

type Permissions struct {
	AllUsersView  bool                 `json:"all_users_view" bson:"all_users_view"`
	ViewGroupIds  []primitive.ObjectID `json:"view_groups" bson:"view_groups"`
	EditGroupsIds []primitive.ObjectID `json:"edit_groups" bson:"edit_groups"`
}

type Node struct {
	Id        primitive.ObjectID `json:"id" bson:"_id"`
	Title     string
	Location  string
	Type      string
	ContentId primitive.ObjectID `json:"content_id", bson:"content_id"`
	ParentId  primitive.ObjectID `json:"parent_id", bson:"parent_id"`
	Access    Permissions        `json:"permissions", bson:"permissions"`
	Url       string             `json:"url", bson:"-"`
}

type FileObject struct {
	Name   string             `json:"file_name" bson:"file_name"`
	Dir    primitive.ObjectID `json:"-" bson:"dir"` //object id of parent dir
	DirHex string             `json:"dir" bson:"-"`
	Awskey string             `json:"-" bson:"aws_key"`
	DlUrl  string             `json:"download_url" bson:"-"`
}
