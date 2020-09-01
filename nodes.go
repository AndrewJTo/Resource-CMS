package main

import (
	"context"
	"errors"
	"log"
	"path"
	"strings"
	"time"

	stru "github.com/AndrewJTo/Resource-CMS/structures"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func recursiveDelete(s *Server, start stru.Node) {
	if start.Type == "dir" {
		children, _ := s.GetDirList(start)
		for _, child := range children {
			recursiveDelete(s, child)
		}
		//We can delete this node now
		if s.DeleteNode(start) != nil {
			log.Println("Could not delete node " + start.Title)
		}
		return
	}
	if s.DeleteNode(start) != nil {
		log.Println("Could not delete node " + start.Title)
	}
}

func (s *Server) DeleteNode(node stru.Node) error {
	filter := bson.M{"_id": node.Id}
	_, err := s.db.Collection("nodes").DeleteOne(context.Background(), filter)
	return err
}

//Recursively find a node given a path
func (s *Server) GetNodeFromPath(fPath string) (stru.Node, error) {
	dir, file := path.Split(fPath)
	//Split dir into parts, excluding first
	parts := strings.Split(dir, "/")[1:]
	root, err := s.GetRootNode()
	if err != nil {
		return *root, err
	}

	return recursiveNodeFind(s, *root, parts, file)
}

func recursiveNodeFind(s *Server, start stru.Node, path []string, target string) (stru.Node, error) {
	log.Printf("Recursion:- START: '%s', PATH: %q, TARGET: '%s'", start.Title, path, target)
	if path[0] == "" {
		//We have reached the directory we are looking for
		if target == "" {
			//We want a dir listing
			return start, nil
		}
		//Return a child node of this directory
		children, err := s.GetDirList(start) //Get dir listing
		if err != nil {
			return start, err
		}
		for _, child := range children { //Loop through child nodes
			if strings.EqualFold(target, child.Title) {
				return child, nil
			}
		}
		//The file has not been found, return error
		return start, errors.New("File not found!")
	}
	//We are looking for a subdir of the curent dir
	children, err := s.GetDirList(start)
	if err != nil {
		return start, err
	}
	log.Printf("Getting dir list: %q\n", children)
	for _, child := range children { //Loop through children
		if strings.EqualFold(path[0], child.Title) {
			return recursiveNodeFind(s, child, path[1:], target) //Step down into dir
		}
	}
	//dir not found
	return start, errors.New("A parent dir was not found")
}

//Pass a directory node to this and it will return its child nodes
func (s *Server) GetDirList(dirNode stru.Node) ([]stru.Node, error) {
	if dirNode.Type != "dir" && dirNode.Type != "root" {
		return nil, errors.New("Node is not a directory")
	}

	var children []stru.Node

	filter := bson.M{"parentid": dirNode.Id}
	cur, err := s.db.Collection("nodes").Find(context.Background(), filter)

	if err != nil {
		return children, err
	}

	if err = cur.All(context.Background(), &children); err != nil {
		return children, err
	}

	return children, nil
}

func (s *Server) GetNode(nID primitive.ObjectID) (stru.Node, error) {
	filter := bson.M{"_id": nID}
	var node stru.Node
	err := s.db.Collection("nodes").FindOne(context.Background(), filter).Decode(&node)

	return node, err
}

func (s *Server) GetRootNode() (*stru.Node, error) {

	if s.rootNode != nil {
		return s.rootNode, nil
	}

	filter := bson.M{"type": "root"}
	var rootNode stru.Node
	err := s.db.Collection("nodes").FindOne(context.Background(), filter).Decode(&rootNode)
	if err != nil {

		//Generate node
		rootNode = stru.Node{
			Id:        primitive.NewObjectID(),
			Title:     "",
			Location:  "/",
			Type:      "root",
			ContentId: primitive.ObjectID{},
			Access: stru.Permissions{
				AllUsersView:  true,
				ViewGroupIds:  []primitive.ObjectID{},
				EditGroupsIds: []primitive.ObjectID{},
			},
			Creation: time.Now(),
		}

		_, err = s.db.Collection("nodes").InsertOne(context.Background(), rootNode)

		if err != nil {
			log.Fatal("Could not insert default root node!")
			return nil, err
		}
	}

	return &rootNode, nil

}
