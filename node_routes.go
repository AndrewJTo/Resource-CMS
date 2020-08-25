package main

import (
	//"log"
	//"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) NodePathGet(c *gin.Context) {
	path := c.Param("path")

	node, err := s.GetNodeFromPath(path)

	if err != nil {
		c.JSON(404, gin.H{"path": path, "msg": err.Error()})
		return
	}

	c.JSON(200, gin.H{"node": node})
	return

}

/*
func (s *Server) CreateDir(c *gin.Context) {
	//Check parent dir exists
	path := c.Param("path")
	newDir := c.Param("new_dir")

	if newDir == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "New dir name not specified"})
		return
	}

	pNode, err := s.GetNodeFromPath(path)

	if err != nil {
		c.JSON(404, gin.H{"path": path, "msg": err.Error()})
		return
	}

	//Ensure the node is a dir
	if dirNode.Type != "dir" && dirNode.Type != "root" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Node is not a directory"})
		return
	}

	//Ensure a node with the desired name does not exist

}*/
