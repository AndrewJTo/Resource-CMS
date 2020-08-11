package main

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	stru "github.com/AndrewJTo/Resource-CMS/structures"
)

func (s *Server) GetGroupId(name string) (primitive.ObjectID, error) {

	var id primitive.ObjectID

	filter := bson.D{{"group_name", primitive.Regex{Pattern: name, Options: "i"}}}
	err := s.db.Collection("groups").FindOne(context.Background(), filter).Decode(&id)

	return id, err
}

func (s *Server) GetGroup(id primitive.ObjectID) (stru.Group, error) {
	var group stru.Group

	filter := bson.M{"_id": id}
	err := s.db.Collection("groups").FindOne(context.Background(), filter).Decode(&group)

	return group, err
}
