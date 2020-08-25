package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	stru "github.com/AndrewJTo/Resource-CMS/structures"
)

func (s *Server) ListGroups(c *gin.Context) {
	//TODO: Truncate this data
	cur, err := s.db.Collection("groups").Find(context.Background(), bson.M{})

	var groups []stru.Group

	if err != nil {
		log.Fatal("Could not find groups")
	}

	for cur.Next(context.Background()) {
		g := stru.Group{}

		err := cur.Decode(&g)

		if err != nil {
			log.Fatal("Could not decode group!")
		}
		g.HexId = g.Id.Hex()

		groups = append(groups, g)
	}

	c.JSON(200, groups)
}

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

	if err != nil {
		group.HexId = group.Id.Hex()
	}
	return group, err
}
