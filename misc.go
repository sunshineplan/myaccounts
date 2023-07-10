package main

import "github.com/sunshineplan/database/mongodb"

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

func addUser(username string) error {
	if err := updateUser("add", username); err != nil {
		return err
	}
	svc.Printf("New User %q has been added.", username)
	return nil
}

func deleteUser(username string) error {
	if err := updateUser("delete", username); err != nil {
		return err
	}
	svc.Printf("User %q has been deleted.", username)
	return nil
}
