package tsvsheet_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uplang/go-tsvsheet"
	"github.com/uplang/go-tsvsheet/internal/constants"
)

func TestDefaultLimits_Values(t *testing.T) {
	t.Parallel()

	l := tsvsheet.DefaultLimits()
	assert.Equal(t, 5_000_000, l.ResultCells)
	assert.Equal(t, 1_000_000, l.GridDim)
	assert.Equal(t, 1<<20, l.ResultBytes)
}

func TestBrowserLimits_Values(t *testing.T) {
	t.Parallel()

	l := tsvsheet.BrowserLimits()
	assert.Equal(t, 100_000, l.ResultCells)
	assert.Equal(t, 20_000, l.GridDim)
	assert.Equal(t, 64<<10, l.ResultBytes)
}

// computeWithCell parses src, computes it with the injected limits, and returns
// cell A1's computed value.
func computeWithCell(t *testing.T, src string, limits tsvsheet.Limits) string {
	t.Helper()
	s, err := tsvsheet.Parse([]byte(src))
	require.NoError(t, err)
	return s.ComputeWith(tsvsheet.ComputeOptions{At: time.Now(), Limits: limits})[0][0]
}

func TestComputeWith_ZeroLimitsFallBackToDefault(t *testing.T) {
	t.Parallel()

	// A zero (unset) Limits is treated as DefaultLimits: a modest array and a
	// modest REPT both compute — a degenerate zero cap would reject them.
	assert.Equal(t, "1", computeWithCell(t, "=sequence(2, 2)\n", tsvsheet.Limits{}))
	assert.Equal(t, "aaa", computeWithCell(t, "=rept(\"a\", 3)\n", tsvsheet.Limits{}))
}

func TestComputeWith_HonorsInjectedCellBudget(t *testing.T) {
	t.Parallel()

	tiny := tsvsheet.Limits{ResultCells: 5, GridDim: 5, ResultBytes: 5}
	assert.Equal(t, "1", computeWithCell(t, "=sequence(2, 2)\n", tiny))                       // 4 cells <= 5
	assert.Equal(t, string(tsvsheet.ErrValue), computeWithCell(t, "=sequence(3, 3)\n", tiny)) // 9 > 5
}

func TestComputeWith_HonorsInjectedByteBudget(t *testing.T) {
	t.Parallel()

	tiny := tsvsheet.Limits{ResultCells: 5, GridDim: 5, ResultBytes: 5}
	assert.Equal(t, "aaaaa", computeWithCell(t, "=rept(\"a\", 5)\n", tiny))                   // 5 bytes <= 5
	assert.Equal(t, string(tsvsheet.ErrValue), computeWithCell(t, "=rept(\"a\", 6)\n", tiny)) // 6 > 5
}

func TestSet_HonorsInjectedGridLimit(t *testing.T) {
	t.Parallel()

	s, err := tsvsheet.Parse([]byte("1\n"))
	require.NoError(t, err)

	tiny := tsvsheet.Limits{ResultCells: 5, GridDim: 5, ResultBytes: 5}
	// Within the grid dimension the edit grows the grid; at or beyond it the edit
	// is rejected before growing (the OOM guard).
	_, err = s.Set(tsvsheet.Address{Row: 4, Col: 0}, "x", tiny)
	require.NoError(t, err)
	_, err = s.Set(tsvsheet.Address{Row: 5, Col: 0}, "x", tiny)
	assert.ErrorIs(t, err, constants.ErrInvalidValue)
}
