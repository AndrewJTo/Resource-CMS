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
	ViewGroupIds  []primitive.ObjectID `json:"view_groups" bson:"view_groups"`
	EditGroupsIds []primitive.ObjectID `json:"edit_groups" bson:"edit_groups"`
}

type DirObject struct {
	Name      string             `json:"directory_name" bson:"directory_name"`
	Parent    primitive.ObjectID `json:"-" bson:"parent_dir"`
	ParentHex string             `json:"parent_id" bson:"-"`
}

type FileObject struct {
	Name   string             `json:"file_name" bson:"file_name"`
	Dir    primitive.ObjectID `json:"-" bson:"dir"`
	DirHex string             `json:"dir" bson:"-"`
	Awskey string             `json:"-" bson:"aws_key"`
	DlUrl  string             `json:"download_url" bson:"-"`
}
