package main

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/utils/log"
)

func run() error {
	if *logPath != "" {
		svc.Logger = log.New(*logPath, "", log.LstdFlags)
		gin.DefaultWriter = svc.Logger
		gin.DefaultErrorWriter = svc.Logger
	}

	if err := initMongo(); err != nil {
		return err
	}

	var r struct{ Endpoint, Username, Password, Secret string }
	if err := meta.Get("account_redis", &r); err != nil {
		return err
	}
	store, err := redis.NewStore(10, "tcp", r.Endpoint, r.Username, r.Password, []byte(r.Secret))
	if err != nil {
		return err
	}
	if err := redis.SetKeyPrefix(store, "account_"); err != nil {
		return err
	}

	router := gin.Default()
	router.TrustedPlatform = "X-Real-IP"
	server.Handler = router

	router.Use(sessions.Sessions("universal", store))
	router.Use(cors.New(cors.Config{
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return strings.Contains(origin, *domain)
		},
		MaxAge: 12 * time.Hour,
	}))

	router.POST("/login", login)
	router.POST("/chgpwd", chgpwd)
	router.POST("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("id")
		if userID == nil {
			c.String(200, "nobody")
			return
		}
		session.Clear()
		session.Options(sessions.Options{
			Domain: *domain,
			MaxAge: -1,
		})
		if err := session.Save(); err != nil {
			svc.Print(err)
			c.String(500, "")
			return
		}
		c.String(200, "bye")
	})

	return server.Run()
}
