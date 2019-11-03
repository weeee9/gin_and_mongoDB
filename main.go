package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Trainer represent pokemon trainer
type Trainer struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

// Credential for mongoDB atlas credential
type Credential struct {
	User     string `json::"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
}

var trainerCollection *mongo.Collection

func main() {
	client := initMongo("./mongo.json")
	trainerCollection = client.Database("pokemon").Collection("trainers")
	router := gin.Default()

	router.GET("/trainers", getAllTrainers)
	router.GET("/trainer/:name", getTrainerByName)
	router.POST("/trainer", createTrainer)
	router.DELETE("/trainers", deleteAll)

	router.Run()
}

func initMongo(credFilePath string) *mongo.Client {
	var c Credential
	file, err := ioutil.ReadFile(credFilePath)
	if err != nil {
		log.Fatalf("File error: %v\n", err)
	}
	json.Unmarshal(file, &c)
	mongoURI := fmt.Sprintf("mongodb+srv://%s:%s@%s", c.User, c.Password, c.Host)

	// Set client options
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	// client, err := mongo.NewClient(clinetOpt)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	return client
}

func getAllTrainers(c *gin.Context) {
	var results []*Trainer
	// Pass these options to the Find method
	findOptions := options.Find()

	// Passing bson.D{{}} as the filter matches all documents in the collection
	cur, err := trainerCollection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		log.Println("[MongoDB] Error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "[MongoBD] " + err.Error(),
		})
		return
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(context.TODO()) {
		// create a value into which the single document can be decoded
		var elem Trainer
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		log.Println("[MongoDB] Error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"CODE":    http.StatusBadRequest,
			"message": "[MongoBD] " + err.Error(),
		})
		return
	}
	// Close the cursor once finished
	cur.Close(context.TODO())
	log.Printf("Found multiple documents (array of pointers): %+v\n", results)

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"trainers": results,
	})
}

func getTrainerByName(c *gin.Context) {
	var result Trainer
	tName := c.Param("name")
	filter := bson.D{{"name", tName}}

	err := trainerCollection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		log.Println("[MongoDB] Error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "[MongoBD] " + err.Error(),
		})
		return
	}
	log.Printf("Found a single document: %+v\n", result)
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"trainer": result,
	})
}

func createTrainer(c *gin.Context) {
	var newTrainer Trainer
	err := c.BindJSON(&newTrainer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "binding json error",
		})
		return
	}
	insertResult, err := trainerCollection.InsertOne(context.TODO(), newTrainer)
	if err != nil {
		log.Println("[MongoDB] Error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "[MongoBD] " + err.Error(),
		})
		return
	}
	log.Println("Inserted a single document: ", insertResult.InsertedID)
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"trainer": newTrainer,
	})
}

func deleteAll(c *gin.Context) {
	deleteResult, err := trainerCollection.DeleteMany(context.TODO(), bson.D{{}})
	if err != nil {
		log.Println("[MongoDB] Error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "[MongoBD] " + err.Error(),
		})
		return
	}

	fmt.Printf("Deleted %v documents in the trainers collection\n", deleteResult.DeletedCount)
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "all trainers deleted",
	})
}
