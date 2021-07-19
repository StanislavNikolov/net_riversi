package riversi

import (
	"testing"
)

func TestGetSquaresToBeFlipped(t *testing.T) {
	board := NewBoard()
	/* . . . .
	 * . 1 0 .
	 * . 0 1 .
	 * . . . .
	 */

	got := board.GetSquaresToBeFlipped(0, 0, 0)
	if len(got) != 0 {
		t.Error("There should be no squares found", got)
	}

	got = board.GetSquaresToBeFlipped(4, 5, 0)
	if len(got) != 1 || got[0] != [2]int{4, 4} {
		t.Error("There should be 1 square found", got)
	}

	board.Squares[4][4] = 0
	board.Squares[4][5] = 0
	/* . . . .
	 * . 1 0 .
	 * . 0 0 0
	 * . . . .
	 */

	got = board.GetSquaresToBeFlipped(2, 2, 1)
	if len(got) != 0 {
		t.Error("There should be no squares found", got)
	}

	got = board.GetSquaresToBeFlipped(3, 5, 1)
	if len(got) != 1 || got[0] != [2]int{3, 4} {
		t.Error("There should be 1 square found", got)
	}

	board.Squares[3][4] = 1
	board.Squares[3][5] = 1
	/* . . . .
	 * . 1 1 1
	 * . 0 0 0
	 * . . . .
	 */
	got = board.GetSquaresToBeFlipped(2, 5, 0)
	if len(got) != 2 || got[0] != [2]int{3, 5} || got[1] != [2]int{3, 4} {
		t.Error("There should be 1 square found", got)
	}
}
