package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type user struct {
	Name string `json:"name"`
	City string `json:"city"`
	Age  int    `json:"age"`
}

var collection *mongo.Collection
var client *mongo.Client

func main() {
	var err error
	client, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	app := fiber.New(fiber.Config{
		Prefork: true,
	})

	collection := client.Database("studentDB").Collection("student")
	api := app.Group("/api")
	api.Post("/createProfile", createProfile(collection))
	api.Get("/getUserProfile", getAllUsers)
	api.Put("/updateProfile", updateProfile)
	api.Post("/deleteProfile", deleteProfile)

	err = app.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
}

func createProfile(collection *mongo.Collection) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")

		var person user

		if err := c.BodyParser(&person); err != nil {
			log.Fatal(err)
			return err
		}
		collection := client.Database("studentDB").Collection("student")
		insertresult, err := collection.InsertOne(context.TODO(), person)
		if err != nil {
			log.Fatal(err)
		}
		return c.JSON(insertresult.InsertedID)

	}
}

func getAllUsers(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/json")
	collection := client.Database("studentDB").Collection("student")
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)
	var student []bson.M
	if err := cursor.All(ctx, &student); err != nil {
		log.Fatal(err)
	}

	return c.Status(fiber.StatusOK).JSON(student)
}

func updateProfile(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/json")
	var body user
	e := c.BodyParser(&body)
	if e != nil {
		log.Fatal(e)
	}
	filter := bson.D{{"name", body.Name}}
	collection := client.Database("studentDB").Collection("student")
	update := bson.D{{"$set", bson.D{{"city", body.City}}}}
	var updateddoc user
	err := collection.FindOneAndUpdate(context.TODO(), filter, update).Decode(&updateddoc)
	if err != nil {
		log.Fatal(err)
	}
	return c.JSON(updateddoc)
}

func deleteProfile(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/json")
	params := c.Params("id")
	fmt.Println(params)
	_id, err := primitive.ObjectIDFromHex(params)
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database("studentDB").Collection("student")
	res, err := collection.DeleteOne(context.TODO(), bson.D{{"_id", _id}})
	if err != nil {
		log.Fatal(err)
	}
	return c.JSON(fiber.Map{"deletedcount": res.DeletedCount})
}
