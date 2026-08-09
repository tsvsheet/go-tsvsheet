package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSpill_ArrayFillsEmptyCells states the spill contract at its own seam: an
// array result writes from its anchor into empty cells, and every spilled cell
// renders the array's value rather than its own emptiness.
func TestSpill_ArrayFillsEmptyCells(t *testing.T) {
	t.Parallel()

	g := compute(t, "=sequence(3)\t\n\t\n\t\n")
	assert.Equal(t, "1", cellAt(t, g, 0, 0))
	assert.Equal(t, "2", cellAt(t, g, 1, 0))
	assert.Equal(t, "3", cellAt(t, g, 2, 0))
}

// TestSpill_BlockedByOccupiedCellIsSpillError states the refusal: a spill
// whose target holds content answers #SPILL! at the anchor and writes nothing
// over the occupant.
func TestSpill_BlockedByOccupiedCellIsSpillError(t *testing.T) {
	t.Parallel()

	g := compute(t, "=sequence(3)\t\nblocker\t\n\t\n")
	assert.Equal(t, "#SPILL!", cellAt(t, g, 0, 0))
	assert.Equal(t, "blocker", cellAt(t, g, 1, 0), "the occupant is never overwritten")
}

// TestSpill_AnchorCellDoesNotBlockItself states the anchor rule: the formula's
// own cell is the spill's first target and never counts as an occupant.
func TestSpill_AnchorCellDoesNotBlockItself(t *testing.T) {
	t.Parallel()

	g := compute(t, "=sequence(2)\n\n")
	assert.Equal(t, "1", cellAt(t, g, 0, 0))
	assert.Equal(t, "2", cellAt(t, g, 1, 0))
}
