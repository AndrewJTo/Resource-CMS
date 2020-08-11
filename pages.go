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

	filter := bson.D{{"page_title", primitive.Regex{Pattern: "^" + title + "$", Options: "i"}}}
	var page stru.Page
	err := s.db.Collection("pages").FindOne(context.Background(), filter).Decode(&page)
	if err != nil {
		c.JSON(404, gin.H{"msg": "Page not found!"})
		return
	}

	//Check access reqs
	permission := false
	if !page.Access.AllUsersView {
		group, _ := GetSessionGroup(c)
		for _, gId := range page.Access.ViewGroupIds {
			if gId == group.Id {
				permission = true
			}
		}
	} else {
		permission = true
	}

	if !permission {
		c.JSON(http.StatusForbidden, gin.H{"msg": "You do not have permission to view this page!"})
		return
	}

	c.JSON(200, page)

}

func (s *Server) ListPages(c *gin.Context) {
	//TODO: Truncate this data
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

func (s *Server) EditPage(c *gin.Context) {

	group, _ := GetSessionGroup(c)
	if !group.WritePages {
		c.JSON(http.StatusForbidden, gin.H{"msg": "Must be admin to edit pages"})
		return
	}

	oldTitle := c.Param("title")

	//Check page with this title exists
	filter := bson.D{{"page_title", primitive.Regex{Pattern: "^" + oldTitle + "$", Options: "i"}}}
	var page stru.Page
	err := s.db.Collection("pages").FindOne(context.Background(), filter).Decode(&page)
	if err != nil {
		c.JSON(404, gin.H{"msg": "Page not found!"})
		return
	}

	var newPage stru.Page

	err = c.ShouldBindJSON(&newPage)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invallid data sent"})
		return
	}

	//Validate changes
	if oldTitle != newPage.Title {
		filter := bson.D{{"page_title", primitive.Regex{Pattern: "^" + newPage.Title + "$", Options: "i"}}}
		no, _ := s.db.Collection("pages").CountDocuments(context.Background(), filter)
		if no != 0 {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "A page already has this title"})
			return
		}
	}

	//Currently accept any changes to anything else
	filter = bson.D{{"page_title", primitive.Regex{Pattern: "^" + oldTitle + "$", Options: "i"}}}
	update := bson.D{{"$set", newPage}}
	s.db.Collection("pages").FindOneAndUpdate(context.Background(), filter, update)

	c.JSON(200, gin.H{"msg": "Page edited", "url": "/pages/" + newPage.Title})

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
	filter := bson.D{{"page_title", primitive.Regex{Pattern: "^" + newPage.Title + "$", Options: "i"}}}
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
