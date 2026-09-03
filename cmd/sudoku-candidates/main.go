// Command sudoku-candidates reads a sudoku board and reports which
// digits are still legal for an empty cell, or for every empty cell.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"sudoku-candidates/sudoku"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "sudoku-candidates:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("sudoku-candidates", flag.ContinueOnError)
	cell := fs.String("cell", "", `query a single cell as "row,col" using 1-9 coordinates (default: every empty cell)`)
	jsonOut := fs.Bool("json", false, "output as JSON instead of plain text, for scripting")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "usage: sudoku-candidates [-cell row,col] [-json] <board-file|->")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return fmt.Errorf("expected exactly one board argument")
	}

	data, err := readBoard(fs.Arg(0))
	if err != nil {
		return err
	}
	board, err := sudoku.ParseBoard(data)
	if err != nil {
		return err
	}
	if err := board.Valid(); err != nil {
		return err
	}

	if *cell != "" {
		row, col, err := parseCell(*cell)
		if err != nil {
			return err
		}
		candidates, err := board.Candidates(row, col)
		if err != nil {
			return err
		}
		if *jsonOut {
			return json.NewEncoder(out).Encode(toJSONCell(row, col, candidates))
		}
		fmt.Fprintln(out, formatCandidates(row, col, candidates))
		return nil
	}

	all := board.AllCandidates()
	if *jsonOut {
		cells := make([]jsonCell, len(all))
		for i, cc := range all {
			cells[i] = toJSONCell(cc.Row, cc.Col, cc.Candidates)
		}
		return json.NewEncoder(out).Encode(cells)
	}
	for _, cc := range all {
		fmt.Fprintln(out, formatCandidates(cc.Row, cc.Col, cc.Candidates))
	}
	return nil
}

// jsonCell is the -json representation of a cell's candidates. Row and
// Col are 1-9, matching the -cell flag and the plain-text output, so
// output from one mode can be checked against the other by hand.
type jsonCell struct {
	Row        int   `json:"row"`
	Col        int   `json:"col"`
	Candidates []int `json:"candidates"`
}

func toJSONCell(row, col int, candidates []int) jsonCell {
	return jsonCell{Row: row + 1, Col: col + 1, Candidates: candidates}
}

func readBoard(path string) (string, error) {
	if path == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	return string(data), nil
}

func parseCell(s string) (row, col int, err error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("cell %q must be \"row,col\"", s)
	}
	r, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || r < 1 || r > 9 {
		return 0, 0, fmt.Errorf("cell %q: row must be 1-9", s)
	}
	c, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || c < 1 || c > 9 {
		return 0, 0, fmt.Errorf("cell %q: col must be 1-9", s)
	}
	return r - 1, c - 1, nil
}

func formatCandidates(row, col int, candidates []int) string {
	if len(candidates) == 0 {
		return fmt.Sprintf("(%d,%d): none", row+1, col+1)
	}
	strs := make([]string, len(candidates))
	for i, v := range candidates {
		strs[i] = strconv.Itoa(v)
	}
	return fmt.Sprintf("(%d,%d): %s", row+1, col+1, strings.Join(strs, " "))
}
