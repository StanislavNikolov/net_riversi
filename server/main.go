package main

import (
	"fmt"
	"log"
	"riversi_server/riversi"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type GameEvent struct {
	Player       int           `json:"player"`
	Message      string        `json:"message"`
	HappenedAt   time.Time     `json:"happenedAt" db:"happened_at"`
	CurrentBoard riversi.Board `json:"currentBoard" db:"current_board"`
	Score        int           `json:"score"`
}

type GameSession struct {
	Id        int64     `json:"id" db:"id"`
	Player0   string    `json:"player0"`
	Player1   string    `json:"player1"`
	StartedAt time.Time `json:"startedAt" db:"started_at"`
	board     riversi.Board
}

func newGameSession(player0 string, player1 string) GameSession {
	var game GameSession
	game.board = riversi.NewBoard()
	game.Player0 = player0
	game.Player1 = player1
	game.StartedAt = time.Now()
	return game
}

func (game *GameSession) log(db *sqlx.DB, player int, message string) {
	event := GameEvent{
		Player:       player,
		Message:      message,
		HappenedAt:   time.Now(),
		CurrentBoard: game.board,
		Score:        game.board.GetScore(),
	}

	_, err := db.Exec(
		"INSERT INTO game_events (game_id, player, message, happened_at, current_board, score) VALUES (?, ?, ?, ?, ?, ?)",
		game.Id, event.Player, event.Message, event.HappenedAt, event.CurrentBoard, event.Score,
	)

	if err != nil {
		log.Println(err)
	}
}

func prepareDatabase() {
	db, err := sqlx.Open("sqlite", "./game_history.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	schema := `
		CREATE TABLE games (
			id INTEGER NOT NULL PRIMARY KEY,
			player0 TEXT,
			player1 TEXT,
			started_at DATETIME
		);
		CREATE TABLE game_events (
			id INTEGER NOT NULL PRIMARY KEY,
			game_id INTEGER NOT NULL,
			player TEXT,
			message TEXT,
			happened_at DATETIME,
			current_board TEXT,
			score INTEGER
		);
	`

	result, err := db.Exec(schema)
	fmt.Print(result, err)
}

func main() {
	prepareDatabase()

	go SetupHttpServer()
	SetupBotServer()
}
