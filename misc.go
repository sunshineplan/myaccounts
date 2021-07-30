package main

import (
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type user struct {
	ID       primitive.ObjectID `bson:"_id"`
	Username string
	Password string
}

func getUserByName(username string) (user, error) {
	return queryUser(bson.M{"username": username})
}

func getUserByID(id interface{}) (user, error) {
	objecdID, err := primitive.ObjectIDFromHex(id.(string))
	if err != nil {
		return user{}, err
	}
	return queryUser(bson.M{"_id": objecdID})
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

func backup(file string) {
	log.Print("Start!")
	if err := initMongo(); err != nil {
		log.Fatal(err)
	}

	if err := config.Backup(file); err != nil {
		log.Fatal(err)
	}
	log.Print("Backup Done!")
}

func restore(file string) {
	log.Print("Start!")
	if _, err := os.Stat(file); err != nil {
		log.Fatalln("File not found:", err)
	}

	if err := initMongo(); err != nil {
		log.Fatal(err)
	}

	if err := config.Restore(file); err != nil {
		log.Fatal(err)
	}
	log.Print("Done!")
}
