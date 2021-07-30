package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sunshineplan/utils"
	"github.com/sunshineplan/utils/database/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var config mongodb.Config
var collection *mongo.Collection

func initMongo() error {
	if err := utils.Retry(func() error {
		return meta.Get("account_mongo", &config)
	}, 3, 20); err != nil {
		return err
	}

	client, err := config.Open()
	if err != nil {
		return err
	}

	collection = client.Database(config.Database).Collection(config.Collection)

	return nil
}

func test() error {
	if err := meta.Get("account_mongo", &config); err != nil {
		return err
	}

	_, err := config.Open()
	return err
}

func queryUser(filter interface{}) (user user, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, filter).Decode(&user)
	return
}

func changePassword(id interface{}, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objecdID, err := primitive.ObjectIDFromHex(id.(string))
	if err != nil {
		return err
	}

	if _, err := collection.UpdateOne(
		ctx, bson.M{"_id": objecdID}, bson.M{"$set": bson.M{"password": password}},
	); err != nil {
		return err
	}

	return nil
}

func updateUser(operation, username string) error {
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
			return fmt.Errorf("username %s not found", username)
		}
	} else {
		if _, err := collection.InsertOne(
			ctx, bson.D{{Key: "username", Value: username}, {Key: "password", Value: "123456"}},
		); err != nil {
			return err
		}
	}

	return nil
}
