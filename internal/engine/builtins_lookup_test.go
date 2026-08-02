package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// lookup evaluates a formula against a 3x2 table in A1:B3 (names, scores); the
// formula sits in row 4.
func lookup(t *testing.T, formula string) string {
	t.Helper()
	src := "Alice\t85\nBob\t72\nCarol\t95\n=" + formula + "\n"
	return cellAt(t, compute(t, src), 3, 0)
}

func TestLookup_VlookupIndexMatch(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`vlookup("Bob", A1:B3, 2)`: "72",
		`vlookup("Zoe", A1:B3, 2)`: string(engine.ErrNA),  // not found
		`vlookup("Bob", A1:B3, 9)`: string(engine.ErrRef), // column out of range
		`index(A1:B3, 2, 2)`:       "72",
		`index(A1:B3, 2)`:          "Bob", // column defaults to 1
		`index(A1:B3, 9, 1)`:       string(engine.ErrRef),
		`index(A1, 1, 1)`:          "Alice", // single-cell range
		`index(5, 1, 1)`:           "5",     // scalar becomes a 1x1
		`match("Carol", A1:A3)`:    "3",
		`match("Zoe", A1:A3)`:      string(engine.ErrNA),
		`rows(A1:B3)`:              "3",
		`columns(A1:B3)`:           "2",
		`choose(2, "x", "y", "z")`: "y",
		`index(A0:A0, 1, 1)`:       string(engine.ErrRef), // off-grid range
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, lookup(t, expr))
		})
	}
}

func TestLookup_Hlookup(t *testing.T) {
	t.Parallel()

	// A 2x3 horizontal table: header row of keys, values below.
	g := compute(
		t,
		"a\tb\tc\n1\t2\t3\n=hlookup(\"b\", A1:C2, 2)\t=hlookup(\"z\", A1:C2, 2)\t=hlookup(\"a\", A1:C2, 9)\n",
	)
	assert.Equal(t, "2", cellAt(t, g, 2, 0))                   // found
	assert.Equal(t, string(engine.ErrNA), cellAt(t, g, 2, 1))  // not found
	assert.Equal(t, string(engine.ErrRef), cellAt(t, g, 2, 2)) // row out of range
}

func TestLookup_ArityAndArgErrors(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`rows(A1:B3, 1)`:   string(engine.ErrValue), // rows takes one arg
		`index(A1:B3)`:     string(engine.ErrValue), // too few
		`match("x")`:       string(engine.ErrValue), // too few
		`vlookup("x", A1)`: string(engine.ErrValue), // too few
		`choose(0, "a")`:   string(engine.ErrValue), // index below 1
		`choose(9, "a")`:   string(engine.ErrValue), // index past end
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, lookup(t, expr))
		})
	}

	// Non-numeric index arguments propagate #VALUE! (A4 holds text "x").
	for _, expr := range []string{
		`=index(A1:B3, A4, 1)`, `=index(A1:B3, 1, A4)`, `=vlookup("Bob", A1:B3, A4)`, `=choose(A4, "a")`,
	} {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			src := "Alice\t85\nBob\t72\nCarol\t95\nx\t" + expr + "\n"
			assert.Equal(t, "#VALUE!", cellAt(t, compute(t, src), 3, 1))
		})
	}
}

// TestLookup_RefusedRangePropagates pins that a wholesale range refusal
// propagates through every lookup shape instead of laundering: before the
// out-of-band refusal, vlookup/match over an unreadable table answered #N/A
// ("searched and found nothing") and rows() answered 1 — both plausible and
// both false, since nothing was ever searched or counted.
func TestLookup_RefusedRangePropagates(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"vlookup(2, A1:B50000000000, 2)":       string(engine.ErrLimit),
		"hlookup(2, A1:ZZZZZZ2, 2)":            string(engine.ErrLimit),
		"match(2, A1:A50000000000)":            string(engine.ErrLimit),
		"index(A1:A50000000000, 1)":            string(engine.ErrLimit),
		"rows(A1:A50000000000)":                string(engine.ErrLimit),
		"columns(A1:DEJTLX1)":                  string(engine.ErrLimit),
		"vlookup(2, ZZZZZZZZZ1:ZZZZZZZZZ9, 1)": string(engine.ErrRef),
		"match(2, \"missing\"!A1:A9)":          string(engine.ErrRef),
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, formula1(t, expr))
		})
	}
}

// TestLookup_ErrorHolesInAResolvedTableAreScannedPast pins the hole tolerance
// the refusal distinction protects: a lookup whose table RESOLVES scans past a
// #REF! hole from a ragged grid and can still match — the reason a wholesale
// refusal cannot simply be spelled as a 1×1 error block.
func TestLookup_ErrorHolesInAResolvedTableAreScannedPast(t *testing.T) {
	t.Parallel()

	// Row 2 is shorter than row 1: B2 is a #REF! hole in A1:B2, but the key 1
	// matches A1 before any hole is touched.
	g := compute(t, "1\thit\t=vlookup(1, A1:B2, 2)\n9\n")
	assert.Equal(t, "hit", cellAt(t, g, 0, 2))
}
