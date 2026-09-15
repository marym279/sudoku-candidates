# sudoku-candidates

When I'm working through a sudoku by hand, or debugging output from a
puzzle generator, the question I actually want answered is small: for
this one empty cell, what digits are still legal right now? Not "solve
the whole board," just that.

This is a command-line tool (and the small Go library behind it) that
answers exactly that question, given a board's current state.

## Install

```
go build -o sudoku-candidates ./cmd/sudoku-candidates
```

## Board format

A board is 81 characters: digits `1`-`9` for filled cells, `.` or `0`
for empty ones. Whitespace and newlines are ignored, so you can write
it as nine lines of nine characters:

```
53..7....
6..195...
.98....6.
8...6...3
4..8.3..1
7...2...6
.6....28.
...419..5
....8..79
```

## Usage

Ask about one cell (1-9 row, 1-9 column):

```
$ sudoku-candidates -cell 1,3 board.txt
(1,3): 1 2 4
```

Or list every empty cell's candidates:

```
$ sudoku-candidates board.txt
(1,3): 1 2 4
(1,4): 2 4
(1,5): 1 2 8
...
```

Read from stdin with `-`:

```
$ cat board.txt | sudoku-candidates -cell 5,5 -
```

Add `-grid` to see every empty cell's candidates at once, laid out where
they sit: each cell is a 3x3 pad, and candidate *n* sits at position
`(n-1)/3, (n-1)%3` within its cell's pad. A filled cell shows its digit
alone in the middle of the pad:

```
$ sudoku-candidates -grid board.txt
      12 | 2    2 |1  12  2 
 5  3 4  |  6 7 4 6|4  4  4  
         |       8 | 89  9 8 
---------+---------+---------
...
```

`-grid` cannot be combined with `-cell` or `-json`; it's a plain-text
view of the whole board, not a machine-readable one.

Add `-json` to get machine-readable output instead, for piping into
`jq` or another script. A single `-cell` query prints one object; the
full-board form prints an array (an empty array `[]` for a fully
solved board, not `null`):

```
$ sudoku-candidates -cell 1,3 -json board.txt
{"row":1,"col":3,"candidates":[1,2,4]}

$ sudoku-candidates -json board.txt
[{"row":1,"col":3,"candidates":[1,2,4]},{"row":1,"col":4,"candidates":[2,4]},...]
```

Row and column in the JSON output are 1-9, matching `-cell` and the
plain-text output.

If the board already breaks a sudoku constraint (a digit repeated in a
row, column, or 3x3 box), the tool reports that instead of candidates -
candidates for a broken board aren't meaningful.

Add `-fill-singles` to auto-fill any cell whose candidates have already
narrowed to exactly one digit before reporting. It sweeps repeatedly,
since placing one single can narrow another cell down to a single too,
and stops once a pass places nothing. This is opt-in and off by
default: it can only ever remove cells from the report (a resolved
cell has nothing left to ask about), and it is not a solver - it won't
guess when a cell still has two or more candidates.

## As a library

```go
b, err := sudoku.ParseBoard(text)
if err != nil {
    // malformed input
}
if err := b.Valid(); err != nil {
    // board already breaks a constraint
}
candidates, err := b.Candidates(row, col) // zero-indexed
```

`Candidates` can return an empty (non-nil) slice: that means the cell
is empty but the board's current state has already ruled out every
digit, which happens on boards with no valid solution.

`FillSingles` is the opt-in auto-fill step behind `-fill-singles`: it
places any digit that's the only legal candidate left for its cell,
repeating until a pass places nothing, and returns the resulting board
along with how many cells it filled.

```go
solved, filled := b.FillSingles()
```

## Generating test boards

`GenerateSolved` and `GeneratePartial` build random valid boards, seeded
through a `*rand.Rand` so a failing case can be reproduced from its
seed. They're meant for tests that want many realistic boards rather
than a handful of hand-written fixtures - `GeneratePartial` in
particular is useful for throwing `AllCandidates` at boards with a
wide range of clue counts:

```go
rng := rand.New(rand.NewSource(seed))
b := sudoku.GeneratePartial(rng, 30) // 30 filled cells, rest empty
b.AllCandidates()
```

`GeneratePartial` does not check that the remaining clues still pin a
unique solution - it's a valid board, not a proper puzzle.
