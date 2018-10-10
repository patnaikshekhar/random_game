package main

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/xo/dburl"
)

var db *sql.DB

func TestConstructGameUpdateMessage(t *testing.T) {

	tt := []struct {
		name            string
		gameID          int
		playerID        int
		turn            bool
		expectedStatus  GameState
		expectedMyState GameBoard
		// opponentState  GameBoard
	}{
		{"When game is won", 2, 1, true, GameStateWon, constructCoords([]Coord{Coord{0, 1, false}, Coord{0, 2, false}, Coord{0, 3, false}, Coord{1, 1, false}, Coord{2, 1, false}})},
		{"When game is lost", 3, 1, true, GameStateLost, constructCoords([]Coord{Coord{0, 1, false}, Coord{0, 2, false}, Coord{0, 3, false}, Coord{1, 1, false}, Coord{2, 1, false}})},
		{"When game is running and its players turn", 1, 1, true, GameStateMyTurn, constructCoords([]Coord{Coord{0, 1, false}, Coord{0, 2, false}, Coord{0, 3, false}, Coord{1, 1, false}, Coord{2, 1, false}})},
		{"When game is running and its not the players turn", 1, 1, false, GameStateNotMyTurn, constructCoords([]Coord{Coord{0, 1, false}, Coord{0, 2, false}, Coord{0, 3, false}, Coord{1, 1, false}, Coord{2, 1, false}})},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			message := ConstructGameUpdateMessage(db, tc.gameID, tc.playerID, tc.turn)

			if message.Status != tc.expectedStatus {
				t.Fatalf("Expecting game status to be %d but was %d", tc.expectedStatus, message.Status)
			}

			if !deepCheck(tc.expectedMyState, message.MyBoard) {
				t.Fatalf("Expecting MyBoard") // state to be" %v was %v", tc.expectedMyState, message.MyBoard)
			}
		})
	}
}

func deepCheck(expected GameBoard, player GameBoard) bool {

	if len(expected.Coords) != len(player.Coords) {
		return false
	}

	for i, _ := range expected.Coords {
		for j, _ := range expected.Coords[i] {
			if expected.Coords[i][j].X != player.Coords[i][j].X ||
				expected.Coords[i][j].Y != player.Coords[i][j].Y ||
				expected.Coords[i][j].Hit != player.Coords[i][j].Hit {
				return false
			}
		}
	}

	return true
}

func TestMain(m *testing.M) {
	err := SetupDB()
	if err != nil {
		log.Printf("Error in Test Setup %s", err.Error())
		os.Exit(-1)
	}

	retCode := m.Run()

	TearDown()

	os.Exit(retCode)
}

func TearDown() {
	defer db.Close()
	db.Exec("DROP TABLE IF EXISTS SHIPS")
	db.Exec("DROP TABLE IF EXISTS GAMES")
	db.Exec("DROP TABLE IF EXISTS USERS")
}

func SetupDB() error {

	var err error

	db, err = dburl.Open("postgres://localhost:5432/battleship?sslmode=disable")

	if err != nil {
		return err
	}

	_, err = db.Exec("DROP TABLE IF EXISTS SHIPS")
	if err != nil {
		return err
	}

	_, err = db.Exec("DROP TABLE IF EXISTS GAMES")
	if err != nil {
		return err
	}

	_, err = db.Exec("DROP TABLE IF EXISTS USERS")
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE USERS (
			Id bigserial primary key, 
			email text, 
			password text
		);`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO USERS (email, password) 
		VALUES ('1@a.com', '1');`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO USERS (email, password) 
		VALUES ('2@a.com', '2');`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE GAMES (
			Id bigserial primary key, 
			player1 bigint references USERS, 
			player2 bigint references USERS, 
			status text DEFAULT 'Started',
			winner bigint
		)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO GAMES (player1, player2) VALUES (1, 2)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO GAMES (player1, player2, Status, Winner) VALUES (1, 2, 'Completed', 1)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO GAMES (player1, player2, Status, Winner) VALUES (1, 2, 'Completed', 2)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE SHIPS (
			Id bigserial primary key, 
			gameID bigint references GAMES, 
			playerID bigint references USERS, 
			size int,
			xlocation1 smallint DEFAULT -1,
			ylocation1 smallint DEFAULT -1,
			location1hit boolean DEFAULT false,
			xlocation2 smallint DEFAULT -1,
			ylocation2 smallint DEFAULT -1,
			location2hit  boolean DEFAULT false,
			xlocation3 smallint DEFAULT -1,
			ylocation3 smallint DEFAULT -1,
			location3hit  boolean DEFAULT false,
			xlocation4 smallint DEFAULT -1,
			ylocation4 smallint DEFAULT -1,
			location4hit  boolean DEFAULT false,
			xlocation5 smallint DEFAULT -1,
			ylocation5 smallint DEFAULT -1,
			location5hit  boolean DEFAULT false,
			sunk boolean DEFAULT false
		);`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO SHIPS (gameID, playerID, size, sunk, xlocation1, ylocation1, xlocation2, ylocation2, xlocation3, ylocation3) 
		VALUES (
			1, 1, 3, false,
			0, 1, 0, 2, 0, 3
			)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO SHIPS (gameID, playerID, size, sunk, xlocation1, ylocation1, xlocation2, ylocation2) 
		VALUES (
			1, 1, 2, false,
			1, 1, 2, 1
			)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO SHIPS (gameID, playerID, size, sunk, xlocation1, ylocation1, xlocation2, ylocation2, xlocation3, ylocation3) 
		VALUES (
			1, 2, 3, false,
			0, 1, 0, 2, 0, 3
			)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO SHIPS (gameID, playerID, size, sunk, xlocation1, ylocation1, xlocation2, ylocation2) 
		VALUES (
			1, 2, 2, false,
			1, 1, 2, 1
			)`)
	if err != nil {
		return err
	}

	return nil
}

func constructCoords(coords []Coord) GameBoard {

	var result GameBoard

	for i := 0; i < GameBoardSize; i++ {
		var row []Coord

		for j := 0; j < GameBoardSize; j++ {

			var locationHit bool

			for _, coord := range coords {
				if coord.X == i && coord.Y == j {
					locationHit = true
				}
			}

			row = append(row, Coord{i, j, locationHit})
		}

		result.Coords = append(result.Coords, row)
	}

	return result
}
