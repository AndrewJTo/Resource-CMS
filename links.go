package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"

	stru "github.com/AndrewJTo/Resource-CMS/structures"
)

func (s *Server) RemoveLink(c *gin.Context) {
	group, _ := GetSessionGroup(c)
	if !group.UserAdmin {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "msg": "Must be admin to remove page"})
		return
	}
	targetID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid ID"})
		return
	}
	filter := bson.M{"_id": targetID}
	_, err = s.db.Collection("links").DeleteOne(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "Clould not delete link: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "msg": "Removed Link"})
}

func (s *Server) AddNewLink(c *gin.Context) {
	group, _ := GetSessionGroup(c)
	if !group.UserAdmin {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "msg": "Must be admin to add link"})
		return
	}
	var newLogonDetails stru.LinkLogon
	err := c.ShouldBindJSON(&newLogonDetails)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid details sent"})
		return
	}
	newLogonDetails.Id = primitive.NewObjectID()
	_, err = s.db.Collection("links").InsertOne(context.Background(), newLogonDetails)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "Clould not insert new link: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "msg": "Added Link"})
}

func (s *Server) UpdateLink(c *gin.Context) {
	var newLogonDetails stru.LinkLogon
	err := c.ShouldBindJSON(&newLogonDetails)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid details sent"})
		return
	}
	group, _ := GetSessionGroup(c)
	if !group.UserAdmin {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "msg": "Must be admin to change other link"})
		return
	}
	targetID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid ID"})
		return
	}
	filter := bson.M{"_id": targetID}
	update := bson.M{"$set": bson.M{"Link": newLogonDetails.Link, "Username": newLogonDetails.Username, "Password": newLogonDetails.Password}}
	err = s.db.Collection("links").FindOneAndUpdate(context.Background(), filter, update).Decode(&newLogonDetails)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Could not change link: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "msg": "Link Changed"})
}

func (s *Server) GetLink(c *gin.Context) {
	var logon stru.LinkLogon
	targetID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid ID"})
		return
	}
	filter := bson.M{"_id": targetID}
	err = s.db.Collection("links").FindOne(context.Background(), filter).Decode(&logon)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "msg": "Link not found"})
		return
	}
	logon.HexId = logon.Id.Hex()
	c.JSON(200, logon)
}

func (s *Server) ListLinks(c *gin.Context) {
	//TODO: Truncate this data
	cur, err := s.db.Collection("links").Find(context.Background(), bson.M{})
	var links []stru.LinkLogon

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "Could not find links"})
		return
	}

	for cur.Next(context.Background()) {
		link := stru.LinkLogon{}
		err := cur.Decode(&link)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "Could not decode logon"})
			return
		}
		link.HexId = link.Id.Hex()
		links = append(links, link)
	}

	c.JSON(200, links)

}
