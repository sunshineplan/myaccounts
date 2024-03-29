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

	if password.IsMaxAttempts(c.ClientIP() + data.Username) {
		c.JSON(200, gin.H{"status": 0, "message": fmt.Sprintf("Max retries exceeded (%d)", maxRetry)})
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
		if priv == nil {
			_, err = password.Compare(c.ClientIP()+data.Username, user.Password, data.Password, false)
		} else {
			_, err = password.CompareRSA(c.ClientIP()+data.Username, user.Password, data.Password, false, priv)
		}
		if err != nil {
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

	if password.IsMaxAttempts(c.ClientIP() + username.(string)) {
		c.JSON(200, gin.H{"status": 0, "message": fmt.Sprintf("Max retries exceeded (%d)", maxRetry), "error": 1})
		return
	}

	var data struct{ Password, Password1, Password2 string }
	if err := c.BindJSON(&data); err != nil {
		c.String(400, "Bad Request")
		return
	}

	id, _ := client.ObjectID(userID.(string))
	user, err := getUserByID(id)
	if err != nil {
		svc.Print(err)
		c.String(500, "Internal Server Error")
		return
	}

	var message, newPassword string
	var errorCode int
	if priv == nil {
		newPassword, err = password.Change(
			c.ClientIP()+user.Username, user.Password, data.Password, data.Password1, data.Password2, false,
		)
	} else {
		newPassword, err = password.ChangeRSA(
			c.ClientIP()+user.Username, user.Password, data.Password, data.Password1, data.Password2, false, priv,
		)
	}
	if err != nil {
		message = err.Error()
		switch {
		case errors.Is(err, password.ErrIncorrectPassword):
			errorCode = 1
		case err == password.ErrConfirmPasswordNotMatch || err == password.ErrSamePassword:
			errorCode = 2
		case err == password.ErrBlankPassword:
		default:
			svc.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
	}

	if message == "" {
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
