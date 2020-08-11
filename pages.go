package main

import (
	"context"
	stru "github.com/AndrewJTo/Resource-CMS/structures"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
)

func (s *Server) GetPage(c *gin.Context) {
	title := c.Param("title")

	filter := bson.D{{"page_title", primitive.Regex{Pattern: title, Options: "i"}}}
	no, _ := s.db.Collection("pages").CountDocuments(context.Background(), filter)
	if no != 0 {
		c.JSON(400, gin.H{"msg": "A page already has this title!"})
		return
	}

}

func (s *Server) ListPages(c *gin.Context) {
	cur, err := s.db.Collection("pages").Find(context.Background(), bson.M{})

	var pages []stru.Page

	if err != nil {
		log.Fatal("Could not find pages")
	}

	for cur.Next(context.Background()) {
		page := stru.Page{}

		err := cur.Decode(&page)

		if err != nil {
			log.Fatal("Could not decode page!")
		}
		//Don't show any permission info
		page.Access = stru.Permissions{}

		pages = append(pages, page)
	}

	c.JSON(200, gin.H{"pages": pages})
}

func (s *Server) AddPage(c *gin.Context) {
	group, _ := GetSessionGroup(c)
	if !group.WritePages {
		c.JSON(http.StatusForbidden, gin.H{"msg": "Must be admin to add pages"})
		return
	}

	var newPage stru.Page

	err := c.ShouldBindJSON(&newPage)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invallid data sent"})
		return
	}

	//Check if title has already been used
	filter := bson.D{{"page_title", primitive.Regex{Pattern: newPage.Title, Options: "i"}}}
	no, _ := s.db.Collection("pages").CountDocuments(context.Background(), filter)
	if no != 0 {
		c.JSON(400, gin.H{"msg": "A page already has this title!"})
		return
	}

	//Check if the groups exist
	for _, gid := range newPage.Access.ViewGroupIds {
		_, err := s.GetGroup(gid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "View group not found", "id": gid.Hex()})
			return
		}
	}
	for _, gid := range newPage.Access.EditGroupsIds {
		_, err := s.GetGroup(gid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "Edit group not found", "id": gid.Hex()})
			return
		}
	}

	//Insert into db
	_, err = s.db.Collection("pages").InsertOne(context.Background(), newPage)

	if err != nil {
		log.Fatal("There was an error inserting a new page!")
		return
	}

	c.JSON(200, gin.H{"msg": "Page added"})
	return

}
