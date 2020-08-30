package main

import (
	"context"
	stru "github.com/AndrewJTo/Resource-CMS/structures"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"strings"
	"time"
)

func (s *Server) NodePathGet(c *gin.Context) {
	path := c.Param("path")

	if path == "/" {
		root, err := s.GetRootNode()
		if err != nil {
			c.JSON(500, gin.H{"success": false, "msg": "Could not find root node! " + err.Error()})
			return
		}
		nodes, err := s.GetDirList(*root)
		if err != nil {
			c.JSON(500, gin.H{"success": false, "msg": "Get root listing! " + err.Error()})
			return
		}
		c.JSON(200, gin.H{"dir": root, "children": nodes})
		return
	} else {
		path := strings.TrimRight(c.Param("path"), "/")
		node, err := s.GetNodeFromPath(path)
		if err != nil {
			c.JSON(404, gin.H{"success": false, "msg": "File not found! " + err.Error()})
			return
		}
		nodes, err := s.GetDirList(node)
		if err != nil {
			c.JSON(500, gin.H{"success": false, "msg": "Get dir listing! " + err.Error()})
			return
		}
		c.JSON(200, gin.H{"dir": node, "children": nodes})
		return
	}

	node, err := s.GetNodeFromPath(path)

	if err != nil {
		c.JSON(404, gin.H{"path": path, "msg": err.Error()})
		return
	}

	c.JSON(200, gin.H{"node": node})
	return

}

func (s *Server) CreateDir(c *gin.Context) {
	//Check parent dir exists
	path := strings.TrimRight(c.Param("path"), "/")
	log.Println("path " + path)
	pathSplit := strings.Split(path, "/")
	newDir := pathSplit[len(pathSplit)-1]
	log.Println("new:" + newDir)
	log.Printf("Path len %d", len(pathSplit))

	parent, err := s.GetNodeFromParts(pathSplit[:len(pathSplit)-1])

	if err != nil {
		c.JSON(404, gin.H{"success": false, "msg": "Parent dir not found " + err.Error()})
		return
	}
	//Ensure the node is a dir
	if parent.Type != "dir" && parent.Type != "root" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Node is not a directory " + err.Error()})
		return
	}
	//Ensure something with this name doesn't exist
	children, err := s.GetDirList(parent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "Could not get parent dir list " + err.Error()})
		return
	}
	for _, n := range children {
		if strings.EqualFold(n.Title, newDir) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Something already has this name in that dir"})
			return
		}
	}
	//All checks passed, make the dir
	newNode := stru.Node{
		Id:        primitive.NewObjectID(),
		Title:     newDir + "/",
		Location:  parent.Location + parent.Title,
		Type:      "dir",
		ContentId: primitive.ObjectID{},
		ParentId:  parent.Id,
		Access: stru.Permissions{
			AllUsersView:  true,
			ViewGroupIds:  []primitive.ObjectID{},
			EditGroupsIds: []primitive.ObjectID{},
		},
		Creation: time.Now(),
	}

	_, err = s.db.Collection("nodes").InsertOne(context.Background(), newNode)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "Insert new node data"})
		return
	}

	c.JSON(200, gin.H{"success": true, "msg": "Created new dir: " + newDir})
	return
}
