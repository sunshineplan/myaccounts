package main

import "go.mongodb.org/mongo-driver/bson"

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

func addUser(username string) error {
	return updateUser("add", username)
}

func deleteUser(username string) error {
	return updateUser("delete", username)
}
