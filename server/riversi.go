package main

type Board struct {
	// Allowed values: 0, 1, 255 TODO: make it an enum
	Squares [8][8]int `json:"squares"`
}

func NewBoard() Board {
	var board Board

	// fill board with empty squares
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			board.Squares[row][col] = 255
		}
	}
	board.Squares[3][3] = 0
	board.Squares[3][4] = 1
	board.Squares[4][3] = 1
	board.Squares[4][4] = 0

	return board
}

func AllowedSquare(board *Board, row int, col int, player int) bool {
	directions := [8][2]int{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},
		{1, 1}, {-1, 1}, {1, -1}, {-1, 1},
	}

	for _, direction := range directions {
		deltaR := direction[0]
		deltaC := direction[1]
		currR := row + deltaR
		currC := col + deltaC
		for currR >= 0 && currC >= 0 && currR < 8 && currC < 8 {
			if board.Squares[currR][currC] == 255 {
				break
			}
			if board.Squares[currR][currC] == player {
				return true
			}
			currR += deltaR
			currC += deltaC
		}
	}

	return false
}
