package tsvsheet_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	tsvsheet "github.com/tsvsheet/go-tsvsheet"
	"github.com/tsvsheet/go-tsvsheet/internal/constants"
)

// TestReExportedSentinels verifies the root package re-exports the engine's
// sentinels with identity preserved: each is the same value the engine emits
// (internal/constants), so a caller matching tsvsheet.Err* with errors.Is
// catches the internal error unchanged, and Parse's syntax failure is
// matchable through the public re-export.
func TestReExportedSentinels(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	want.Equal(constants.ErrSyntax, tsvsheet.ErrSyntax)
	want.Equal(constants.ErrInvalidValue, tsvsheet.ErrInvalidValue)
	want.Equal(constants.ErrNotFound, tsvsheet.ErrNotFound)
	want.Equal(constants.ErrReadInput, tsvsheet.ErrReadInput)
	want.Equal(constants.ErrWriteFile, tsvsheet.ErrWriteFile)

	_, err := tsvsheet.Parse([]byte("1\t2\n3\t=sum(\n")) // B2 malformed formula
	want.ErrorIs(err, tsvsheet.ErrSyntax)
	want.True(errors.Is(err, tsvsheet.ErrSyntax))
}
