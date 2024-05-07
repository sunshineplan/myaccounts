package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb"
	"github.com/sunshineplan/password"
)

type info struct {
	username any
	ip       string
}

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

	if password.IsMaxAttempts(info{data.Username, c.ClientIP()}) {
		c.JSON(200, gin.H{"status": 0, "message": fmt.Sprintf("Max retries exceeded (%d)", *maxRetry)})
		return
	}

	var message string
	user, err := getUserByName(data.Username)
	if err != nil {
		if err == mongodb.ErrNoDocuments {
			message = "Incorrect username"
		} else {
			svc.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
	} else {
		if err = password.CompareHashAndPassword(info{data.Username, c.ClientIP()}, user.Password, data.Password); err != nil {
			if errors.Is(err, password.ErrIncorrectPassword) {
				message = err.Error()
			} else {
				svc.Print(err)
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
				Domain:   *domain,
				HttpOnly: true,
			}

			if data.Rememberme {
				options.MaxAge = *maxage
			} else {
				options.MaxAge = 60 * 60 * 12
			}

			session.Options(options)
			if err := session.Save(); err != nil {
				svc.Print(err)
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
	userID, username := session.Get("id"), session.Get("username")
	if userID == nil || username == nil {
		c.String(401, "")
		return
	}

	if password.IsMaxAttempts(info{username, c.ClientIP()}) {
		c.JSON(200, gin.H{"status": 0, "message": fmt.Sprintf("Max retries exceeded (%d)", *maxRetry), "error": 1})
		return
	}

	var data struct{ Password, Password1, Password2 string }
	if err := c.BindJSON(&data); err != nil {
		c.String(400, "Bad Request")
		return
	}
	var err error
	if priv != nil {
		data.Password, err = password.DecryptPKCS1v15(priv, data.Password)
		if err != nil {
			svc.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
		data.Password1, err = password.DecryptPKCS1v15(priv, data.Password1)
		if err != nil {
			svc.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
		data.Password2, err = password.DecryptPKCS1v15(priv, data.Password2)
		if err != nil {
			svc.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
	}

	id, _ := client.ObjectID(userID.(string))
	user, err := getUserByID(id)
	if err != nil {
		svc.Print(err)
		c.String(500, "Internal Server Error")
		return
	}

	var message string
	var errorCode int
	switch err = password.CompareHashAndPassword(info{username, c.ClientIP()}, user.Password, data.Password); {
	case errors.Is(err, password.ErrIncorrectPassword):
		message = err.Error()
		errorCode = 1
	case err != nil:
		svc.Print(err)
		c.String(500, "Internal Server Error")
		return
	case data.Password1 != data.Password2:
		message = "confirm password doesn't match new password"
		errorCode = 2
	case data.Password1 == data.Password:
		message = "new password cannot be the same as old password"
		errorCode = 2
	case data.Password1 == "":
		message = "new password cannot be blank"
	}

	if message == "" {
		newPassword, err := password.HashPassword(data.Password1)
		if err != nil {
			svc.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
		if err := changePassword(id, newPassword); err != nil {
			svc.Print(err)
			c.String(500, "Internal Server Error")
			return
		}

		session.Clear()
		session.Options(sessions.Options{
			Domain: *domain,
			MaxAge: -1,
		})
		if err := session.Save(); err != nil {
			svc.Print(err)
			c.String(500, "Internal Server Error")
			return
		}

		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}
