package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testBoard = `
53..7....
6..195...
.98....6.
8...6...3
4..8.3..1
7...2...6
.6....28.
...419..5
....8..79
`

func writeTestBoard(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "board.txt")
	if err := os.WriteFile(path, []byte(testBoard), 0o644); err != nil {
		t.Fatalf("writing test board: %v", err)
	}
	return path
}

func TestRunTextSingleCell(t *testing.T) {
	var out bytes.Buffer
	path := writeTestBoard(t)
	if err := run([]string{"-cell", "1,3", path}, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	want := "(1,3): 1 2 4\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestRunJSONSingleCell(t *testing.T) {
	var out bytes.Buffer
	path := writeTestBoard(t)
	if err := run([]string{"-cell", "1,3", "-json", path}, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	var got jsonCell
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("unmarshaling output %q: %v", out.String(), err)
	}
	want := jsonCell{Row: 1, Col: 3, Candidates: []int{1, 2, 4}}
	if got.Row != want.Row || got.Col != want.Col || !equalInts(got.Candidates, want.Candidates) {
		t.Fatalf("decoded = %+v, want %+v", got, want)
	}
}

func TestRunJSONAllCells(t *testing.T) {
	var out bytes.Buffer
	path := writeTestBoard(t)
	if err := run([]string{"-json", path}, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	var got []jsonCell
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("unmarshaling output %q: %v", out.String(), err)
	}
	if len(got) == 0 {
		t.Fatalf("decoded 0 cells, want at least one empty cell")
	}
	first := got[0]
	if first.Row != 1 || first.Col != 3 {
		t.Fatalf("first cell = (%d,%d), want (1,3)", first.Row, first.Col)
	}
}

func TestRunJSONSolvedBoardIsEmptyArrayNotNull(t *testing.T) {
	solved := `
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
	path := filepath.Join(t.TempDir(), "solved.txt")
	if err := os.WriteFile(path, []byte(solved), 0o644); err != nil {
		t.Fatalf("writing solved board: %v", err)
	}

	var out bytes.Buffer
	if err := run([]string{"-json", path}, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "[]" {
		t.Fatalf("output = %q, want %q", got, "[]")
	}
}

func TestRunJSONInvalidBoardReportsError(t *testing.T) {
	broken := strings.Repeat(".", 80) + "x"
	path := filepath.Join(t.TempDir(), "broken.txt")
	if err := os.WriteFile(path, []byte(broken), 0o644); err != nil {
		t.Fatalf("writing broken board: %v", err)
	}

	var out bytes.Buffer
	err := run([]string{"-json", path}, &out)
	if err == nil {
		t.Fatalf("run: expected error, got nil")
	}
	if out.Len() != 0 {
		t.Fatalf("output = %q, want empty on error", out.String())
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
