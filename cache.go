package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/utils/cache"
)

var cacheStore = cache.New(true)

func verify(c *gin.Context, username string) bool {
	v, ok := cacheStore.Get(c.ClientIP() + username)
	if !ok || v.(int) < maxRetry {
		return true
	}

	return false
}

func wrong(c *gin.Context, username string) (n int) {
	ip := c.ClientIP()
	key := ip + username

	v, ok := cacheStore.Get(key)
	if !ok {
		n = 1
	} else {
		n = v.(int) + 1
	}

	if n >= maxRetry {
		log.Printf("%s(%s) exceeded %d retries", username, ip, n)
	}

	cacheStore.Set(key, n, 24*time.Hour, nil)

	return
}
