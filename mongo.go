package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sunshineplan/utils/database/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var config mongodb.Config
var collection *mongo.Collection

func initMongo() error {
	if err := meta.Get("account_mongo", &config); err != nil {
		log.Fatal(err)
	}

	client, err := config.Open()
	if err != nil {
		return err
	}

	collection = client.Database(config.Database).Collection(config.Collection)

	return nil
}

func queryUser(filter interface{}) (user, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user user
	if err := collection.FindOne(ctx, filter).Decode(&user); err != nil {
		return user, err
	}

	return user, nil
}

func updatePassword(id, password interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"password": password}}); err != nil {
		return err
	}

	return nil
}

func updateUser(operation string, username interface{}) error {
	if err := initMongo(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if operation == "delete" {
		result, err := collection.DeleteOne(ctx, bson.M{"username": username})
		if err != nil {
			return err
		}
		if result.DeletedCount == 0 {
			return fmt.Errorf("Username %s not found", username)
		}
	} else {
		if _, err := collection.InsertOne(ctx, bson.M{"username": username, "password": "123456"}); err != nil {
			return err
		}
	}

	return nil
}
