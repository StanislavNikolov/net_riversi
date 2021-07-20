package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"riversi_server/riversi"
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
	Player       int           `json:"player"`
	Message      string        `json:"message"`
	Time         time.Time     `json:"time"`
	CurrentBoard riversi.Board `json:"currentBoard"`
}

type GameSession struct {
	Board   riversi.Board `json:"board"`
	Events  []GameEvent   `json:"events"`
	Players [2]string     `json:"players"`
	Started time.Time     `json:"started"`
}

func newGameSession(playerA string, playerB string) GameSession {
	var game GameSession
	game.Board = riversi.NewBoard()
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

	return square, endTime.Sub(beginTime), nil
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

func applyMove(square int, player int, board *riversi.Board) (bool, error) {
	row := square / 8
	col := square % 8

	// Technically this if is not needed, but provides more verbose output.
	if board.Squares[row][col] != 255 {
		return false, errors.New("non-empty square played")
	}

	// Check if player should be allowed to play there.
	if !board.IsSquareAllowed(row, col, player) {
		return false, errors.New("not allowed to place on that sqaure")
	}

	board.Squares[row][col] = player
	for _, coord := range board.GetSquaresToBeFlipped(row, col, player) {
		board.Squares[coord[0]][coord[1]] = player
	}

	return board.CheckPossibleMovesExist(nextPlayer(player)), nil
}

func fight(clientA *Client, clientB *Client) {
	game := newGameSession(clientA.token, clientB.token)

	game.log(255, "Starting game")

	gameSessionsMutex.Lock()
	gameSessions = append(gameSessions, &game)
	gameSessionsMutex.Unlock()

	players := [2]*Client{clientA, clientB}
	playerOnTurn := 0

	var loser int
	for {
		possible_to_play := game.Board.CheckPossibleMovesExist(playerOnTurn)
		if !possible_to_play {
			// No moves left. Find the winner
			score := game.Board.GetScore()
			if score > 0 {
				loser = 1
			}
			if score == 0 {
				loser = -1
			}
			if score < 0 {
				loser = 0
			}
			break
		}

		square, timeTaken, err := getMoveFromClient(players[playerOnTurn], playerOnTurn, &game)
		if err != nil {
			message := fmt.Sprintf("Player %d lost due to a communication error. Reason: %s", playerOnTurn, err)
			game.log(playerOnTurn, message)
			break
		}

		goToNextPlayer, err := applyMove(square, playerOnTurn, &game.Board)

		message := fmt.Sprintf("Player %d made move on square %d in %s", playerOnTurn, square, timeTaken)
		game.log(playerOnTurn, message)

		if err != nil {
			message := fmt.Sprintf("Player %d lost. Reason: %s", playerOnTurn, err)
			game.log(255, message)
			loser = playerOnTurn
			break
		}

		if goToNextPlayer {
			playerOnTurn = nextPlayer(playerOnTurn)
		} else {
			message := fmt.Sprintf("Skipping turn because player %d has no possible moves", nextPlayer(playerOnTurn))
			game.log(255, message)
		}
	}

	if loser == -1 {
		game.log(255, "Game over. Draw.")
		players[0].conn.Write([]byte("CTRL 😐 It was a draw!\n"))
		players[0].conn.Write([]byte("EXIT\n"))
		players[1].conn.Write([]byte("CTRL 😐 It was a draw!\n"))
		players[1].conn.Write([]byte("EXIT\n"))
		return
	}

	winner := nextPlayer(loser)
	game.log(255, fmt.Sprintf("Game over. Player %d won.", winner))

	players[loser].conn.Write([]byte("CTRL 😢 You lost!\n"))
	players[loser].conn.Write([]byte("EXIT\n"))
	players[winner].conn.Write([]byte("CTRL 😃 You won!\n"))
	players[winner].conn.Write([]byte("EXIT\n"))
}
