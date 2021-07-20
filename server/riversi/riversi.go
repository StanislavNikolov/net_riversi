package riversi

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
	board.Squares[3][3] = 1
	board.Squares[3][4] = 0
	board.Squares[4][3] = 0
	board.Squares[4][4] = 1

	return board
}

func (board *Board) GetSquaresToBeFlipped(row int, col int, player int) [][2]int {
	directions := [8][2]int{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},
		{1, 1}, {-1, -1}, {1, -1}, {-1, 1},
	}

	var output [][2]int

	for _, direction := range directions {
		var maybeFlippable [][2]int

		deltaR := direction[0]
		deltaC := direction[1]
		currR := row + deltaR
		currC := col + deltaC
		for currR >= 0 && currC >= 0 && currR < 8 && currC < 8 {
			if board.Squares[currR][currC] == 255 {
				break
			}

			if board.Squares[currR][currC] == player {
				output = append(output, maybeFlippable...)
				break
			}

			maybeFlippable = append(maybeFlippable, [2]int{currR, currC})
			currR += deltaR
			currC += deltaC
		}
	}

	return output
}

func (board *Board) IsSquareAllowed(row int, col int, player int) bool {
	return board.Squares[row][col] == 255 && len(board.GetSquaresToBeFlipped(row, col, player)) > 0
}

func (board *Board) CheckPossibleMovesExist(player int) bool {
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			if board.IsSquareAllowed(row, col, player) {
				return true
			}
		}
	}
	return false
}

func (board *Board) GetScore() int {
	score := 0
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			if board.Squares[row][col] == 0 {
				score++
			}
			if board.Squares[row][col] == 1 {
				score--
			}
		}
	}
	return score
}
