package sudoku

import (
	"math/rand"
	"testing"
)

func TestGenerateSolvedIsFullAndValid(t *testing.T) {
	for seed := int64(0); seed < 20; seed++ {
		rng := rand.New(rand.NewSource(seed))
		b := GenerateSolved(rng)
		if err := b.Valid(); err != nil {
			t.Fatalf("seed %d: GenerateSolved produced an invalid board: %v", seed, err)
		}
		for row := 0; row < 9; row++ {
			for col := 0; col < 9; col++ {
				if b[row][col] == 0 {
					t.Fatalf("seed %d: cell (%d,%d) is empty in a supposedly solved board", seed, row, col)
				}
			}
		}
	}
}

func TestGenerateSolvedIsDeterministic(t *testing.T) {
	a := GenerateSolved(rand.New(rand.NewSource(42)))
	b := GenerateSolved(rand.New(rand.NewSource(42)))
	if a != b {
		t.Fatalf("same seed produced different boards:\n%v\n%v", a, b)
	}
}

func TestGeneratePartialLeavesExactClueCount(t *testing.T) {
	cases := []int{0, 1, 17, 30, 81}
	for _, clues := range cases {
		rng := rand.New(rand.NewSource(int64(clues)))
		b := GeneratePartial(rng, clues)
		got := 0
		for row := 0; row < 9; row++ {
			for col := 0; col < 9; col++ {
				if b[row][col] != 0 {
					got++
				}
			}
		}
		if got != clues {
			t.Fatalf("clues=%d: board has %d filled cells, want %d", clues, got, clues)
		}
		if err := b.Valid(); err != nil {
			t.Fatalf("clues=%d: board is invalid: %v", clues, err)
		}
	}
}

func TestGeneratePartialClampsOutOfRangeClues(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	if b := GeneratePartial(rng, -5); len(b.AllCandidates()) != 81 {
		t.Fatalf("clues=-5: want a fully empty board, got %d filled cells", 81-len(b.AllCandidates()))
	}
	rng = rand.New(rand.NewSource(1))
	if b := GeneratePartial(rng, 200); len(b.AllCandidates()) != 0 {
		t.Fatalf("clues=200: want a fully filled board, got %d empty cells", len(b.AllCandidates()))
	}
}

// AllCandidates on GeneratePartial output should never itself panic or
// error regardless of how few clues remain; this is the actual use case
// the generator exists for, so exercise it directly at fuzz-test scale.
func TestGeneratePartialFeedsAllCandidatesAtScale(t *testing.T) {
	for seed := int64(0); seed < 200; seed++ {
		rng := rand.New(rand.NewSource(seed))
		clues := rng.Intn(82)
		b := GeneratePartial(rng, clues)
		b.AllCandidates()
	}
}
