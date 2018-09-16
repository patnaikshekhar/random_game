package main

import (
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

// SessionTime defines the time to session timeout
const SessionTime = time.Minute * 60

// ConnectCache Connects to the cache (redis)
func ConnectCache() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: os.Getenv("CACHE_URL"),
	})

	_, err := client.Ping().Result()

	if err != nil {
		panic("Cannot connect to redis " + err.Error())
	}

	return client

}

// GenerateSession creates a UUID for the session and returns
func GenerateSession(client *redis.Client, userID int) string {
	sessionID := uuid.New().String()
	client.Set("Session-"+sessionID, userID, SessionTime)
	return sessionID
}

// CheckSession checks the current session
func CheckSession(client *redis.Client, sessionID string) (int, error) {
	userID, err := client.Get("Session-" + sessionID).Result()

	if err != nil {
		return -1, err
	}

	i, err2 := strconv.Atoi(userID)

	return i, err2
}

// CheckIfSomeOneIsWaiting checks if someone is waiting for connection
func CheckIfSomeOneIsWaiting(client *redis.Client) int {
	userID, err := client.LPop("WaitingQueue").Result()
	if err != nil {
		return -1
	}

	i, _ := strconv.Atoi(userID)

	return i
}

// AddToEndOfQueue adds the current user to the end of the queue
func AddToEndOfQueue(client *redis.Client, userID int) {
	client.RPush("WaitingQueue", userID)
}
