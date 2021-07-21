package main

import (
	"database/sql"
	"log"

	"github.com/gofiber/fiber/v2"
)

type indexEntry struct {
	GameId int    `json:"id"`
	Data   string `json:"data"`
}

func SetupHttpServer() {
	db, err := sql.Open("sqlite", "./game_history.db")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	app := fiber.New()

	app.Static("/", "./frontend")

	app.Get("/api/games", func(c *fiber.Ctx) error {
		var index []indexEntry

		rows, err := db.Query("SELECT id, json FROM games ORDER BY id DESC LIMIT 1000")
		if err != nil {
			log.Println(err)
			return err
		}

		defer rows.Close()
		for rows.Next() {
			var ie indexEntry
			err := rows.Scan(&ie.GameId, &ie.Data)
			if err != nil {
				log.Println(err)
				return err
			}
			index = append(index, ie)
		}

		return c.JSON(index)
	})

	app.Get("/api/game/:gameId", func(c *fiber.Ctx) error {
		row := db.QueryRow("SELECT json FROM games WHERE id = ?", c.Params("gameId"))

		var json string
		switch err := row.Scan(&json); err {
		case nil:
			return c.JSON(json)
		default:
			return err
		}
	})

	app.Get("/api/events/:gameId", func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT json FROM game_events WHERE game_id = ? ORDER BY id", c.Params("gameId"))
		if err != nil {
			log.Println(err)
			return err
		}
		defer rows.Close()

		var data []string
		for rows.Next() {
			var json string
			err := rows.Scan(&json)
			if err != nil {
				log.Println(err)
				return err
			}
			data = append(data, json)
		}

		return c.JSON(data)
	})

	app.Listen(":3000")
}
