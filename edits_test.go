package tsvsheet_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tsvsheet "github.com/tsvsheet/go-tsvsheet"
)

func TestFacadeRevisionAndApply(t *testing.T) {
	doc, err := tsvsheet.ParseDocument([]byte("1\t2\n"))
	require.NoError(t, err)
	edits, err := tsvsheet.ParseEdits([]byte("#.base\t" + string(tsvsheet.Revision(doc)) + "\nsetCell\tB1\t9\n"))
	require.NoError(t, err)
	assert.Equal(t, 1, edits.Len())
	out, err := tsvsheet.Apply(doc, edits, tsvsheet.DefaultLimits())
	require.NoError(t, err)
	assert.Equal(t, "1\t9\n", string(out.Text()))
	assert.Len(t, string(tsvsheet.Revision(out)), 64)
}

func TestFacadeEditsSentinels(t *testing.T) {
	doc, err := tsvsheet.ParseDocument([]byte("1\n"))
	require.NoError(t, err)
	for name, tc := range map[string]struct {
		want  error
		edits string
	}{
		"unknown op":    {edits: "nope\tA1\n", want: tsvsheet.ErrEditsOp},
		"arity":         {edits: "insertRow\n", want: tsvsheet.ErrEditsArity},
		"address":       {edits: "deleteRow\tx\n", want: tsvsheet.ErrEditsAddress},
		"block":         {edits: "paste\tA1\tA1\t!!!\n", want: tsvsheet.ErrEditsBlock},
		"base mismatch": {edits: "#.base\t" + strings.Repeat("f", 64) + "\nsetCell\tA1\tx\n", want: tsvsheet.ErrEditsBase},
	} {
		t.Run(name, func(t *testing.T) {
			edits, err := tsvsheet.ParseEdits([]byte(tc.edits))
			if err == nil {
				_, err = tsvsheet.Apply(doc, edits, tsvsheet.DefaultLimits())
			}
			require.Error(t, err)
			assert.ErrorIs(t, err, tc.want)
		})
	}
}

func TestFacadeEditsApplySentinel(t *testing.T) {
	doc, err := tsvsheet.ParseDocument([]byte("1\n"))
	require.NoError(t, err)
	edits, err := tsvsheet.ParseEdits([]byte("setCell\tZZ99\tx\n"))
	require.NoError(t, err)
	_, err = tsvsheet.Apply(doc, edits, tsvsheet.Limits{ResultCells: 1, GridDim: 1, ResultBytes: 1})
	require.Error(t, err)
	assert.ErrorIs(t, err, tsvsheet.ErrEditsApply)
}

// TestParseEditsWith_FacadeGates pins the public bounded edits parse:
// over-budget batches refuse matchable as ErrDocTooLarge; in-budget batches
// parse as ParseEdits does.
func TestParseEditsWith_FacadeGates(t *testing.T) {
	t.Parallel()

	over := []byte("setCell\tA1\t1\nsetCell\tA2\t2\n")
	_, err := tsvsheet.ParseEditsWith(over, tsvsheet.Limits{ResidentCells: 1})
	require.ErrorIs(t, err, tsvsheet.ErrDocTooLarge)

	batch, err := tsvsheet.ParseEditsWith([]byte("setCell\tA1\t1\n"), tsvsheet.DefaultLimits())
	require.NoError(t, err)
	assert.Equal(t, 1, batch.Len())
}
