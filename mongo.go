package main

import (
	"fmt"

	"github.com/sunshineplan/database/mongodb"
	"github.com/sunshineplan/database/mongodb/driver"
	"github.com/sunshineplan/utils/retry"
)

var client driver.Client

func initMongo() error {
	if err := retry.Do(func() error {
		return meta.Get("account_mongo", &client)
	}, 3, 20,
	); err != nil {
		return err
	}

	return client.Connect()
}

func test() error {
	return initMongo()
}

func queryUser(filter any) (user user, err error) {
	err = client.FindOne(filter, nil, &user)
	return
}

func changePassword(id mongodb.ObjectID, password string) (err error) {
	_, err = client.UpdateOne(mongodb.M{"_id": id.Interface()}, mongodb.M{"$set": mongodb.M{"password": password}}, nil)
	return
}

func updateUser(operation, username string) error {
	if err := initMongo(); err != nil {
		return err
	}

	if operation == "delete" {
		result, err := client.DeleteOne(mongodb.M{"username": username})
		if err != nil {
			return err
		}
		if result == 0 {
			return fmt.Errorf("username %s not found", username)
		}
	} else {
		if _, err := client.InsertOne(
			struct {
				Username string `json:"username" bson:"username"`
				Password string `json:"password" bson:"password"`
			}{username, "123456"},
		); err != nil {
			return err
		}
	}

	return nil
}
