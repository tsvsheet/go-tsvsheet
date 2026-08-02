package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// TestDeleteHiCollapsesARangeThatWasExactlyTheDeletedLine pins the endpoint
// arithmetic at its only interesting point. A range spanning several lines
// shrinks when one is deleted; a range that WAS the deleted line has nothing
// left to name, and must become #REF! rather than silently inverting into a
// range that reads its neighbours.
func TestDeleteHiCollapsesARangeThatWasExactlyTheDeletedLine(t *testing.T) {
	t.Parallel()
	sheet, err := engine.Parse([]byte("1\n2\n3\n=sum(A2:A2)\t=sum(A1:A3)\n"))
	require.NoError(t, err)

	edited := sheet.DeleteRow(engine.Address{Row: 1})
	computed := edited.Compute()

	assert.Equal(t, "#REF!", computed[2][0], "the range was exactly the deleted line")
	assert.Equal(t, "4", computed[2][1], "while a range around it simply shrinks")
}
