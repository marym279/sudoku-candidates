package sudoku

import "math/rand"

// GenerateSolved returns a random, fully filled, valid board using rng
// to choose between otherwise-equal choices. Two calls with rng values
// seeded the same way produce the same board.
//
// This exists for tests that want realistic boards at scale rather than
// the handful of hand-written fixtures in board_test.go: seed rng once
// per test case and the failure is reproducible.
func GenerateSolved(rng *rand.Rand) Board {
	var b Board
	fillFrom(&b, 0, rng)
	return b
}

// GeneratePartial returns a board derived from a random solved board by
// clearing cells until only clues remain filled. clues is clamped to
// [0, 81]. The result always satisfies Valid, since clearing cells from
// a valid board cannot introduce a duplicate, but it is not a proper
// puzzle: it doesn't check the remaining clues still pin a unique
// solution, only that they're consistent with the one used to build it.
func GeneratePartial(rng *rand.Rand, clues int) Board {
	if clues < 0 {
		clues = 0
	}
	if clues > 81 {
		clues = 81
	}
	b := GenerateSolved(rng)
	for _, idx := range rng.Perm(81)[:81-clues] {
		b[idx/9][idx%9] = 0
	}
	return b
}

// fillFrom fills cells [idx, 81) of b via randomized backtracking,
// leaving cells before idx untouched. It reports whether a full
// assignment was found; for idx == 0 on an empty board this is always
// true, since every partially filled sudoku prefix here starts from
// nothing and standard backtracking always completes an empty grid.
func fillFrom(b *Board, idx int, rng *rand.Rand) bool {
	if idx == 81 {
		return true
	}
	row, col := idx/9, idx%9
	// Candidates ignores its error here: row/col are always in range
	// and the cell is always empty at this point in the recursion.
	candidates, _ := b.Candidates(row, col)
	rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	for _, v := range candidates {
		b[row][col] = v
		if fillFrom(b, idx+1, rng) {
			return true
		}
	}
	b[row][col] = 0
	return false
}
