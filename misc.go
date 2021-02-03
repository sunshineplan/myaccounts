package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/sunshineplan/utils/mail"
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

func backup() {
	log.Print("Start!")
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
		Value struct {
			From, SMTPServer, Password string
			SMTPServerPort             int
			To                         []string
		}
	}
	if err := meta.Get("account_backup", &backup); err != nil {
		log.Fatal(err)
	}

	if err := (&mail.Dialer{
		Host:     backup.Value.SMTPServer,
		Port:     backup.Value.SMTPServerPort,
		Account:  backup.Value.From,
		Password: backup.Value.Password,
	}).Send(&mail.Message{
		To:          backup.Value.To,
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
		log.Fatal("Restore file can not be empty.")
	} else {
		if _, err := os.Stat(file); err != nil {
			log.Fatalln("File not found:", err)
		}
	}
	if err := config.Restore(file); err != nil {
		log.Fatal(err)
	}
	log.Print("Done!")
}
