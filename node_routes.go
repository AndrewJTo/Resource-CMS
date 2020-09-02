package main

import (
	"context"
	stru "github.com/AndrewJTo/Resource-CMS/structures"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
)

func (s *Server) DeleteObj(c *gin.Context) {
	group, _ := GetSessionGroup(c)
	if !group.WriteRes {
		c.JSON(401, gin.H{"success": false, "msg": "Must be admin to delete objects"})
		return
	}

	fPath := c.Param("path")
	fPath = path.Clean(fPath)
	node, err := s.GetNodeFromPath(fPath)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "path": fPath, "msg": err.Error()})
		return
	}
	if node.Type == "s3_file" {
		//Try s3
		err := s.DeleteNode(node)
		if err != nil {
			c.JSON(500, gin.H{"success": false, "error": "Could not delete s3 object: " + err.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true, "msg": "Deleted s3 Object"})
		return
	} else {
		//delete the dir
		if node.Type == "dir" {
			log.Println("This is a dir, recursive delete")
			recursiveDelete(s, node)
			c.JSON(200, gin.H{"success": true, "msg": "Deleted Dir"})
		}
	}
}

func (s *Server) NodePathGet(c *gin.Context) {
	fPath := c.Param("path")
	fPath = path.Clean(fPath)

	node, err := s.GetNodeFromPath(fPath)

	if err != nil {
		c.JSON(404, gin.H{"success": false, "msg": "File not found"})
		return
	}

	if node.Type == "s3_file" {
		//Try s3
		_, err := s.s3svc.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(s.bucketName),
			Key:    aws.String(strings.TrimLeft(fPath, "/")),
		})
		if err != nil {
			c.JSON(404, gin.H{"success": false, "path": fPath, "msg": err.Error()})
			return
		}
		req, _ := s.s3svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(s.bucketName),
			Key:    aws.String(strings.TrimLeft(fPath, "/")),
		})
		urlStr, err := req.Presign(15 * time.Minute)
		if err != nil {
			c.JSON(500, gin.H{"success": false, "path": fPath, "msg": err.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true, "msg": urlStr})
		return
	}

	if node.Type == "dir" || node.Type == "root" {
		//Include a dir listing
		children, err := s.GetDirList(node)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"path": fPath, "msg": err.Error()})
			return
		}
		c.JSON(200, gin.H{"node": node, "children": children})
		return
	} else {
		c.JSON(200, gin.H{"node": node})
		return
	}
}

type NewObj struct {
	Type string
	Name string
}

func (s *Server) CreateObj(c *gin.Context) {
	group, _ := GetSessionGroup(c)
	if !group.WriteRes {
		c.JSON(401, gin.H{"success": false, "msg": "Must be admin to create objects"})
		return
	}
	fPath := strings.TrimRight(c.Param("path"), "/") //Trim trailing slash
	fPath = path.Clean(fPath)
	dir, file := path.Split(fPath)

	var details NewObj
	err := c.ShouldBindJSON(&details)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "msg": "No info sent"})
		return
	}
	if details.Type == "dir" {
		s.CreateDir(c, dir, file, fPath)
		return
	} else {
		//Creating a file, generate PUT URL
		req, _ := s.s3svc.PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String(s.bucketName),
			Key:    aws.String(fPath),
		})
		str, err := req.Presign(15 * time.Minute)
		if err != nil {
			c.JSON(500, gin.H{"success": false, "msg": err.Error()})
			return
		}
		//Creat DB entry
		parentDir, err := s.GetNodeFromPath(dir)
		if err != nil {
			c.JSON(404, gin.H{"success": false, "msg": "Parent dir not found: " + err.Error()})
			return
		}
		newNode := stru.Node{
			Id:        primitive.NewObjectID(),
			Title:     file,
			Location:  dir,
			Type:      "s3_file",
			ContentId: primitive.ObjectID{},
			ParentId:  parentDir.Id,
			Url:       fPath,
			Access: stru.Permissions{
				AllUsersView:  true,
				ViewGroupIds:  []primitive.ObjectID{},
				EditGroupsIds: []primitive.ObjectID{},
			},
			Creation: time.Now(),
		}
		_, err = s.db.Collection("nodes").InsertOne(context.Background(), newNode)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "Could not insert new node data"})
			return
		}
		c.JSON(200, gin.H{"success": true, "msg": str})
		return
	}
}

func (s *Server) CreateDir(c *gin.Context, dir, file, fPath string) {
	//Get parent dir
	node, err := s.GetNodeFromPath(dir)
	if err != nil {
		c.JSON(404, gin.H{"path": fPath, "msg": err.Error()})
		return
	}
	//Ensure the node is a dir
	if node.Type != "dir" && node.Type != "root" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Node is not a directory " + err.Error()})
		return
	}
	//Ensure an object with this name does not already exist
	children, err := s.GetDirList(node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "Could not get parent dir list " + err.Error()})
		return
	}
	for _, n := range children {
		if strings.EqualFold(n.Title, file) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Something already has this name in that dir"})
			return
		}
	}
	//All checks passed, make the dir
	newNode := stru.Node{
		Id:        primitive.NewObjectID(),
		Title:     file,
		Location:  dir,
		Type:      "dir",
		ContentId: primitive.ObjectID{},
		ParentId:  node.Id,
		Url:       dir + file,
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

	c.JSON(200, gin.H{"success": true, "msg": fPath})
}
