package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	stru "github.com/AndrewJTo/Resource-CMS/structures"
)

func (s *Server) ChangeUser(c *gin.Context) {
	var targetID primitive.ObjectID

	authed := false

	user, _ := GetSessionUser(c)
	if c.Param("id") == "me" || c.Param("id") == user.HexId {
		targetID = user.Id
	} else {
		group, _ := GetSessionGroup(c)
		if !group.UserAdmin {
			c.JSON(401, gin.H{"msg": "Must be admin to change other users"})
			return
		}
		authed = true

	}
	targetID, err := primitive.ObjectIDFromHex(c.Param("id"))

	if err != nil {
		c.JSON(400, gin.H{"msg": "There is a problem with the provided userID!"})
		return
	}

	var ch stru.NewUser
	err = c.ShouldBindJSON(&ch)

	if !authed {
		u, err := s.login(ch.OldEmail, ch.OldPass)
		if err != nil {
			c.JSON(400, err.Error())
			return
		}

		if u.Id != targetID {
			c.JSON(402, gin.H{"msg": "Incorrect login details!"})
			return
		}
		authed = true
	}

	var chEmail, chPass, chFName, chLName, chGroup = false, false, false, false, false

	if ch.Email != "" {
		filter := bson.M{"email_address": ch.Email}
		no, _ := s.db.Collection("logins").CountDocuments(context.Background(), filter)
		if no != 0 {
			c.JSON(400, gin.H{"msg": "Email address already associated with an account!"})
			return
		}
		//Chnage email
		filter = bson.M{"userid": targetID}
		update := bson.D{{"$set", bson.D{{"email_address", ch.Email}}}}
		s.db.Collection("logins").FindOneAndUpdate(context.Background(), filter, update)

		chEmail = true
	}

	if ch.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(ch.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Error:" + err.Error())
		}
		filter := bson.M{"userid": targetID}
		update := bson.D{{"$set", bson.D{{"hash", hash}}}}
		s.db.Collection("logins").FindOneAndUpdate(context.Background(), filter, update)

		chPass = true
	}

	if ch.FirstName != "" {
		filter := bson.M{"_id": targetID}
		update := bson.D{{"$set", bson.D{{"first_name", ch.FirstName}}}}
		s.db.Collection("users").FindOneAndUpdate(context.Background(), filter, update)

		chFName = true
	}

	if ch.LastName != "" {
		filter := bson.M{"_id": targetID}
		update := bson.D{{"$set", bson.D{{"last_name", ch.LastName}}}}
		s.db.Collection("users").FindOneAndUpdate(context.Background(), filter, update)

		chLName = true
	}

	if ch.GroupId != "" {
		//TODO: Need to check id is valid first
		if err != nil {
			c.JSON(501, gin.H{"msg": "Group chaning not implemented !"})
			return
		}
		/*
			filter := bson.M{"_id": targetID}
			update := bson.D{"$set":{"group_id":ch.GroupId}}
			err = s.db.Collection("users").FindOneAndUpdate(context.Background(), filter, update)

			if err != nil {
				log.Fatal("Couldn't chanange group? Eh")
				return
			}
			chGroup = true
		*/
	}

	//Done
	c.JSON(200, gin.H{"msg": "Updated Account", "changed": gin.H{"email": chEmail, "password": chPass, "first_name": chFName, "last_name": chLName, "group": chGroup}})
}

func (s *Server) CreateUser(c *gin.Context) {

	group, _ := GetSessionGroup(c)
	if !group.UserAdmin {
		c.JSON(401, gin.H{"msg": "Must be admin"})
		return
	}

	var newUser stru.NewUser

	err := c.ShouldBindJSON(&newUser)

	if err != nil {
		c.JSON(400, gin.H{"msg": "Data won't bind!"})
		return
	}

	//Check if email exists in system
	filter := bson.M{"email_address": newUser.Email}
	no, _ := s.db.Collection("logins").CountDocuments(context.Background(), filter)

	if no != 0 {
		c.JSON(400, gin.H{"msg": "Email address already associated with an account!"})
		return
	}

	log.Println("Inserting new user")

	//Get group
	groupId, _ := primitive.ObjectIDFromHex(newUser.GroupId)
	var newUserGroup stru.Group
	filter = bson.M{"_id": groupId}
	err = s.db.Collection("groups").FindOne(context.Background(), filter).Decode(&newUserGroup)

	if err != nil {
		c.JSON(400, gin.H{"msg": "Group not found!"})
		return
	}

	//Create super user
	nuser := stru.User{}
	nuser.Id = primitive.NewObjectID()
	nuser.GroupId = groupId
	nuser.Creation = time.Now()
	nuser.FirstName = newUser.FirstName
	nuser.LastName = newUser.LastName

	newUserId, err := s.db.Collection("users").InsertOne(context.Background(), nuser)

	if err != nil {
		log.Fatal("Failed new user creation: " + err.Error())
		return
	}

	//Create user logon
	hash, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Password hashing error:" + err.Error())
		return
	}
	logon := stru.Login{}
	logon.UserId = newUserId.InsertedID.(primitive.ObjectID)
	logon.Email = newUser.Email
	logon.Hash = string(hash)
	_, err = s.db.Collection("logins").InsertOne(context.Background(), logon)

	if err != nil {
		log.Fatal("Failed to insert new user login info")
		return
	}

	c.JSON(200, gin.H{"msg": "New user created", "user_id": logon.UserId.String()})
	return

}

func (s *Server) GetUser(id primitive.ObjectID) (stru.User, error) {
	var user stru.User
	filter := bson.M{"_id": id}
	err := s.db.Collection("users").FindOne(context.Background(), filter).Decode(&user)

	return user, err
}
