// Package sudoku parses 9x9 sudoku boards and answers one question:
// given the board as it stands, which digits could legally go in a
// particular empty cell. It does not solve boards or judge difficulty.
package sudoku

import "fmt"

// Board is a 9x9 grid of digits. 0 means the cell is empty.
type Board [9][9]int

// ParseBoard reads an 81-cell board from s. Whitespace (spaces, tabs,
// newlines, carriage returns) is ignored, so a board can be written as
// nine lines of nine characters or as one long line. Both '0' and '.'
// mark an empty cell.
func ParseBoard(s string) (Board, error) {
	var b Board
	idx := 0
	for _, r := range s {
		switch r {
		case '\n', '\r', ' ', '\t':
			continue
		}
		if idx >= 81 {
			return Board{}, fmt.Errorf("sudoku: board has more than 81 cells")
		}
		var v int
		switch {
		case r == '.':
			v = 0
		case r >= '0' && r <= '9':
			v = int(r - '0')
		default:
			return Board{}, fmt.Errorf("sudoku: invalid character %q at cell %d", r, idx)
		}
		b[idx/9][idx%9] = v
		idx++
	}
	if idx != 81 {
		return Board{}, fmt.Errorf("sudoku: board has %d cells, want 81", idx)
	}
	return b, nil
}

// Valid reports whether the filled cells in b already break a sudoku
// constraint: a repeated digit in some row, column, or 3x3 box. It does
// not check that the board is solvable, only that it isn't already
// broken.
func (b Board) Valid() error {
	for row := 0; row < 9; row++ {
		if v, dup := firstDuplicate(rowValues(b, row)); dup {
			return fmt.Errorf("sudoku: row %d has duplicate value %d", row+1, v)
		}
	}
	for col := 0; col < 9; col++ {
		if v, dup := firstDuplicate(colValues(b, col)); dup {
			return fmt.Errorf("sudoku: column %d has duplicate value %d", col+1, v)
		}
	}
	for boxRow := 0; boxRow < 3; boxRow++ {
		for boxCol := 0; boxCol < 3; boxCol++ {
			if v, dup := firstDuplicate(boxValues(b, boxRow*3, boxCol*3)); dup {
				return fmt.Errorf("sudoku: box at row %d, col %d has duplicate value %d", boxRow*3+1, boxCol*3+1, v)
			}
		}
	}
	return nil
}

func firstDuplicate(values []int) (int, bool) {
	seen := [10]bool{}
	for _, v := range values {
		if v == 0 {
			continue
		}
		if seen[v] {
			return v, true
		}
		seen[v] = true
	}
	return 0, false
}

func rowValues(b Board, row int) []int {
	return b[row][:]
}

func colValues(b Board, col int) []int {
	values := make([]int, 9)
	for row := 0; row < 9; row++ {
		values[row] = b[row][col]
	}
	return values
}

func boxValues(b Board, startRow, startCol int) []int {
	values := make([]int, 0, 9)
	for row := startRow; row < startRow+3; row++ {
		for col := startCol; col < startCol+3; col++ {
			values = append(values, b[row][col])
		}
	}
	return values
}

// Candidates returns the digits 1-9 that could legally be placed in the
// cell at (row, col) given the board's current filled cells. row and
// col are zero-indexed. It returns an error if the coordinates are out
// of range or the cell is already filled.
func (b Board) Candidates(row, col int) ([]int, error) {
	if row < 0 || row > 8 || col < 0 || col > 8 {
		return nil, fmt.Errorf("sudoku: cell (%d, %d) is out of range", row, col)
	}
	if b[row][col] != 0 {
		return nil, fmt.Errorf("sudoku: cell (%d, %d) is already filled with %d", row, col, b[row][col])
	}
	used := [10]bool{}
	for _, v := range rowValues(b, row) {
		used[v] = true
	}
	for _, v := range colValues(b, col) {
		used[v] = true
	}
	for _, v := range boxValues(b, (row/3)*3, (col/3)*3) {
		used[v] = true
	}
	candidates := make([]int, 0, 9)
	for v := 1; v <= 9; v++ {
		if !used[v] {
			candidates = append(candidates, v)
		}
	}
	return candidates, nil
}

// CellCandidates pairs a cell's coordinates with its legal candidates.
type CellCandidates struct {
	Row, Col   int
	Candidates []int
}

// AllCandidates returns the candidates for every empty cell on the
// board, in row-major order. A solved board returns an empty slice.
func (b Board) AllCandidates() []CellCandidates {
	var out []CellCandidates
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			if b[row][col] != 0 {
				continue
			}
			// Candidates can't fail here: row/col are in range and the
			// cell is confirmed empty.
			c, _ := b.Candidates(row, col)
			out = append(out, CellCandidates{Row: row, Col: col, Candidates: c})
		}
	}
	return out
}

// FillSingles repeatedly places the digit into any empty cell that has
// exactly one legal candidate. Placing that digit can only remove
// options from the rest of the board, never add one, so it can turn
// other cells into singles too; FillSingles keeps sweeping the board
// until a pass places nothing. It returns the resulting board and how
// many cells it filled.
//
// This is not a solver: a board can reach a fixed point with unfilled
// cells remaining, each with two or more candidates, and FillSingles
// will stop there rather than guess.
func (b Board) FillSingles() (Board, int) {
	filled := 0
	for {
		progressed := false
		for row := 0; row < 9; row++ {
			for col := 0; col < 9; col++ {
				if b[row][col] != 0 {
					continue
				}
				// Candidates can't fail here: row/col are in range and
				// the cell is confirmed empty.
				c, _ := b.Candidates(row, col)
				if len(c) == 1 {
					b[row][col] = c[0]
					filled++
					progressed = true
				}
			}
		}
		if !progressed {
			break
		}
	}
	return b, filled
}
