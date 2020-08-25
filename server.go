package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	stru "github.com/AndrewJTo/Resource-CMS/structures"
)

type Server struct {
	domain   string
	port     string
	sec      bool
	router   *gin.Engine
	db       *mongo.Database
	store    redis.Store
	rootNode *stru.Node
}

func (s *Server) init() {

	s.router = gin.New()

	s.router.Use(gin.Logger())
	s.router.Use(sessions.Sessions("ressys_sessions", s.store))

	s.router.LoadHTMLGlob("templates/*.tmpl")
	s.router.Static("/Static", "static")
	s.router.Static("/gojs", "gojs")
	s.router.POST("/api/login", s.loginRoute)

	//Counter
	s.router.GET("/test", func(c *gin.Context) {
		session := sessions.Default(c)
		var count int
		v := session.Get("count")
		if v == nil {
			count = 0
		} else {
			count = v.(int)
			count++
		}
		session.Set("count", count)
		err := session.Save()
		if err != nil {
			c.JSON(501, gin.H{"error": "Sessin save error!", "count": count})
		} else {
			c.JSON(200, gin.H{"count": count})
		}
	})

	//TODO: CSRF STUFF HERE
	auth := s.router.Group("/api")
	auth.Use(AuthMiddleware())
	{
		auth.GET("/getsession", GetSession)
		auth.GET("/logout", logout)
		auth.GET("/users", s.ListUsers)
		auth.GET("/user/:id", s.GetUserInfo)
		auth.POST("/user/:id", s.ChangeUser)
		auth.PUT("/user/new", s.CreateUser)
		auth.GET("/site/sidebar", s.GetSideBarRoute)
		auth.POST("/site/sidebar", s.AddNewSideBarLink)
		auth.PUT("/site/sidebar", s.SetSideBarRoute)
		auth.GET("/pages", s.ListPages)
		auth.PUT("/pages", s.AddPage)
		auth.POST("/pages/:title", s.EditPage)
		auth.DELETE("/pages/:title", s.RemovePage)
		auth.GET("/pages/:title", s.GetPage)
		auth.GET("/groups", s.ListGroups)
		auth.GET("/files/:path", s.NodePathGet)
		auth.GET("/links", s.ListLinks)
		auth.GET("/links/:id", s.GetLink)
		auth.POST("/links/:id", s.UpdateLink)
		auth.DELETE("/links/:id", s.RemoveLink)
		auth.PUT("/links", s.AddNewLink)

	}

	s.router.GET("/", func(c *gin.Context) {
		c.HTML(200, "home.tmpl", nil)
	})

	if s.sec {
		log.Fatal(autotls.Run(s.router, s.domain))
	} else {
		s.router.Run(":" + s.port)
	}
}

func isFirstTime(s *Server) bool {
	//TODO: Defer canceling.
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	no, err := s.db.Collection("logins").CountDocuments(ctx, bson.M{})

	if err != nil {
		log.Fatal("First time use check error: " + err.Error())
	}

	if no > 0 {
		return false
	}
	return true
}

func firstTimeSetup(s *Server) {

	log.Println("Doing first time setup...")

	//Create groups
	users := stru.Group{primitive.NewObjectID(), "", "user", false, false, false, false, false}
	admins := stru.Group{primitive.NewObjectID(), "", "admin", true, true, true, true, false}

	opts := options.InsertMany().SetOrdered(true)
	res, err := s.db.Collection("groups").InsertMany(context.Background(), []interface{}{users, admins}, opts)

	if err != nil {
		log.Fatal("Group creation error: " + err.Error())
	}

	//Create super user
	super := stru.User{}
	super.Id = primitive.NewObjectID()
	super.GroupId = res.InsertedIDs[1].(primitive.ObjectID)
	super.Creation = time.Now()
	super.FirstName = "Super"
	super.LastName = "User"

	userId, err := s.db.Collection("users").InsertOne(context.Background(), super)

	if err != nil {
		log.Fatal("Failed super user creation: " + err.Error())
	}

	//Create super logon
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Error:" + err.Error())
	}
	logon := stru.Login{}
	logon.UserId = userId.InsertedID.(primitive.ObjectID)
	logon.Email = "super"
	logon.Hash = string(hash)
	_, err = s.db.Collection("logins").InsertOne(context.Background(), logon)

	if err != nil {
		log.Fatal("Failed to insert super user login info")
	}

}
