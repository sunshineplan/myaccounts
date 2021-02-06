package main

import (
	"log"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/utils/password"
	"go.mongodb.org/mongo-driver/mongo"
)

func login(c *gin.Context) {
	var data struct {
		Username, Password string
		Rememberme         bool
	}
	if err := c.BindJSON(&data); err != nil {
		c.String(400, "Bad Request")
		return
	}
	data.Username = strings.TrimSpace(strings.ToLower(data.Username))

	var message string
	user, err := getUserByName(data.Username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			message = "Incorrect username"
		} else {
			log.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
	} else {
		ok, err := password.Compare(user.Password, data.Password, false)
		if err != nil {
			log.Print(err)
			c.String(500, "Internal Server Error")
			return
		} else if !ok {
			message = "Incorrect password"
		}

		if message == "" {
			session := sessions.Default(c)
			session.Clear()
			session.Set("id", user.ID.Hex())
			session.Set("username", user.Username)

			options := sessions.Options{
				Domain:   domain,
				HttpOnly: true,
			}

			if data.Rememberme {
				options.MaxAge = 60 * 60 * 24 * 30
			} else {
				options.MaxAge = 60 * 60 * 12
			}

			session.Options(options)
			if err := session.Save(); err != nil {
				log.Print(err)
				c.String(500, "Internal Server Error")
				return
			}

			c.JSON(200, gin.H{"status": 1})
			return
		}
	}

	c.JSON(200, gin.H{"status": 0, "message": message})
}

func chgpwd(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("id")
	if userID == nil {
		c.String(401, "")
		return
	}

	var data struct{ Password, Password1, Password2 string }
	if err := c.BindJSON(&data); err != nil {
		c.String(400, "Bad Request")
		return
	}

	user, err := getUserByID(userID)
	if err != nil {
		log.Print(err)
		c.String(500, "Internal Server Error")
		return
	}

	var message string
	var errorCode int
	newPassword, err := password.Change(user.Password, data.Password, data.Password1, data.Password2, false)
	if err != nil {
		message = err.Error()
		switch err {
		case password.ErrIncorrectPassword:
			errorCode = 1
		case password.ErrConfirmPasswordNotMatch, password.ErrSamePassword:
			errorCode = 2
		case password.ErrBlankPassword:
		default:
			log.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
	}

	if message == "" {
		if err := changePassword(userID, newPassword); err != nil {
			log.Print(err)
			c.String(500, "Internal Server Error")
			return
		}

		session.Clear()
		session.Options(sessions.Options{
			Domain: domain,
			MaxAge: -1,
		})
		if err := session.Save(); err != nil {
			log.Print(err)
			c.String(500, "Internal Server Error")
			return
		}

		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}
