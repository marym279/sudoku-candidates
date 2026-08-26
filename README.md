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

If the board already breaks a sudoku constraint (a digit repeated in a
row, column, or 3x3 box), the tool reports that instead of candidates -
candidates for a broken board aren't meaningful.

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
