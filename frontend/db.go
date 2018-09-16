package main

import (
	"database/sql"
	"errors"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/xo/dburl"
)

// Connect connects to the database and returns the database connection
func ConnectDB() *sql.DB {

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		panic("DATABASE_URL not specified")
	}

	db, err := dburl.Open(url)
	if err != nil {
		panic("Cannot connect to database " + err.Error())
	} else {
		return db
	}

}

// ValidateLogin will check the username and password supplied. If it does not match then it will throw an error
func ValidateLogin(db *sql.DB, email string, password string) (int, error) {

	var resultEmail, resultPassword string
	var resultID int

	row := db.QueryRow("SELECT id, email, password FROM Users WHERE email = $1", email)
	err := row.Scan(&resultID, &resultEmail, &resultPassword)

	if err != nil {
		row := db.QueryRow(`
			INSERT INTO Users (email, password)
			VALUES ($1, $2) RETURNING Id
		`, email, password)

		err := row.Scan(&resultID)

		if err != nil {
			return -1, err
		}

		return resultID, nil
	}

	if resultPassword == password {
		return resultID, nil
	}

	return -1, errors.New("Invalid Password")
}

// CreateNewGame creates a new game in the database
func CreateNewGame(db *sql.DB, firstUser int, secondUser int) int {

	var gameID int

	row := db.QueryRow(`
		INSERT INTO GAMES (player1, player2) 
		VALUES ($1, $2) RETURNING Id`,
		firstUser, secondUser)
	err := row.Scan(&gameID)

	if err != nil {
		log.Fatal(err.Error())
		return -1
	}

	return gameID
}

// FindLatestGameForPlayer finds the latest Started game for player
func FindLatestGameForPlayer(db *sql.DB, userID int) int {
	var gameID int

	row := db.QueryRow(`
		SELECT Id FROM GAMES 
		WHERE (player1 = $1 OR player2 = $1)
		AND Status = 'Started'
		ORDER BY Id desc
		LIMIT 1
	`, userID)

	err := row.Scan(&gameID)

	if err != nil {
		log.Fatal(err.Error())
		return -1
	}

	return gameID
}

// CreateShipsInDatabase creates the ships for the players
func CreateShipsInDatabase(db *sql.DB, userID int, gameID int, ships []Ship) error {

	for _, ship := range ships {

		xlocation1 := ship.Location[0].X
		ylocation1 := ship.Location[0].Y
		xlocation2, ylocation2, xlocation3, ylocation3, xlocation4, ylocation4, xlocation5, ylocation5 := -1, -1, -1, -1, -1, -1, -1, -1

		if ship.Size > 1 {
			xlocation2 = ship.Location[1].X
			ylocation2 = ship.Location[1].Y
		}

		if ship.Size > 2 {
			xlocation3 = ship.Location[2].X
			ylocation3 = ship.Location[2].Y
		}

		if ship.Size > 3 {
			xlocation4 = ship.Location[3].X
			ylocation4 = ship.Location[3].Y
		}

		if ship.Size > 4 {
			xlocation5 = ship.Location[4].X
			ylocation5 = ship.Location[4].Y
		}

		_, err := db.Exec(`
			INSERT INTO SHIPS (
				gameID, playerID, size, sunk,
				xlocation1, ylocation1,
				xlocation2, ylocation2,
				xlocation3, ylocation3,
				xlocation4, ylocation4,
				xlocation5, ylocation5
			) VALUES (
				$1, $2, $3, $4,
				$5, $6,
				$7, $8,
				$9, $10,
				$11, $12,
				$13, $14
			)
		`, gameID, userID, ship.Size, ship.Sunk,
			xlocation1, ylocation1,
			xlocation2, ylocation2,
			xlocation3, ylocation3,
			xlocation4, ylocation4,
			xlocation5, ylocation5)

		if err != nil {
			return err
		}
	}

	return nil
}

func HaveBothPlayersPlacedShips(db *sql.DB, gameID int) bool {
	row := db.QueryRow(`
		SELECT COUNT(1)
		FROM
			(SELECT 
				PLAYERID, COUNT(1) as count 
			FROM 
				SHIPS 
			WHERE
				(PLAYERID = (SELECT Player1 FROM GAMES WHERE ID = $1) 
			OR PLAYERID = (SELECT Player2 FROM GAMES WHERE ID = $1)) 
			AND GAMEID = $1 GROUP BY PLAYERID) a;
	`, gameID)

	var count int
	err := row.Scan(&count)

	if err != nil {
		log.Fatalf("Error reading from database %s", err.Error())
		return false
	}

	if count == 2 {
		return true
	} else {
		return false
	}
}

func getPlayerForGame(db *sql.DB, gameID int, playerID int) int {

	var userID int
	var row *sql.Row
	if playerID == 0 {
		row = db.QueryRow("SELECT PLAYER1 FROM GAMES WHERE ID = $1", gameID)
	} else {
		row = db.QueryRow("SELECT PLAYER2 FROM GAMES WHERE ID = $1", gameID)
	}

	err := row.Scan(&userID)

	if err != nil {
		log.Fatalf("Error reading from database %s", err.Error())
		return -1
	}

	return userID
}
