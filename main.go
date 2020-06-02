package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	port := os.Getenv("PORT")
	mongourl := os.Gentenv("DATABASE_URL")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	if mongourl == "" {
		log.Fatal("$MONGOURL must be set")
	}

	clientOptions := options.Client().ApplyURI("mongodb://" + mongourl)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl")
	router.Static("/Static", "static")
	router.Static("/gojs", "gojs")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.tmpl", nil)
	})

	router.Run(":" + port)
}
