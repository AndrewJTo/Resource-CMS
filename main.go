package main

import (
	"context"
	"log"
	"os"
	"fmt"
	"encoding/gob"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-contrib/sessions/redis"

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

	svc := s3.New(sess)
	result, err := svc.ListBuckets(nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("Buckets:")
	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n", aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}

	redisStore, err := redis.NewStore(10, "tcp", redisUrl, "", []byte("secret"))
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("Redis store created!")

	if isFirstTime(&s){
		firstTimeSetup(&s)
	}

	s.store = redisStore
	s.port = port
	s.init()
}
