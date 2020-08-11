package main

import (
	"errors"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	stru "github.com/AndrewJTo/Resource-CMS/structures"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")

		if user == nil {
			c.JSON(http.StatusForbidden, gin.H{"Error": "Not logged in!"})
			c.Abort()
		}

		group := session.Get("group")

		c.Set("user", user)
		c.Set("group", group)
	}
}

func GetSessionUser(c *gin.Context) (*stru.User, error) {

	u, _ := c.Get("user")

	if u == nil {

		return nil, errors.New("Not logged in")
	}

	return u.(*stru.User), nil
}

func GetSessionGroup(c *gin.Context) (*stru.Group, error) {

	g, _ := c.Get("group")

	if g == nil {

		return nil, errors.New("Not logged in")
	}

	return g.(*stru.Group), nil
}

func LoggedIn(c *gin.Context) bool {
	_, exits := c.Get("user")

	return exits
}

func GetSession(c *gin.Context) {

	u, err := GetSessionUser(c)

	if err != nil {
		c.JSON(401, gin.H{"error": "Not logged in!"})
		return
	}

	g, err := GetSessionGroup(c)

	if err != nil {
		c.JSON(401, gin.H{"error": "Not logged in! (2)"})
		return
	}

	c.JSON(200, gin.H{"status": "Signed in!", "user": u, "group": g})

}
