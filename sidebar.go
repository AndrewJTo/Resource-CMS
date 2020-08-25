package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"

	stru "github.com/AndrewJTo/Resource-CMS/structures"
)

func (s *Server) GetSideBarRoute(c *gin.Context) {

	sideBar, err := s.getSideBar()
	if err != nil {
		log.Println("Creating sidebar")
		sideBar, err = s.makeEmptySideBar()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "Could not find the sidebar!"})
			return
		}
	}

	c.JSON(200, sideBar)
	return
}

func (s *Server) SetSideBarRoute(c *gin.Context) {
	group, _ := GetSessionGroup(c)
	if !group.UserAdmin {
		c.JSON(http.StatusForbidden, gin.H{"msg": "Must be admin to change other users"})
		return
	}

	var sideBar stru.SideBar

	err := c.ShouldBindJSON(&sideBar)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invallid data sent"})
		return
	}
	sideBar.Key = "config"

	err = s.setSideBar(sideBar)

	if err != nil {
		log.Fatal("Could not update sidebar" + err.Error())
		return
	}

	c.JSON(http.StatusOK, sideBar)
	return

}

func (s *Server) AddNewSideBarLink(c *gin.Context) {
	group, _ := GetSessionGroup(c)
	if !group.UserAdmin {
		c.JSON(http.StatusForbidden, gin.H{"msg": "Must be admin to change other users"})
		return
	}

	var link stru.Link

	err := c.ShouldBindJSON(&link)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invallid data sent"})
		return
	}

	sideBar, err := s.getSideBar()
	if err != nil {
		log.Fatal("Could not find sidebar")
		return
	}

	sideBar.Links = append(sideBar.Links, link)

	err = s.setSideBar(sideBar)

	if err != nil {
		log.Fatal("Could not update sidebar" + err.Error())
		return
	}

	c.JSON(http.StatusOK, sideBar)
	return
}

func (s *Server) setSideBar(sideBar stru.SideBar) error {
	filter := bson.M{"key": "config"}
	update := bson.D{{"$set", sideBar}}
	_, err := s.db.Collection("settings").UpdateOne(context.Background(), filter, update)

	return err
}

func (s *Server) getSideBar() (stru.SideBar, error) {
	var sideBar stru.SideBar
	err := s.db.Collection("settings").FindOne(context.Background(), bson.M{"key": "config"}).Decode(&sideBar)

	if err != nil {
		log.Println("Could not find side bar" + err.Error())
		return stru.SideBar{}, err
	}
	return sideBar, nil
}

func (s *Server) makeEmptySideBar() (stru.SideBar, error) {
	links := []stru.Link{stru.Link{Location: "/app/page/home", Text: "Home"}, stru.Link{Location: "/app/links", Text: "Links"}}
	sideBar := stru.SideBar{Title: "main", Links: links, Key: "config"}

	_, err := s.db.Collection("settings").InsertOne(context.Background(), sideBar)

	if err != nil {
		log.Fatal("Failed to add new side bar: " + err.Error())
		return sideBar, err
	}
	return sideBar, nil
}
