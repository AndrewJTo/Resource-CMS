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

func (s *Server) ListEvents(c *gin.Context) {
	//Just return all for now
	cur, err := s.db.Collection("events").Find(context.Background(), bson.M{})
	var events []stru.Event
	if err != nil {
		log.Fatal("Could not find events")
	}
	for cur.Next(context.Background()) {
		event := stru.Event{}
		err := cur.Decode(&event)
		if err != nil {
			log.Fatal("Could not decode event!")
		}
		event.HexId = event.Id.Hex()
		events = append(events, event)
	}
	c.JSON(200, events)
}

func (s *Server) GetEvent(c *gin.Context) {
	eID, err := primitive.ObjectIDFromHex(c.Param("event"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid details sent"})
		return
	}
	log.Println(eID.Hex())
	filter := bson.M{"_id": eID}
	var event stru.Event
	err = s.db.Collection("events").FindOne(context.Background(), filter).Decode(&event)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "msg": "Could not find event: " + err.Error()})
		return
	}
	event.HexId = event.Id.Hex()
	c.JSON(200, event)
}

func (s *Server) DeleteEvent(c *gin.Context) {
	group, _ := GetSessionGroup(c)
	if !group.WriteRes {
		c.JSON(401, gin.H{"success": false, "msg": "Must be admin to delete events"})
		return
	}
	eID, err := primitive.ObjectIDFromHex(c.Param("event"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid details sent"})
		return
	}
	filter := bson.M{"_id": eID}
	_, err = s.db.Collection("events").DeleteOne(context.Background(), filter)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "msg": "Could not find event: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "msg": "Event deleted"})
}

func (s *Server) UpdateEvent(c *gin.Context) {
	group, _ := GetSessionGroup(c)
	if !group.WriteRes {
		c.JSON(401, gin.H{"success": false, "msg": "Must be admin to edit events"})
		return
	}
	eID, err := primitive.ObjectIDFromHex(c.Param("event"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid details sent"})
		return
	}
	var event stru.Event
	err = c.ShouldBindJSON(&event)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid data sent: " + err.Error()})
		return
	}
	event.Id = eID
	filter := bson.M{"_id": eID}
	update := bson.D{{"$set", event}}
	_, err = s.db.Collection("events").UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "msg": "Could not update event: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "msg": "Event updated"})
}

func (s *Server) CreateEvent(c *gin.Context) {
	group, _ := GetSessionGroup(c)
	if !group.WriteRes {
		c.JSON(401, gin.H{"success": false, "msg": "Must be admin to create events"})
		return
	}
	var event stru.Event
	err := c.ShouldBindJSON(&event)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "CR: Invalid data sent: " + err.Error()})
		return
	}
	event.Id = primitive.NewObjectID()
	_, err = s.db.Collection("events").InsertOne(context.Background(), event)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "msg": "Could not create event: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "msg": "Event created"})
}
