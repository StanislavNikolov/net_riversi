package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var users = map[string]string{}
var lobby = make(chan *Client)

var gameSessionsMutex sync.Mutex
var gameSessions []*GameSession

type Client struct {
	token string
	conn  net.Conn
}

type GameEvent struct {
	Player       int       `json:"player"`
	Message      string    `json:"message"`
	Time         time.Time `json:"time"`
	CurrentBoard Board     `json:"currentBoard"`
}

type GameSession struct {
	Board   Board       `json:"board"`
	Events  []GameEvent `json:"events"`
	Players [2]string   `json:"players"`
	Started time.Time   `json:"started"`
}

func newGameSession(playerA string, playerB string) GameSession {
	var game GameSession
	game.Board = NewBoard()
	game.Players = [2]string{playerA, playerB}
	game.Started = time.Now()
	return game
}

func (game *GameSession) log(player int, message string) {
	event := GameEvent{
		Player:       player,
		Message:      message,
		Time:         time.Now(),
		CurrentBoard: game.Board,
	}
	game.Events = append(game.Events, event)
}

func main() {
	listenSocket, err := net.Listen("tcp", ":8081")

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer listenSocket.Close()
	fmt.Println("Waiting for bots on localhost:8081")

	go SetupHttpServer()

	go makeMatches()

	for {
		clientConn, err := listenSocket.Accept()
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return
		}

		fmt.Println("Client connected from:", clientConn.RemoteAddr())
		go authenticate(clientConn)
	}
}

func authenticate(conn net.Conn) {
	buffer, err := bufio.NewReader(conn).ReadBytes('\n')

	if err != nil {
		log.Println("Client left.")
		conn.Close()
		return
	}

	loginPacket := string(buffer[:len(buffer)-1])
	// [randomToken, username]
	tokens := strings.Split(loginPacket, " ")

	if len(tokens) != 2 {
		log.Println("Client failed to authenticate", conn.RemoteAddr())
		conn.Close()
		return
	}

	users[tokens[0]] = tokens[1]

	log.Println("Client logged in:", tokens)
	lobby <- &Client{tokens[1], conn}
}

func makeMatches() {
	for {
		clientA := <-lobby
		clientB := <-lobby

		log.Printf("Starting game between %s and %s\n", clientA.token, clientB.token)
		go func() {
			fight(clientA, clientB)
			clientA.conn.Close()
			clientB.conn.Close()
		}()
	}
}

func getMoveFromClient(client *Client, player int, game *GameSession) (int, time.Duration, error) {
	ser := game.serializeFromPerspective(player)

	err := client.conn.SetDeadline(time.Now().Add(game.geTimeLeftForPlayer(player)))

	if err != nil {
		return 0, 0, err
	}

	beginTime := time.Now()

	_, err = client.conn.Write([]byte(ser + "\n"))

	if err != nil {
		return 0, 0, err
	}

	move, err := bufio.NewReader(client.conn).ReadString('\n')
	if err != nil {
		return 0, 0, err
	}

	square, err := strconv.Atoi(strings.TrimSpace(move))
	if err != nil {
		return 0, 0, errors.New("failed to parse move as number")
	}

	if square < 0 || square >= 64 {
		return 0, 0, errors.New("invalid square played")
	}

	endTime := time.Now()

	return square, endTime.Sub(beginTime), nil // remove newline at the end of the answer
}

func (game GameSession) serializeFromPerspective(player int) string {
	var serialized string

	serialized += strconv.Itoa(int(game.geTimeLeftForPlayer(player) / time.Millisecond))
	serialized += " "

	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			switch game.Board.Squares[x][y] {
			case 255:
				serialized += "E"
			case player:
				serialized += "W"
			default:
				serialized += "B"
			}
		}
	}

	return serialized
}

func (game GameSession) geTimeLeftForPlayer(player int) time.Duration {
	/*
		totalTimeTaken := time.Duration(0)
		// Initially every player has some time in the bank.
		bonusTime := 5000 * time.Millisecond

		for _, move := range game.Moves {
			if move.Player != player {
				// Ignore the other player's moves.
				continue
			}

			totalTimeTaken += move.TimeTaken
			// Every move made gives you a bonus.
			bonusTime += 500 * time.Millisecond
		}

		return bonusTime - totalTimeTaken
	*/
	return 5000 * time.Millisecond
}

func nextPlayer(currentPlayer int) int {
	return (currentPlayer + 1) % 2
}

func applyMove(square int, player int, board *Board) (bool, error) {
	row := square / 8
	col := square % 8

	if board.Squares[row][col] != 255 {
		return false, errors.New("non-empty square played")
	}

	// check if player should be allowed to play there
	if !AllowedSquare(board, row, col, player) {
		return false, errors.New("not allowed to place on that sqaure")
	}
	board.Squares[row][col] = player

	// TODO flip pieces

	// check if the other player has any moves possible
	return AllowedSquare(board, row, col, nextPlayer(player)), nil
}

func fight(clientA *Client, clientB *Client) {
	game := newGameSession(clientA.token, clientB.token)

	gameSessionsMutex.Lock()
	gameSessions = append(gameSessions, &game)
	gameSessionsMutex.Unlock()

	players := [2]*Client{clientA, clientB}
	playerOnTurn := 0

	var loser int
	for {
		square, timeTaken, err := getMoveFromClient(players[playerOnTurn], playerOnTurn, &game)
		if err != nil {
			message := fmt.Sprintf("Player %d lost due to a network error. Reason: %s", playerOnTurn, err)
			game.log(playerOnTurn, message)
			break
		}

		message := fmt.Sprintf("Player %d made move on square %d in %s", playerOnTurn, square, timeTaken)
		game.log(playerOnTurn, message)

		goToNextPlayer, err := applyMove(square, playerOnTurn, &game.Board)
		if err != nil {
			message := fmt.Sprintf("Player %d lost. Reason: %s", playerOnTurn, err)
			game.log(playerOnTurn, message)
			loser = playerOnTurn
			break
		}

		if goToNextPlayer {
			playerOnTurn = nextPlayer(playerOnTurn)
		}
	}

	winner := nextPlayer(loser)

	players[loser].conn.Write([]byte("CTRL 😢 You lost!"))
	players[winner].conn.Write([]byte("CTRL 😃 You won!"))
}