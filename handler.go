package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
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
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password)); err != nil {
			if (err == bcrypt.ErrHashTooShort && user.Password != data.Password) ||
				err == bcrypt.ErrMismatchedHashAndPassword {
				message = "Incorrect password"
			} else if user.Password != data.Password {
				log.Print(err)
				c.String(500, "Internal Server Error")
				return
			}
		}
		if message == "" {
			session := sessions.Default(c)
			session.Clear()
			session.Set("id", user.ID)
			session.Set("username", user.Username)

			options := sessions.Options{
				Path:     "/",
				Domain:   domain,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
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
	var data struct{ Password, Password1, Password2 string }
	if err := c.BindJSON(&data); err != nil {
		c.String(400, "Bad Request")
		return
	}

	session := sessions.Default(c)
	userID := session.Get("id")

	user, err := getUserByID(userID)
	if err != nil {
		log.Print(err)
		c.String(500, "Internal Server Error")
		return
	}

	var message string
	var errorCode int
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password))
	switch {
	case err != nil:
		if (err == bcrypt.ErrHashTooShort && data.Password != user.Password) ||
			err == bcrypt.ErrMismatchedHashAndPassword {
			message = "Incorrect password."
			errorCode = 1
		} else if data.Password != user.Password {
			log.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
	case data.Password1 != data.Password2:
		message = "Confirm password doesn't match new password."
		errorCode = 2
	case data.Password1 == data.Password:
		message = "New password cannot be the same as your current password."
		errorCode = 2
	case data.Password1 == "":
		message = "New password cannot be blank."
	}

	if message == "" {
		newPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password1), bcrypt.MinCost)
		if err != nil {
			log.Print(err)
			c.String(500, "Internal Server Error")
			return
		}

		if err := changePassword(userID, string(newPassword)); err != nil {
			log.Print(err)
			c.String(500, "Internal Server Error")
			return
		}

		session.Clear()
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
