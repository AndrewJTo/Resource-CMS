package main

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	stru "github.com/AndrewJTo/Resource-CMS/structures"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//Recursively find a node given a path
func (s *Server) GetNodeFromPath(path string) (stru.Node, error) {
	log.Println("paht: " + path)
	pathSplit := strings.Split(path, "/")
	log.Printf("path size: %d \n", len(pathSplit))
	return s.GetNodeFromParts(pathSplit)
}

func (s *Server) GetNodeFromParts(parts []string) (stru.Node, error) {
	root, err := s.GetRootNode()
	node := stru.Node{}
	if err != nil {
		return node, err
	}
	return recursiveNodeFind(s, *root, parts)

}

func recursiveNodeFind(s *Server, start stru.Node, path []string) (stru.Node, error) {

	log.Println("Start node: " + start.Title + " type " + start.Type)
	if len(path) == 1 {
		return start, nil
	}
	target := path[0]
	log.Println("target: " + target)
	listing, err := s.GetDirList(start)
	if err != nil {
		return start, err
	}
	//Loop through and find the target
	for _, n := range listing {
		if n.Title == target {
			return recursiveNodeFind(s, n, path[1:])
		}
	}
	return start, errors.New("Not found in dir")
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
