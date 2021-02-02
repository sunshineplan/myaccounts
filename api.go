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
	var login struct {
		Username, Password string
		Rememberme         bool
	}
	if err := c.BindJSON(&login); err != nil {
		c.String(400, "")
		return
	}
	login.Username = strings.ToLower(login.Username)

	statusCode := 200
	var message string
	user, err := getUserByName(login.Username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			statusCode = 403
			message = "Incorrect username"
		} else {
			log.Print(err)
			statusCode = 500
			message = "Critical Error! Please contact your system administrator."
		}
	} else {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
			if (err == bcrypt.ErrHashTooShort && user.Password != login.Password) ||
				err == bcrypt.ErrMismatchedHashAndPassword {
				statusCode = 403
				message = "Incorrect password"
			} else if user.Password != login.Password {
				log.Print(err)
				statusCode = 500
				message = "Critical Error! Please contact your system administrator."
			}
		}
		if message == "" {
			session := sessions.Default(c)
			session.Clear()
			session.Set("ID", user.ID)
			session.Set("Username", user.Username)

			options := sessions.Options{
				Path:     "/",
				Domain:   domain,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			}

			if login.Rememberme {
				options.MaxAge = 856400 * 30
			}

			session.Options(options)
			if err := session.Save(); err != nil {
				log.Print(err)
				statusCode = 500
				message = "Failed to save session."
			} else {
				message = "OK"
			}
		}
	}
	c.String(statusCode, message)
}

func setting(c *gin.Context) {
	var setting struct{ Password, Password1, Password2 string }
	if err := c.BindJSON(&setting); err != nil {
		c.String(400, "")
		return
	}

	session := sessions.Default(c)
	userID := session.Get("ID")

	user, err := getUserByID(userID.(string))
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	var message string
	var errorCode int
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(setting.Password))
	switch {
	case err != nil:
		if (err == bcrypt.ErrHashTooShort && setting.Password != user.Password) ||
			err == bcrypt.ErrMismatchedHashAndPassword {
			message = "Incorrect password."
			errorCode = 1
		} else if setting.Password != user.Password {
			log.Print(err)
			c.String(500, "")
			return
		}
	case setting.Password1 != setting.Password2:
		message = "Confirm password doesn't match new password."
		errorCode = 2
	case setting.Password1 == setting.Password:
		message = "New password cannot be the same as your current password."
		errorCode = 2
	case setting.Password1 == "":
		message = "New password cannot be blank."
	}

	if message == "" {
		newPassword, err := bcrypt.GenerateFromPassword([]byte(setting.Password1), bcrypt.MinCost)
		if err != nil {
			log.Print(err)
			c.String(500, "")
			return
		}
		if err = changePassword(userID.(string), string(newPassword)); err != nil {
			log.Print(err)
			c.String(500, "")
			return
		}
		session.Clear()
		if err := session.Save(); err != nil {
			log.Print(err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}
