package main

import (
	"context"
	"encoding/gob"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-contrib/sessions/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	stru "github.com/AndrewJTo/Resource-CMS/structures"
)

func init() {
	gob.Register(&stru.User{})
	gob.Register(&stru.Group{})
}

func main() {

	port := os.Getenv("PORT")
	mongourl := os.Getenv("MONGODB_URI")
	redisUrl := os.Getenv("REDIS_URL")
	domain := os.Getenv("DOMAIN")
	staticDir := os.Getenv("STATIC")
	bucketName := os.Getenv("BUCKET")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	if mongourl == "" {
		log.Fatal("$MONGODB_URI must be set")
	}

	if redisUrl == "" {
		log.Fatal("$REDIS_URL must be set")
	}

	s := Server{}
	if os.Getenv("SECURE") == "TRUE" {
		s.sec = true
		if domain == "" {
			log.Fatal("$SECURE is set, provide $DOMAIN")
		} else {
			s.domain = domain
		}
	} else {
		s.sec = false
	}

	if staticDir == "" {
		log.Println("Static var not set. Setting as Static")
		s.static = "Static"
	} else {
		s.static = staticDir
	}

	clientOptions := options.Client().ApplyURI(mongourl)
	dbClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = dbClient.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("Connected to MongoDB!")
	s.db = dbClient.Database("resourcesys_db")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2")},
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	s.s3svc = s3.New(sess)
	s.bucketName = bucketName

	redisStore, err := redis.NewStore(10, "tcp", redisUrl, "", []byte("secret"))
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("Redis store created!")

	if isFirstTime(&s) {
		firstTimeSetup(&s)
	}

	s.store = redisStore
	s.port = port
	s.init()
}
