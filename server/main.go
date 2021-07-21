package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func prepareDatabase() {
	db, err := sql.Open("sqlite", "./game_history.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	result, err := db.Exec("CREATE TABLE games (id INTEGER NOT NULL PRIMARY KEY, json STRING)")
	fmt.Print(result, err)

	result, err = db.Exec("CREATE TABLE game_events (id INTEGER NOT NULL PRIMARY KEY, game_id INTEGER NOT NULL, json STRING)")
	fmt.Print(result, err)
}

func main() {
	prepareDatabase()

	go SetupHttpServer()
	SetupBotServer()
}
