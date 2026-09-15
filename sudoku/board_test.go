package sudoku

import (
	"reflect"
	"strings"
	"testing"
)

const emptyBoard = `
.........
.........
.........
.........
.........
.........
.........
.........
.........
`

// A known valid, fully solved board (no empty cells).
const solvedBoard = `
534678912
672195348
198342567
859761423
426853791
713924856
961537284
287419635
345286179
`

// Row 0 supplies 2-9 as candidates already used; a 1 sits in column 0 at
// row 1, inside the same box. Together row+col+box block every digit,
// so cell (0,0) has zero candidates even though the board is valid.
const noCandidatesBoard = `
.23456789
1........
.........
.........
.........
.........
.........
.........
.........
`

// A 9 sits in the top-left box (0,0) and a 4 sits in the top-middle box
// (0,3). Cell (2,5) shares a box with the 4 but not with the 9, so only
// the 4 should be excluded from its candidates.
const boxBoundaryBoard = `
9..4.....
.........
.........
.........
.........
.........
.........
.........
.........
`

const rowDuplicateBoard = `
5.....5..
.........
.........
.........
.........
.........
.........
.........
.........
`

const colDuplicateBoard = `
5........
.........
.........
5........
.........
.........
.........
.........
.........
`

// 5 appears at (0,0) and (1,1): same box, different row and column.
const boxDuplicateBoard = `
5........
.5.......
.........
.........
.........
.........
.........
.........
.........
`

func TestParseBoard(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantErr string // substring, empty means no error
	}{
		{name: "empty board with dots", in: emptyBoard},
		{name: "zeros instead of dots", in: strings.ReplaceAll(emptyBoard, ".", "0")},
		{name: "mixed dots and zeros", in: "." + strings.Repeat("0", 40) + strings.Repeat(".", 40)},
		{name: "solved board", in: solvedBoard},
		{name: "too short", in: strings.Repeat(".", 80), wantErr: "want 81"},
		{name: "too long", in: strings.Repeat(".", 82), wantErr: "more than 81"},
		{name: "invalid character", in: strings.Repeat(".", 40) + "x" + strings.Repeat(".", 40), wantErr: "invalid character"},
		{name: "empty string", in: "", wantErr: "want 81"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseBoard(c.in)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("ParseBoard(%q): unexpected error: %v", c.name, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ParseBoard(%q): expected error containing %q, got nil", c.name, c.wantErr)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("ParseBoard(%q): error %q does not contain %q", c.name, err, c.wantErr)
			}
		})
	}
}

func TestValid(t *testing.T) {
	cases := []struct {
		name    string
		board   string
		wantErr string
	}{
		{name: "empty board", board: emptyBoard},
		{name: "solved board", board: solvedBoard},
		{name: "row duplicate", board: rowDuplicateBoard, wantErr: "row"},
		{name: "column duplicate", board: colDuplicateBoard, wantErr: "column"},
		{name: "box duplicate, distinct row and column", board: boxDuplicateBoard, wantErr: "box"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b, err := ParseBoard(c.board)
			if err != nil {
				t.Fatalf("ParseBoard: %v", err)
			}
			err = b.Valid()
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("Valid(): unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Valid(): expected error containing %q, got nil", c.wantErr)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("Valid(): error %q does not contain %q", err, c.wantErr)
			}
		})
	}
}

func TestCandidates(t *testing.T) {
	cases := []struct {
		name    string
		board   string
		row     int
		col     int
		want    []int
		wantErr string
	}{
		{
			name:  "empty board, corner cell",
			board: emptyBoard,
			row:   0, col: 0,
			want: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name:    "already filled cell",
			board:   solvedBoard,
			row:     0, col: 0,
			wantErr: "already filled",
		},
		{
			name:    "row out of range",
			board:   emptyBoard,
			row:     9, col: 0,
			wantErr: "out of range",
		},
		{
			name:    "negative column",
			board:   emptyBoard,
			row:     0, col: -1,
			wantErr: "out of range",
		},
		{
			name:  "contradiction leaves zero candidates",
			board: noCandidatesBoard,
			row:   0, col: 0,
			want: []int{},
		},
		{
			name:  "box boundary excludes same-box value only",
			board: boxBoundaryBoard,
			row:   2, col: 5,
			want: []int{1, 2, 3, 5, 6, 7, 8, 9},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b, err := ParseBoard(c.board)
			if err != nil {
				t.Fatalf("ParseBoard: %v", err)
			}
			got, err := b.Candidates(c.row, c.col)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("Candidates(%d, %d): unexpected error: %v", c.row, c.col, err)
				}
				if !reflect.DeepEqual(got, c.want) {
					t.Fatalf("Candidates(%d, %d) = %v, want %v", c.row, c.col, got, c.want)
				}
				return
			}
			if err == nil {
				t.Fatalf("Candidates(%d, %d): expected error containing %q, got nil", c.row, c.col, c.wantErr)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("Candidates(%d, %d): error %q does not contain %q", c.row, c.col, err, c.wantErr)
			}
		})
	}
}

// Cell (0,8) is a naked single: row 0 supplies every digit but 9.
// Filling it completes column 8's missing-1-and-9 gap down to just 1,
// which is what (5,8) needs to become a single itself only after the
// first fill happens - a genuine two-step cascade, not two independent
// singles.
const cascadingSinglesBoard = `
12345678.
........2
........3
........4
........5
.........
........6
........7
........8
`

func TestAllCandidates(t *testing.T) {
	t.Run("solved board has nothing to report", func(t *testing.T) {
		b, err := ParseBoard(solvedBoard)
		if err != nil {
			t.Fatalf("ParseBoard: %v", err)
		}
		if got := b.AllCandidates(); len(got) != 0 {
			t.Fatalf("AllCandidates() on solved board = %v, want empty", got)
		}
	})

	t.Run("empty board reports all 81 cells", func(t *testing.T) {
		b, err := ParseBoard(emptyBoard)
		if err != nil {
			t.Fatalf("ParseBoard: %v", err)
		}
		got := b.AllCandidates()
		if len(got) != 81 {
			t.Fatalf("AllCandidates() returned %d cells, want 81", len(got))
		}
		first := got[0]
		if first.Row != 0 || first.Col != 0 {
			t.Fatalf("first cell = (%d, %d), want (0, 0)", first.Row, first.Col)
		}
		if len(first.Candidates) != 9 {
			t.Fatalf("first cell candidates = %v, want all 9 digits", first.Candidates)
		}
	})
}

func TestFillSingles(t *testing.T) {
	t.Run("solved board fills nothing", func(t *testing.T) {
		b, err := ParseBoard(solvedBoard)
		if err != nil {
			t.Fatalf("ParseBoard: %v", err)
		}
		got, filled := b.FillSingles()
		if filled != 0 {
			t.Fatalf("filled = %d, want 0", filled)
		}
		if got != b {
			t.Fatalf("FillSingles changed a solved board")
		}
	})

	t.Run("wide open board fills nothing", func(t *testing.T) {
		b, err := ParseBoard(emptyBoard)
		if err != nil {
			t.Fatalf("ParseBoard: %v", err)
		}
		_, filled := b.FillSingles()
		if filled != 0 {
			t.Fatalf("filled = %d, want 0: every cell has 9 candidates, none is a single", filled)
		}
	})

	t.Run("cascading singles resolve in one call", func(t *testing.T) {
		b, err := ParseBoard(cascadingSinglesBoard)
		if err != nil {
			t.Fatalf("ParseBoard: %v", err)
		}
		got, filled := b.FillSingles()
		if filled != 2 {
			t.Fatalf("filled = %d, want 2", filled)
		}
		if got[0][8] != 9 {
			t.Fatalf("(0,8) = %d, want 9", got[0][8])
		}
		if got[5][8] != 1 {
			t.Fatalf("(5,8) = %d, want 1: only reachable after (0,8) is filled", got[5][8])
		}
		if err := got.Valid(); err != nil {
			t.Fatalf("FillSingles produced an invalid board: %v", err)
		}
		// Plenty of cells in columns 0-7 are still wide open; FillSingles
		// must not have guessed at any of them.
		if len(got.AllCandidates()) == 0 {
			t.Fatalf("board looks fully solved, want cells left open")
		}
	})
}
