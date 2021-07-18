package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type indexEntry struct {
	GameId  int       `json:"id"`
	Players [2]string `json:"players"`
	Started time.Time `json:"started"`
}

func SetupHttpServer() {
	app := fiber.New()

	app.Static("/", "./frontend")

	app.Get("/api/games", func(c *fiber.Ctx) error {
		// return c.SendString(strconv.Itoa(len(gameSessions)))
		var index []indexEntry

		for i, game := range gameSessions {
			index = append(index, indexEntry{i, game.Players, game.Started})
		}

		return c.JSON(index)
	})

	app.Get("/api/game/:gameId", func(c *fiber.Ctx) error {
		// return c.SendString(strconv.Itoa(len(gameSessions)))
		idx, err := strconv.Atoi(c.Params("gameId"))
		if err != nil {
			return err
		}

		if idx < 0 || idx >= len(gameSessions) {
			return errors.New(fmt.Sprint("no game with such id ", idx))
		}

		return c.JSON(gameSessions[idx])
	})

	app.Listen(":3000")
}
