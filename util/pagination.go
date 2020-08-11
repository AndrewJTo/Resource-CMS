package util

import(
	"strconv"
	"context"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetNext(c *gin.Context) primitive.ObjectID {

	nxt := c.Query("next")

	if nxt == "" {
		return primitive.NilObjectID
	}

	next, err := primitive.ObjectIDFromHex(nxt)

	if err != nil {
		return primitive.NilObjectID
	}

	return next

}

func GetPageSize(c *gin.Context) int64 {

	sizeQuery := c.Query("page_size")

	if sizeQuery == "" {
		return 20
	}

	size, err := strconv.ParseInt(sizeQuery, 10, 64)

	if err != nil {
		return 0
	}

	return size

}

func Paginate(c *mongo.Collection, startID primitive.ObjectID, perPage int64) ([]bson.Raw, primitive.ObjectID, error){

	filter := bson.M{"_id": bson.M{"$gt": startID}}

	findOptions := options.Find()
	findOptions.SetLimit(perPage + 1)
	findOptions.SetSort(bson.M{"_id": 1})	//Ascending

	cursor, _ := c.Find(context.Background(), filter, findOptions)

	var results []bson.Raw
	var lastID primitive.ObjectID

	for cursor.Next(context.Background()) {
		var e bson.D
		err := cursor.Decode(&e)
		if err != nil {
			return results, lastID, err
		}
		results = append(results, cursor.Current)
		lastID = cursor.Current.Lookup("_id").ObjectID()
	}

	return results, lastID, nil
}
