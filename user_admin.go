package main

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	stru "github.com/AndrewJTo/Resource-CMS/structures"
	"github.com/AndrewJTo/Resource-CMS/util"
)

func (s *Server) GetUserInfo(c *gin.Context) {

	user, _ := GetSessionUser(c)
	if c.Param("id") == "me" {
		id := user.Id

		u, err := s.GetUser(id)

		//Update session
		session := sessions.Default(c)
		session.Set("user", user)
		err = session.Save()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Session store error!"})
			return
		}

		c.JSON(200, gin.H{"user": u})

	} else {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error(), "msg": "Invalid ID"})
			return
		}

		u, err := s.GetUser(id)

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error(), "msg": "ID not found"})
			return
		}

		group, _ := GetSessionGroup(c)

		if u.Id == user.Id || group.UserAdmin {
			c.JSON(200, gin.H{"user": u})
			return
		}
		c.JSON(401, gin.H{"msg": "You do not have permission"})
	}
}

func (s *Server) ListUsers(c *gin.Context) {

	group, _ := GetSessionGroup(c)
	if !group.UserAdmin {
		c.JSON(401, gin.H{"msg": "Must be admin"})
		return
	}

	nxt := util.GetNext(c)
	no := util.GetPageSize(c)

	col := s.db.Collection("users")

	raw, last, err := util.Paginate(col, nxt, no)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error(), "msg": "Could not paginate"})
		return
	}

	var users []stru.User

	total := int64(len(raw))

	if total < no {
		no = total
	}

	for i := int64(0); i < no; i++ {
		u := stru.User{}
		err := bson.Unmarshal(raw[i], &u)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error(), "msg": "BSON Marshaling error"})
			return
		}

		users = append(users, u)
	}

	c.JSON(200, gin.H{"next_id": last.Hex(), "users": users})
}
