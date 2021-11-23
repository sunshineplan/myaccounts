package main

import (
	"log"

	"github.com/sunshineplan/database/mongodb/api"
)

type user struct {
	ID       string `json:"_id"`
	Username string
	Password string
}

func getUserByName(username string) (user, error) {
	return queryUser(api.M{"username": username})
}

func getUserByID(id string) (user, error) {
	return queryUser(api.M{"_id": api.ObjectID(id)})
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
