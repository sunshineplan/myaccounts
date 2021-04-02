package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/sunshineplan/utils/mail"
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

func backup() {
	log.Print("Start!")
	if err := initMongo(); err != nil {
		log.Fatal(err)
	}

	tmpfile, err := ioutil.TempFile("", "tmp")
	if err != nil {
		log.Fatal(err)
	}
	tmpfile.Close()
	if err := config.Backup(tmpfile.Name()); err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	var backup struct {
		From, SMTPServer, Password string
		SMTPServerPort             int
		To                         []string
	}
	if err := meta.Get("account_backup", &backup); err != nil {
		log.Fatal(err)
	}

	if err := (&mail.Dialer{
		Host:     backup.SMTPServer,
		Port:     backup.SMTPServerPort,
		Account:  backup.From,
		Password: backup.Password,
	}).Send(&mail.Message{
		To:          backup.To,
		Subject:     fmt.Sprintf("My Accounts Backup-%s", time.Now().Format("20060102")),
		Attachments: []*mail.Attachment{{Path: tmpfile.Name(), Filename: "database"}},
	}); err != nil {
		log.Fatal(err)
	}
	log.Print("Backup Done!")
}

func restore(file string) {
	log.Print("Start!")
	if file == "" {
		log.Fatal("Restore file is blank.")
	} else {
		if _, err := os.Stat(file); err != nil {
			log.Fatalln("File not found:", err)
		}
	}

	if err := initMongo(); err != nil {
		log.Fatal(err)
	}

	if err := config.Restore(file); err != nil {
		log.Fatal(err)
	}
	log.Print("Done!")
}
