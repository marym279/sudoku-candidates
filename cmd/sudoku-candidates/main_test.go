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

func TestRunGrid(t *testing.T) {
	var out bytes.Buffer
	path := writeTestBoard(t)
	if err := run([]string{"-grid", path}, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	// 9 rows of cells + 2 box separators = 29 lines.
	if len(lines) != 29 {
		t.Fatalf("got %d lines, want 29:\n%s", len(lines), out.String())
	}
	for i, line := range lines {
		wantSep := i == 9 || i == 19
		gotSep := strings.HasPrefix(line, "---")
		if gotSep != wantSep {
			t.Fatalf("line %d = %q, box separator mismatch (want separator: %v)", i, line, wantSep)
		}
	}
	// Board's (1,3) cell (1-9 coordinates) is empty with candidates 1 2 4.
	// Candidates 1 and 2 sit at pad positions 0 and 1 of the pad's top
	// row, so they render adjacently as "12".
	if !strings.Contains(lines[0], "12") {
		t.Fatalf("line 0 = %q, want it to contain candidates 1 2 for cell (1,3)", lines[0])
	}
	// A filled cell (1,1) holds a 5; it should render as a lone digit
	// with blank padding around it, not a candidate pad.
	if !strings.Contains(lines[1], " 5 ") {
		t.Fatalf("line 1 = %q, want it to contain the filled digit 5", lines[1])
	}
}

// Row 0 has every digit but 9, so (0,8) is a naked single; filling it
// with 9 has no effect on the rest of this board's column/box, so the
// only observable change is that (0,8) drops out of the report.
const singleAtCornerBoard = `
12345678.
.........
.........
.........
.........
.........
.........
.........
.........
`

func TestRunFillSinglesRemovesResolvedCellFromOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "single.txt")
	if err := os.WriteFile(path, []byte(singleAtCornerBoard), 0o644); err != nil {
		t.Fatalf("writing test board: %v", err)
	}

	var without bytes.Buffer
	if err := run([]string{"-json", path}, &without); err != nil {
		t.Fatalf("run: %v", err)
	}
	var withoutCells []jsonCell
	if err := json.Unmarshal(without.Bytes(), &withoutCells); err != nil {
		t.Fatalf("unmarshaling output %q: %v", without.String(), err)
	}

	var with bytes.Buffer
	if err := run([]string{"-fill-singles", "-json", path}, &with); err != nil {
		t.Fatalf("run: %v", err)
	}
	var withCells []jsonCell
	if err := json.Unmarshal(with.Bytes(), &withCells); err != nil {
		t.Fatalf("unmarshaling output %q: %v", with.String(), err)
	}

	if len(withCells) != len(withoutCells)-1 {
		t.Fatalf("-fill-singles: got %d cells, want %d (one fewer than %d)", len(withCells), len(withoutCells)-1, len(withoutCells))
	}
	for _, c := range withCells {
		if c.Row == 1 && c.Col == 9 {
			t.Fatalf("-fill-singles: (1,9) still reported after it should have been auto-filled")
		}
	}
}

func TestRunGridRejectsJSONAndCell(t *testing.T) {
	path := writeTestBoard(t)
	if err := run([]string{"-grid", "-json", path}, &bytes.Buffer{}); err == nil {
		t.Fatalf("run(-grid -json): expected error, got nil")
	}
	if err := run([]string{"-grid", "-cell", "1,3", path}, &bytes.Buffer{}); err == nil {
		t.Fatalf("run(-grid -cell): expected error, got nil")
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
