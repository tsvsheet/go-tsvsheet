package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

func TestParseBlock_SplitsRowsAndCells(t *testing.T) {
	t.Parallel()

	assert.Equal(t, block([]string{"a", "b"}, []string{"c"}), engine.ParseBlock("a\tb\nc"))
}

func TestParseBlock_NormalizesLineEndingsAndTrailingNewline(t *testing.T) {
	t.Parallel()

	// CRLF and lone CR both split rows (the Windows and legacy-Mac clipboard
	// forms); exactly one trailing newline is ignored, so a second one is a
	// genuine empty row.
	assert.Equal(t, block([]string{"a"}, []string{"b"}), engine.ParseBlock("a\r\nb\r\n"))
	assert.Equal(t, block([]string{"a"}, []string{"b"}), engine.ParseBlock("a\rb"))
	assert.Equal(t, block([]string{"a"}, []string{""}), engine.ParseBlock("a\n\n"))
}

func TestParseBlock_EmptyTextIsASingleEmptyCell(t *testing.T) {
	t.Parallel()

	// The TSV serialization of one empty cell IS the empty string, so the
	// decode is its inverse — which is what lets a paste clear a single cell.
	assert.Equal(t, block([]string{""}), engine.ParseBlock(""))
	assert.Equal(t, block([]string{""}), engine.ParseBlock("\n"))
}

func TestParseBlock_EveryLineIsData(t *testing.T) {
	t.Parallel()

	// A clipboard block has no comment or directive semantics: a `#.` line and
	// a `# ` line are cell data, where a .tsvt parse would skip them.
	got := engine.ParseBlock("#. note\n# legacy\n#N/A")
	assert.Equal(t, block([]string{"#. note"}, []string{"# legacy"}, []string{"#N/A"}), got)
}
