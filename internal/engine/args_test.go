package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// How a call's arguments are built from its operands. The rule that matters is
// positional: a slot declared to take cells swallows a whole range, and a slot
// declared to take one value takes one — so a multi-cell operand cannot push
// the arguments after it out of position.

// computedRow parses a one-row sheet and returns its computed cells.
func computedRow(t *testing.T, source string) []string {
	t.Helper()
	sheet, err := engine.Parse([]byte(source))
	require.NoError(t, err)
	return sheet.Compute()[0]
}

func TestArgValuesKeepsAMultiCellOperandFromShiftingLaterArguments(t *testing.T) {
	// go-tsvsheet#2: round's second slot is scalar, so a range in the first
	// slot must not slide the digit count into the range's place. If it did,
	// round(A1:B1, 1) would round to "B1 digits" and give a different answer.
	got := computedRow(t, "1.23456\t2.34567\t=round(sum(A1:B1), 2)\t=round(A1, 2)")

	assert.Equal(t, "3.58", got[2], "the range is summed, then rounded to 2 places")
	assert.Equal(t, "1.23", got[3])
}

func TestArgModeGivesAScalarSlotExactlyOneValue(t *testing.T) {
	// A scalar slot takes one value and says so when handed several: the
	// multi-cell operand is refused rather than expanded across the remaining
	// slots, which is the same guarantee seen from the failing side.
	got := computedRow(t, "5\t7\t=round(A1:B1, 0)\t=abs(A1:B1)\t=round(A1, 0)")

	assert.Equal(t, "#VALUE!", got[2], "two cells are not one value")
	assert.Equal(t, "#VALUE!", got[3])
	assert.Equal(t, "5", got[4], "and one cell in the same slot computes")
}

func TestArgCellsExpandsARangeAndAnArrayAlike(t *testing.T) {
	// An expression that evaluates to an array is consumed exactly like a
	// range, so an aggregate over a sorted block aggregates the block.
	got := computedRow(t, "3\t1\t2\t=sum(A1:C1)\t=sum(sort(A1:C1))\t=count(sort(A1:C1))")

	assert.Equal(t, got[3], got[4], "a range and the array of the same cells aggregate alike")
	assert.Equal(t, "6", got[4])
	assert.Equal(t, "3", got[5])
}
