package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/go-redis/redis"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var allSockets map[int]*websocket.Conn

func SocketHandler(db *sql.DB, cache *redis.Client, producer *kafka.Producer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Printf("In SocketHandler")

		userID := getUserID(cache, r)

		if userID == -1 {
			http.Redirect(w, r, "/login", 302)
			return
		}

		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.Write([]byte("Error " + err.Error()))
			return
		}

		go handleSocketConnection(db, cache, producer, conn, userID)
	}
}

func handleSocketConnection(db *sql.DB, cache *redis.Client, producer *kafka.Producer, conn *websocket.Conn, userID int) {

	// Add to list of sockets
	if allSockets == nil {
		allSockets = make(map[int]*websocket.Conn)
	}

	conn.WriteJSON(EventMessage{
		Event:   ConnectedEvent,
		Payload: os.Getenv("HOSTNAME"),
	})

	allSockets[userID] = conn

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var message EventMessage
		json.Unmarshal(p, &message)

		if message.Event == JoinEvent {
			JoinGame(db, cache, producer, conn, userID)
		}

		if message.Event == PlaceShipsEvent {
			var placeShipsMessage PlaceShipsEventMessage
			json.Unmarshal(p, &placeShipsMessage)
			log.Printf("Here in %v", placeShipsMessage)
			PlaceShips(db, cache, producer, placeShipsMessage, userID)
		}
	}
}
