package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/sunshineplan/utils/mail"
)

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
