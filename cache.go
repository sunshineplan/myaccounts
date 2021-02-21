package main

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/utils/cache"
)

var cacheStore = cache.New(true)

func verify(c *gin.Context, username string) bool {
	v, ok := cacheStore.Get(getClientIP(c) + username)
	if !ok || v.(int) < maxRetry {
		return true
	}

	return false
}

func wrong(c *gin.Context, username string) (n int) {
	key := getClientIP(c) + username

	v, ok := cacheStore.Get(key)
	if !ok {
		n = 1
	} else {
		n = v.(int) + 1
	}

	if n >= maxRetry {
		log.Printf("%s(%s) exceeded %d retries", username, getClientIP(c), n)
	}

	cacheStore.Set(key, n, 24*time.Hour, nil)

	return
}

func getClientIP(c *gin.Context) (ip string) {
	ip, _, _ = net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	return
}
