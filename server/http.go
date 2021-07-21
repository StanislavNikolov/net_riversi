package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

func SetupHttpServer() {
	db, err := sqlx.Open("sqlite", "./game_history.db")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	app := fiber.New()

	app.Static("/", "./frontend")

	app.Get("/api/games", func(c *fiber.Ctx) error {
		var index []GameSession
		err := db.Select(&index, "SELECT id, started_at, player0, player1 FROM games ORDER BY id LIMIT 1000")

		if err != nil {
			log.Println(err)
			return c.JSON(nil)
		}

		return c.JSON(index)
	})

	app.Get("/api/game/:gameId", func(c *fiber.Ctx) error {
		var game GameSession
		err := db.Get(&game, "SELECT id, started_at, player0, player1 FROM games WHERE id = ?", c.Params("gameId"))

		if err != nil {
			log.Println(err)
			return c.JSON(nil)
		}

		return c.JSON(game)
	})

	app.Get("/api/events/:gameId", func(c *fiber.Ctx) error {
		var events []GameEvent
		err := db.Select(&events, "SELECT player, message, happened_at, current_board, score FROM game_events WHERE game_id = ? ORDER BY id", c.Params("gameId"))

		if err != nil {
			log.Println(err)
			return c.JSON(nil)
		}

		return c.JSON(events)
	})

	app.Listen(":3000")
}
