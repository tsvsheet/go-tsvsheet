package engine_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
	"github.com/tsvsheet/go-tsvsheet/internal/index"
)

func TestCells_ProjectsNonEmpty(t *testing.T) {
	t.Parallel()

	s, err := engine.Parse([]byte("a\t\t=A1\n"))
	require.NoError(t, err)

	cells := s.Cells()
	require.Len(t, cells, 2) // the empty middle cell is omitted

	assert.Equal(t, "A1", cells[0].Address.String())
	assert.Equal(t, "a", cells[0].Text)
	assert.False(t, cells[0].IsFormula)

	assert.Equal(t, "C1", cells[1].Address.String())
	assert.Equal(t, "=A1", cells[1].Text)
	assert.True(t, cells[1].IsFormula)
}

func TestSet_LiteralInPlace(t *testing.T) {
	t.Parallel()

	s, err := engine.Parse([]byte("1\t2\n"))
	require.NoError(t, err)

	next, err := s.Set(engine.Address{Row: 0, Col: 0}, "9", engine.DefaultLimits())
	require.NoError(t, err)

	// The new sheet reflects the edit; the original is unchanged (immutable Set).
	assert.Equal(t, "9", next.Source()[0][0])
	assert.Equal(t, "1", s.Source()[0][0])
	assert.Equal(t, "9", next.Compute()[0][0])
}

func TestSet_FormulaComputes(t *testing.T) {
	t.Parallel()

	s, err := engine.Parse([]byte("2\t3\n"))
	require.NoError(t, err)

	next, err := s.Set(engine.Address{Row: 0, Col: 1}, "=A1*10", engine.DefaultLimits())
	require.NoError(t, err)
	assert.Equal(t, "=A1*10", next.Source()[0][1])
	assert.Equal(t, "20", next.Compute()[0][1])
}

func TestSet_GrowsGrid(t *testing.T) {
	t.Parallel()

	s, err := engine.Parse([]byte("1\t2\n"))
	require.NoError(t, err)

	// Write well beyond the current bounds: new rows and new columns appear,
	// padded with empty cells.
	next, err := s.Set(engine.Address{Row: 2, Col: 3}, "x", engine.DefaultLimits())
	require.NoError(t, err)

	src := next.Source()
	require.Len(t, src, 3)
	assert.Equal(t, "x", src[2][3])
	assert.Equal(t, "", src[2][0])  // padded within the grown row
	assert.Equal(t, "1", src[0][0]) // original row preserved
	assert.Empty(t, src[1])         // intervening row stays empty
}

func TestSet_MalformedFormulaIsSyntaxError(t *testing.T) {
	t.Parallel()

	s, err := engine.Parse([]byte("1\n"))
	require.NoError(t, err)

	_, err = s.Set(engine.Address{Row: 0, Col: 0}, "=sum(", engine.DefaultLimits())
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrSyntax)
}

func TestSet_KeepsOtherRows(t *testing.T) {
	t.Parallel()

	// A grid taller than the edited row: growCells keeps the existing row count
	// (maxInt returns the source length, not row+1).
	s, err := engine.Parse([]byte("1\n2\n3\n"))
	require.NoError(t, err)

	next, err := s.Set(engine.Address{Row: 0, Col: 0}, "9", engine.DefaultLimits())
	require.NoError(t, err)

	src := next.Source()
	require.Len(t, src, 3)
	assert.Equal(t, "9", src[0][0])
	assert.Equal(t, "2", src[1][0])
	assert.Equal(t, "3", src[2][0])
}

func TestSet_RejectsNegativeAddress(t *testing.T) {
	t.Parallel()

	s, err := engine.Parse([]byte("1\n"))
	require.NoError(t, err)

	_, err = s.Set(engine.Address{Row: -1, Col: 0}, "x", engine.DefaultLimits())
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrInvalidValue)

	// An address beyond the grid limit is rejected before growing (OOM guard).
	_, err = s.Set(engine.Address{Row: 2_000_000, Col: 0}, "x", engine.DefaultLimits())
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrInvalidValue)
	_, err = s.Set(engine.Address{Row: 0, Col: 2_000_000}, "x", engine.DefaultLimits())
	assert.ErrorIs(t, err, constants.ErrInvalidValue)
}

func TestSource_RoundTrips(t *testing.T) {
	t.Parallel()

	s, err := engine.Parse([]byte("a\t=A1\n"))
	require.NoError(t, err)

	src := s.Source()
	assert.Equal(t, "a", src[0][0])
	assert.Equal(t, "=A1", src[0][1]) // formula source kept verbatim
}

// TestParseMapsScanFailuresToErrReadInput pins the load contract across the
// indexed read path: a source the scanner refuses (an over-long line) is the
// documented ErrReadInput, with the index layer's ErrScan preserved as its
// cause — never a bare or re-branded error.
func TestParseMapsScanFailuresToErrReadInput(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", (1<<20)+2) + "\n"
	_, err := engine.Parse([]byte(long))
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrReadInput)
	assert.ErrorIs(t, err, index.ErrScan, "the scan cause survives the mapping")
}

// TestParseWithRefusesOverBudgetBeforeAnyCellParses pins the 018 census gate
// with its own discriminator: the over-budget document carries a MALFORMED
// formula, and the refusal is ErrDocTooLarge — not ErrSyntax — because the
// gate reads bytes, never cells; nothing materialized. The refusal names the
// census and the budget.
func TestParseWithRefusesOverBudgetBeforeAnyCellParses(t *testing.T) {
	t.Parallel()

	src := []byte("a\tb\n=)(malformed\td\n")
	_, err := engine.ParseWith(src, engine.Limits{ResidentCells: 3})
	require.ErrorIs(t, err, constants.ErrDocTooLarge)
	assert.NotErrorIs(t, err, constants.ErrSyntax, "the gate must precede cell parsing")
	assert.Contains(t, err.Error(), "cells", "the refusal names the census")

	_, err = engine.Parse(src)
	assert.ErrorIs(t, err, constants.ErrSyntax, "unbounded Parse still reaches the malformed cell")
}

// TestParseWithBoundaryMatchesParse pins the exact ceiling: a document AT the
// resident budget parses byte-identically to Parse; one cell over refuses.
func TestParseWithBoundaryMatchesParse(t *testing.T) {
	t.Parallel()

	src := []byte("a\tb\nc\td\n") // 4 cells
	bounded, err := engine.ParseWith(src, engine.Limits{ResidentCells: 4})
	require.NoError(t, err)
	unbounded, err := engine.Parse(src)
	require.NoError(t, err)
	assert.Equal(t, unbounded.Source(), bounded.Source(), "at the budget the bounded parse is Parse")

	_, err = engine.ParseWith(src, engine.Limits{ResidentCells: 3})
	assert.ErrorIs(t, err, constants.ErrDocTooLarge, "one cell over refuses")

	_, err = engine.ParseWith(src, engine.Limits{})
	require.NoError(t, err, "the zero value falls back to DefaultLimits, not refuse-everything")
}

// TestParseWithMapsScanFailures pins the gate's failure path: a source the
// scanner itself refuses (a line past the ceiling) is ErrReadInput, exactly
// as Parse reports it.
func TestParseWithMapsScanFailures(t *testing.T) {
	t.Parallel()

	long := append(bytes.Repeat([]byte{'a'}, (1<<20)+2), '\n')
	_, err := engine.ParseWith(long, engine.Limits{ResidentCells: 100})
	assert.ErrorIs(t, err, constants.ErrReadInput)
}
