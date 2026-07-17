package engine_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

func TestDefaultLimits_Values(t *testing.T) {
	t.Parallel()

	l := engine.DefaultLimits()
	assert.Equal(t, 5_000_000, l.ResultCells)
	assert.Equal(t, 1_000_000, l.GridDim)
	assert.Equal(t, 1<<20, l.ResultBytes)
}

func TestBrowserLimits_Values(t *testing.T) {
	t.Parallel()

	l := engine.BrowserLimits()
	assert.Equal(t, 100_000, l.ResultCells)
	assert.Equal(t, 20_000, l.GridDim)
	assert.Equal(t, 64<<10, l.ResultBytes)
}

// computeWithCell parses src, computes it with the injected limits, and returns
// cell A1's computed value.
func computeWithCell(t *testing.T, src string, limits engine.Limits) string {
	t.Helper()
	s, err := engine.Parse([]byte(src))
	require.NoError(t, err)
	return s.ComputeWith(engine.ComputeOptions{At: time.Now(), Limits: limits})[0][0]
}

func TestComputeWith_ZeroLimitsFallBackToDefault(t *testing.T) {
	t.Parallel()

	// A zero (unset) Limits is treated as DefaultLimits: a modest array and a
	// modest REPT both compute — a degenerate zero cap would reject them.
	assert.Equal(t, "1", computeWithCell(t, "=sequence(2, 2)\n", engine.Limits{}))
	assert.Equal(t, "aaa", computeWithCell(t, "=rept(\"a\", 3)\n", engine.Limits{}))
}

func TestComputeWith_HonorsInjectedCellBudget(t *testing.T) {
	t.Parallel()

	tiny := engine.Limits{ResultCells: 5, GridDim: 5, ResultBytes: 5}
	assert.Equal(t, "1", computeWithCell(t, "=sequence(2, 2)\n", tiny))                     // 4 cells <= 5
	assert.Equal(t, string(engine.ErrValue), computeWithCell(t, "=sequence(3, 3)\n", tiny)) // 9 > 5
}

func TestComputeWith_HonorsInjectedByteBudget(t *testing.T) {
	t.Parallel()

	tiny := engine.Limits{ResultCells: 5, GridDim: 5, ResultBytes: 5}
	assert.Equal(t, "aaaaa", computeWithCell(t, "=rept(\"a\", 5)\n", tiny))                 // 5 bytes <= 5
	assert.Equal(t, string(engine.ErrValue), computeWithCell(t, "=rept(\"a\", 6)\n", tiny)) // 6 > 5
}

func TestSet_HonorsInjectedGridLimit(t *testing.T) {
	t.Parallel()

	s, err := engine.Parse([]byte("1\n"))
	require.NoError(t, err)

	tiny := engine.Limits{ResultCells: 5, GridDim: 5, ResultBytes: 5}
	// Within the grid dimension the edit grows the grid; at or beyond it the edit
	// is rejected before growing (the OOM guard).
	_, err = s.Set(engine.Address{Row: 4, Col: 0}, "x", tiny)
	require.NoError(t, err)
	_, err = s.Set(engine.Address{Row: 5, Col: 0}, "x", tiny)
	assert.ErrorIs(t, err, constants.ErrInvalidValue)
}
