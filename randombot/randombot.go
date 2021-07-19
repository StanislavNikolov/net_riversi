package main

import (
	"bufio"
	"fmt"
	"os"
	"riversi_server/riversi"
	"strings"
)

func readBoard() riversi.Board {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	strboard := strings.Split(line, " ")[1]

	var board riversi.Board
	for i, char := range strboard {
		row := i / 8
		col := i % 8

		if char == 'E' {
			board.Squares[row][col] = 255
		}
		if char == 'W' {
			board.Squares[row][col] = 0
		}
		if char == 'B' {
			board.Squares[row][col] = 1
		}
	}

	//fmt.Fprintf(os.Stderr, "your message here")

	return board
}

func play(board riversi.Board) int {
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			if !board.IsSquareAllowed(row, col, 0) {
				continue
			}
			fl := board.GetSquaresToBeFlipped(row, col, 0)
			if len(fl) > 0 {
				return row*8 + col
			}
		}
	}
	return 0
}

func main() {
	for {
		board := readBoard()
		move := play(board)
		fmt.Println(move)
	}
}
