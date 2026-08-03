// The 018 bounded-parse fuzz oracle, split from fuzz_test.go for size: the
// bounded forms must never behave differently than Parse except by refusing.
package tsvsheet_test

import (
	"errors"
	"testing"

	tsvsheet "github.com/tsvsheet/go-tsvsheet"
)

// FuzzParseWith is the 018 bounded-parse parity oracle: under a generous
// budget the bounded parse must agree with Parse exactly — same error or same
// grid — and under a one-cell budget it must either refuse as ErrDocTooLarge
// or agree with Parse (a document of at most one cell), never a third
// behavior.
func FuzzParseWith(f *testing.F) {
	f.Add([]byte("1\t=A1+1\n=sum(A1:B1)\tx\n"))
	f.Add([]byte("#. comment\na\n"))
	f.Add([]byte(""))
	f.Fuzz(func(t *testing.T, data []byte) {
		want, wantErr := tsvsheet.Parse(data)
		got, gotErr := tsvsheet.ParseWith(data, tsvsheet.DefaultLimits())
		if (wantErr == nil) != (gotErr == nil) {
			t.Fatalf("generous budget diverged from Parse: %v vs %v", wantErr, gotErr)
		}
		if wantErr == nil && !equalGrids(want.Source(), got.Source()) {
			t.Fatal("generous budget parsed a different grid than Parse")
		}
		_, tightErr := tsvsheet.ParseWith(data, tsvsheet.Limits{ResidentCells: 1})
		if tightErr != nil && !errors.Is(tightErr, tsvsheet.ErrDocTooLarge) && wantErr == nil {
			t.Fatalf("tight budget invented an error Parse does not have: %v", tightErr)
		}
	})
}

// equalGrids reports deep equality of two grids.
func equalGrids(a, b tsvsheet.Grid) bool {
	if len(a) != len(b) {
		return false
	}
	for r := range a {
		if len(a[r]) != len(b[r]) {
			return false
		}
		for c := range a[r] {
			if a[r][c] != b[r][c] {
				return false
			}
		}
	}
	return true
}
