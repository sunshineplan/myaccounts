package main

import (
	"log"

	"github.com/sunshineplan/database/mongodb"
)

type user struct {
	ID       string `json:"_id" bson:"_id"`
	Username string
	Password string
}

func getUserByName(username string) (user, error) {
	return queryUser(mongodb.M{"username": username})
}

func getUserByID(id mongodb.ObjectID) (user, error) {
	return queryUser(mongodb.M{"_id": id.Interface()})
}

func addUser(username string) {
	if err := updateUser("add", username); err != nil {
		log.Fatal(err)
	}
	log.Printf("New User %q has been added.", username)
}

func deleteUser(username string) {
	if err := updateUser("delete", username); err != nil {
		log.Fatal(err)
	}
	log.Printf("User %q has been deleted.", username)
}
