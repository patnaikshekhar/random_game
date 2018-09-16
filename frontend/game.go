package main

import (
	"database/sql"
	"math/rand"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

// Coord defines a coordinate
type Coord struct {
	X   int
	Y   int
	Hit bool
}

// Ship defines a ship
type Ship struct {
	Size     int
	Location []Coord
	Sunk     bool
}

// GameBoard stores the state of the board
type GameBoard struct {
	Coords []Coord
}

// EventName defines the type of event that can be triggered
type EventName int

const (
	// ConnectedEvent Emited from Server to Client letting the client know which server
	ConnectedEvent EventName = 0

	// JoinEvent Emited from Client to Server letting the server know that the Client wants to join
	JoinEvent EventName = 1

	// GameStartedEvent Emited from Server to Client letting the client know they should start placing ships
	GameStartedEvent EventName = 2

	// YourTurnEvent Emited from Server to Client letting the client know whos move it is
	YourTurnEvent EventName = 3

	// PlaceShipsEvent emitted from Client to Server letting the server know where to place ships
	PlaceShipsEvent EventName = 4

	// MakeMoveEvent makes a move for a player
	MakeMoveEvent EventName = 5

	// MoveResultEvent lets the user know the result of the last move
	MoveResultEvent EventName = 6

	// ErrorEvent for generic errors
	ErrorEvent EventName = 7
)

type GameStartedEventMessage struct {
	GameID int
}

type PlaceShipsEventMessage struct {
	EventMessage
	Ships []Ship
}

type MakeMoveEventMessage struct {
	Location Coord
}

type MoveResultEventMessage struct {
	Location Coord
	Outcome  MoveOutcome
	MyBoard  GameBoard
	HitBoard GameBoard
}

type MoveOutcome int

const (
	OutcomeWon      MoveOutcome = 1
	OutcomeLost     MoveOutcome = 2
	OutcomeShipSunk MoveOutcome = 3
	OutcomeShipHit  MoveOutcome = 4
	OutcomeShipMiss MoveOutcome = 5
)

type ErrorEventMessage struct {
	Err string
}

/*
JoinGame checks redis to see if anyone is waiting for a game
if someone is then it pairs the two people and then creates a game
Once the game is created then it informs the other sockets via Kafka
*/
func JoinGame(db *sql.DB, client *redis.Client, producer *kafka.Producer, conn *websocket.Conn, userID int) {

	// Check Redis to see if someone is waiting
	userWaiting := CheckIfSomeOneIsWaiting(client)

	// ---- If no one is waiting then add to cache
	if userWaiting == -1 {
		AddToEndOfQueue(client, userID)
		return
	}

	// ---- If someone is waiting
	// -------- Pop person out of redis - Done in CheckIfSomeOneIsWaiting

	// -------- Create a game in postgres
	gameID := CreateNewGame(db, userID, userWaiting)

	gameUpdateMessagePlayer1 := EventMessage{
		Event: GameStartedEvent,
		To:    userID,
		Payload: GameStartedEventMessage{
			GameID: gameID,
		},
	}

	gameUpdateMessagePlayer1.Send(producer)

	gameUpdateMessagePlayer2 := EventMessage{
		Event: GameStartedEvent,
		Payload: GameStartedEventMessage{
			GameID: gameID,
		},
		To: userWaiting,
	}

	gameUpdateMessagePlayer2.Send(producer)
}

// PlaceShips places the ships on the board and randomly emits a player who will start
func PlaceShips(db *sql.DB, cache *redis.Client, producer *kafka.Producer, message PlaceShipsEventMessage, userID int) {

	// Look for the latest game by this player
	gameID := FindLatestGameForPlayer(db, userID)

	if gameID == -1 {
		PublishErrorEvent(producer, "Could not find game", userID)
		return
	}

	// Create Ships in Database
	err := CreateShipsInDatabase(db, userID, gameID, message.Ships)

	if err != nil {
		PublishErrorEvent(producer, err.Error(), userID)
		return
	}

	// If Both Players have placed ships
	if HaveBothPlayersPlacedShips(db, gameID) {
		// -- Pick a random player
		randomPlayer := rand.Intn(2)
		playerID := getPlayerForGame(db, gameID, randomPlayer)

		// -- Emit the Your Turn Event for that player
		gameUpdateMessagePlayer := EventMessage{
			Event: YourTurnEvent,
			To:    playerID,
		}

		gameUpdateMessagePlayer.Send(producer)
	}
}

/*
PublishErrorEvent sends an error message to a client
*/
func PublishErrorEvent(producer *kafka.Producer, err string, playerID int) {

	gameUpdateMessage := EventMessage{
		Event: ErrorEvent,
		Payload: ErrorEventMessage{
			Err: err,
		},
		To: playerID,
	}

	gameUpdateMessage.Send(producer)
}
