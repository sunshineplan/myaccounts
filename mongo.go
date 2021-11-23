package main

import (
	"fmt"

	"github.com/sunshineplan/database/mongodb/api"
	"github.com/sunshineplan/utils"
)

var mongo api.Client

func initMongo() error {
	return utils.Retry(func() error {
		return meta.Get("account_mongo", &mongo)
	}, 3, 20)
}

func test() error {
	return meta.Get("account_mongo", &mongo)
}

func queryUser(filter interface{}) (user user, err error) {
	err = mongo.FindOne(filter, nil, &user)
	return
}

func changePassword(id string, password string) (err error) {
	_, err = mongo.UpdateOne(api.M{"_id": api.ObjectID(id)}, api.M{"$set": api.M{"password": password}}, nil)
	return
}

func updateUser(operation, username string) error {
	if err := initMongo(); err != nil {
		return err
	}

	if operation == "delete" {
		result, err := mongo.DeleteOne(api.M{"username": username})
		if err != nil {
			return err
		}
		if result == 0 {
			return fmt.Errorf("username %s not found", username)
		}
	} else {
		if _, err := mongo.InsertOne(api.M{"username": username, "password": "123456"}); err != nil {
			return err
		}
	}

	return nil
}
