package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// TestComputeRowsAnswersARefusedCellDeterministically pins the refused phase:
// a cell past the touched budget answers #LIMIT! on its first read AND on
// every re-read — the memo remembers the refusal rather than re-deciding.
func TestComputeRowsAnswersARefusedCellDeterministically(t *testing.T) {
	t.Parallel()

	src := "1\t2\t3\n=sum(C1, C1)\tz\n" // sum collects both reads; + would short-circuit the second
	starved := engine.Limits{
		ResultCells: 100, GridDim: 100, ResultBytes: 100,
		ResidentCells: 1, TouchedCells: 1,
	}
	_, windowed := open(t, src, starved)
	require.NotNil(t, windowed)
	got, err := windowed.ComputeRows(1, 1, engine.ComputeOptions{Limits: starved})
	require.NoError(t, err)
	assert.Equal(t, string(engine.ErrLimit), got[0][0],
		"the second read of the refused C1 hits the remembered refusal")
}
