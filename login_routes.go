package main

import(
	"net/http"
	"log"
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"

	stru "github.com/AndrewJTo/Resource-CMS/structures"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func logout (c *gin.Context) {

	if !LoggedIn(c) {
		c.JSON(401, gin.H{"Error": true, "msg":"Not logged in!!"})
		return
	}

	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.JSON(http.StatusOK, gin.H{"logout": true, "msg":"Logged out, bye!"})
}

func (s *Server) login (email, password string) (stru.User, error) {
	//Find user
	var result stru.Login

	log.Println("Looking for email: " + email)

	filter := bson.M{"email_address": email}
	err := s.db.Collection("logins").FindOne(context.Background(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return stru.User{}, errors.New("Wrong email address")
		}
		return stru.User{}, err
	}

	//Check password
	err = bcrypt.CompareHashAndPassword([]byte(result.Hash), []byte(password))

	if err != nil {
		return stru.User{}, errors.New("Wrong password")
	}

	//Get actual user
	user, err := s.GetUser(result.UserId)

	if err != nil {
		return stru.User{}, errors.New("Could not find user data?")
	}
	user.HexId = user.Id.Hex()

	return user, nil
}

func (s *Server) loginRoute (c *gin.Context) {
	var details stru.LoginDetails

	err := c.ShouldBindJSON(&details)

	if err != nil {
		//TODO: the message isn't very helpful most the time
		c.JSON(http.StatusBadRequest, gin.H{"login": false, "msg": err.Error()})
		return
	}

	if details.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"login": false, "msg": "Email not sent!"})
		return
	}
	if details.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"login": false, "msg": "Password not sent!"})
		return
	}

	user, err := s.login(details.Email, details.Password)

	if err != nil {
		c.JSON(402, gin.H{"login":false, "msg":err.Error()})
		return
	}

	//Get group
	var group stru.Group
	filter := bson.M{"_id": user.GroupId}
	err = s.db.Collection("groups").FindOne(context.Background(), filter).Decode(&group)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"login": false, "msg": "Could not find group data!"})
		return
	}

	//Set session
	session := sessions.Default(c)
	session.Set("user", user)
	session.Set("group", group)	//This might not be a good idea
	err = session.Save()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Session store error!"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"login": true, "user": user, "group": group,})
}
