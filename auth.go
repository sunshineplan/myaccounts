package main

import (
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

type user struct {
	ID       string
	Username string
	Password string
}

func getUserByName(username interface{}) (user, error) {
	return queryUser(bson.M{"username": username})
}

func getUserByID(id interface{}) (user, error) {
	return queryUser(bson.M{"_id": id})
}

func changePassword(id, password interface{}) error {
	return updatePassword(id, password)
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
