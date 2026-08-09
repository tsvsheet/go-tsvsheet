package engine_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// span builds a fill target from 0-based corner coordinates.
func span(loRow, loCol, hiRow, hiCol int) engine.Span {
	return engine.Span{
		From: engine.Address{Row: loRow, Col: loCol},
		To:   engine.Address{Row: hiRow, Col: hiCol},
	}
}

func TestFill_RelativeReferencesShift(t *testing.T) {
	t.Parallel()

	// B1's formula fills down two rows: the unpinned A1 follows each target.
	s := parse(t, "10\t=A1*2\n20\n30\n")
	got := s.Fill(addr(0, 1), span(1, 1, 2, 1))

	assert.Equal(t, "=A2 * 2", sourceAt(t, got, 1, 1))
	assert.Equal(t, "=A3 * 2", sourceAt(t, got, 2, 1))
	g := got.Compute()
	assert.Equal(t, "40", cellAt(t, g, 1, 1))
	assert.Equal(t, "60", cellAt(t, g, 2, 1))
}

func TestFill_PinnedCoordinatesHold(t *testing.T) {
	t.Parallel()

	// $A$1 is fully pinned; the bare A1 shifts; the range's pinned endpoint
	// holds while its unpinned endpoint follows ($B$1:B1 → $B$1:B2).
	s := parse(t, "5\t7\t=$A$1+A1\t=sum($B$1:B1)\n6\t8\n")
	got := s.Fill(addr(0, 2), span(1, 2, 1, 2)).Fill(addr(0, 3), span(1, 3, 1, 3))

	assert.Equal(t, "=$A$1 + A2", sourceAt(t, got, 1, 2))
	assert.Equal(t, "=sum($B$1:B2)", sourceAt(t, got, 1, 3))
	g := got.Compute()
	assert.Equal(t, "11", cellAt(t, g, 1, 2)) // 5 + 6
	assert.Equal(t, "15", cellAt(t, g, 1, 3)) // 7 + 8
}

func TestFill_MixedPinsShiftPerAxis(t *testing.T) {
	t.Parallel()

	// B$1 pins the row and follows the column; $A2 pins the column and follows
	// the row — filled diagonally, each axis moves independently.
	s := parse(t, "1\t2\n3\t=B$1+$A2\n")
	got := s.Fill(addr(1, 1), span(2, 2, 2, 2))

	assert.Equal(t, "=C$1 + $A3", sourceAt(t, got, 2, 2))
}

func TestFill_OffGridBecomesRefError(t *testing.T) {
	t.Parallel()

	// Filling B2 upward pushes its A1 reference above row 1; filling leftward
	// pushes it before column A. Both render #REF! and compute #REF!.
	s := parse(t, "1\t2\n3\t=A1\n")
	up := s.Fill(addr(1, 1), span(0, 1, 0, 1))
	left := s.Fill(addr(1, 1), span(1, 0, 1, 0))

	assert.Equal(t, "=#REF!", sourceAt(t, up, 0, 1))
	assert.Equal(t, "=#REF!", sourceAt(t, left, 1, 0))
	assert.Equal(t, "#REF!", cellAt(t, up.Compute(), 0, 1))
}

func TestFill_CrossSheetReferencesCopyUnshifted(t *testing.T) {
	t.Parallel()

	// A cross-sheet reference addresses another sheet: the fill never rebases
	// it, mirroring the structural-edit ruling.
	s := parse(t, "=\"rates.tsvt\"!B2\nx\n")
	got := s.Fill(addr(0, 0), span(1, 0, 1, 0))

	assert.Equal(t, `="rates.tsvt"!B2`, sourceAt(t, got, 1, 0))
}

func TestFill_LiteralAndEmptySources(t *testing.T) {
	t.Parallel()

	// A literal copies verbatim; an empty source (beyond the grid) clears its
	// target, as copying an empty cell does.
	s := parse(t, "hello\t=A1\n")
	lit := s.Fill(addr(0, 0), span(0, 1, 0, 1))
	cleared := s.Fill(addr(5, 5), span(0, 1, 0, 1))

	assert.Equal(t, "hello", sourceAt(t, lit, 0, 1))
	assert.Equal(t, "", sourceAt(t, cleared, 0, 1))
}

func TestFill_SourceInsideSpanIsSkipped(t *testing.T) {
	t.Parallel()

	// A span containing the source fills around it: the source keeps its
	// original spelling byte-for-byte (no self-rewrite to canonical form).
	s := parse(t, "1\n=A1*2\n2\n")
	got := s.Fill(addr(1, 0), span(1, 0, 2, 0))

	assert.Equal(t, "=A1*2", sourceAt(t, got, 1, 0)) // untouched original
	assert.Equal(t, "=A2 * 2", sourceAt(t, got, 2, 0))
}

func TestFill_UnorderedCornersAndNoOps(t *testing.T) {
	t.Parallel()

	// Corners in any order describe the same span; a negative source or span
	// corner is a no-op, mirroring the structural quartet.
	s := parse(t, "1\t=A1\n2\n3\n")
	reversed := s.Fill(addr(0, 1), span(2, 1, 1, 1))
	assert.Equal(t, "=A2", sourceAt(t, reversed, 1, 1))
	assert.Equal(t, "=A3", sourceAt(t, reversed, 2, 1))

	assert.Equal(t, s.Source(), s.Fill(addr(-1, 0), span(0, 0, 0, 0)).Source())
	assert.Equal(t, s.Source(), s.Fill(addr(0, 0), span(-1, 0, 0, 0)).Source())
}

func TestFill_GrowsTheGrid(t *testing.T) {
	t.Parallel()

	// Targets beyond the grid grow it minimally, as Set's growth does.
	s := parse(t, "=1+1\n")
	got := s.Fill(addr(0, 0), span(2, 1, 2, 1))

	require.Len(t, got.Source(), 3)
	assert.Equal(t, "=1 + 1", sourceAt(t, got, 2, 1))
	assert.Equal(t, "2", cellAt(t, got.Compute(), 2, 1))
}

func TestFill_PipeSpellingPreserved(t *testing.T) {
	t.Parallel()

	// The canonical re-render keeps the author's pipe spelling.
	s := parse(t, "3.14159\t=A1 | round(2)\nx\n")
	got := s.Fill(addr(0, 1), span(1, 1, 1, 1))

	assert.Equal(t, "=A2 | round(2)", sourceAt(t, got, 1, 1))
}

func TestDocument_FillGrowsLayout(t *testing.T) {
	t.Parallel()

	// Filling past the last row appends row markers after trailing comments,
	// and the comment lines survive serialization.
	doc, err := engine.ParseDocument([]byte("# note\n=1+1\n"))
	require.NoError(t, err)
	got := doc.Fill(engine.Address{Row: 0, Col: 0}, span(1, 0, 2, 0))

	assert.Equal(t, "# note\n=1+1\n=1 + 1\n=1 + 1\n", string(got.Text()))
}

// TestFill_ReproducesShippedExamples is the 007 acceptance round-trip: filling
// the anchor formulas of the shipped tsvsheet.examples sheets (copied verbatim
// into testdata) over their hand-authored rows reproduces each file
// byte-for-byte — fill authors exactly what was written by hand.
func TestFill_ReproducesShippedExamples(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		file  string
		fills []struct{ from, lo, hi engine.Address }
	}{
		"bounce": {file: "testdata/bounce.tsvt", fills: []struct{ from, lo, hi engine.Address }{
			{engine.Address{Row: 1, Col: 1}, engine.Address{Row: 2, Col: 1}, engine.Address{Row: 8, Col: 1}},
			{engine.Address{Row: 1, Col: 2}, engine.Address{Row: 2, Col: 2}, engine.Address{Row: 8, Col: 2}},
		}},
		"flipbook": {file: "testdata/flipbook.tsvt", fills: []struct{ from, lo, hi engine.Address }{
			{engine.Address{Row: 1, Col: 0}, engine.Address{Row: 2, Col: 0}, engine.Address{Row: 3, Col: 0}},
		}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			src, err := os.ReadFile(tc.file)
			require.NoError(t, err)
			doc, err := engine.ParseDocument(src)
			require.NoError(t, err)
			for _, f := range tc.fills {
				doc = doc.Fill(f.from, engine.Span{From: f.lo, To: f.hi})
			}
			assert.Equal(t, string(src), string(doc.Text()))
		})
	}
}

// TestCloneRowKeepsAFilledSheetFromAliasingItsSource pins the copy. A fill that
// aliased the source row would make a later edit to one row silently rewrite
// the other — the aliasing bug that stays invisible until someone edits the
// copy and watches the original change under them.
func TestCloneRowKeepsAFilledSheetFromAliasingItsSource(t *testing.T) {
	t.Parallel()
	sheet, err := engine.Parse([]byte("7\t8\n\t\n"))
	require.NoError(t, err)

	filled := sheet.Fill(addr(0, 0), span(1, 0, 1, 1))
	edited, err := filled.Set(addr(1, 0), "99", engine.DefaultLimits())
	require.NoError(t, err)

	assert.Equal(t, "7", edited.Source()[0][0], "editing the copy leaves the source row alone")
	assert.Equal(t, "99", edited.Source()[1][0])
}

// TestNames_FillDropsTheClause is the 023 fill ruling: the expression is
// content and fills; the clause is identity and stays with the cell it was
// written in. The filled copy is unnamed and the original keeps the name.
func TestNames_FillDropsTheClause(t *testing.T) {
	t.Parallel()

	s, err := engine.Parse([]byte("10\t=A1*2 |@ named(Double)\n20\t\n"))
	require.NoError(t, err)

	filled := s.Fill(engine.Address{Row: 0, Col: 1}, engine.Span{
		From: engine.Address{Row: 1, Col: 1},
		To:   engine.Address{Row: 1, Col: 1},
	})
	src := filled.Source()
	assert.Equal(t, "=A1*2 |@ named(Double)", src[0][1], "the original keeps its name")
	assert.Equal(t, "=A2 * 2", src[1][1], "the copy is the rebased expression, unnamed")
}

// TestNames_DuplicateRowDropsTheClause states that duplication is a copy, and
// a copy takes the expression alone — the duplicate binds nothing, so the
// sheet gains no duplicate-name defect from the gesture.
func TestNames_DuplicateRowDropsTheClause(t *testing.T) {
	t.Parallel()

	s, err := engine.Parse([]byte("=1 |@ named(X)\n"))
	require.NoError(t, err)

	src := s.DuplicateRow(engine.Address{Row: 0}).Source()
	assert.Equal(t, "=1 |@ named(X)", src[0][0])
	assert.Equal(t, "=1", src[1][0], "the duplicate is unnamed")
}

// TestNames_PasteDropsTheClause is the paste half of the ruling (paste
// behaves as fill does): pasted text carrying a clause lands without it.
func TestNames_PasteDropsTheClause(t *testing.T) {
	t.Parallel()

	s, err := engine.Parse([]byte("a\tb\n"))
	require.NoError(t, err)

	pasted, err := s.Paste(engine.Address{Row: 1, Col: 0}, engine.Address{Row: 1, Col: 0},
		engine.Grid{{"=1 |@ named(X)"}}, engine.DefaultLimits())
	require.NoError(t, err)
	assert.Equal(t, "=1", pasted.Source()[1][0], "the clause does not travel through a paste")
}

// TestNames_FillLeavesANameReferenceAlone states Decision 5 of the 023 design:
// a name is absolute — there is nothing in `@Rate` for a fill to shift — so
// the copy rebases the cell references around it and re-renders the name
// verbatim.
func TestNames_FillLeavesANameReferenceAlone(t *testing.T) {
	t.Parallel()

	s, err := engine.Parse([]byte("5\t=@Rate + A1\n7\t\n=2 |@ named(Rate)\n"))
	require.NoError(t, err)

	filled := s.Fill(engine.Address{Row: 0, Col: 1}, engine.Span{
		From: engine.Address{Row: 1, Col: 1},
		To:   engine.Address{Row: 1, Col: 1},
	})
	assert.Equal(t, "=@Rate + A2", filled.Source()[1][1],
		"the cell reference shifted; the name did not")
}
