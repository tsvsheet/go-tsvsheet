package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

func TestImportCellScalar(t *testing.T) {
	t.Parallel()

	var accept engine.MediaType
	grid := importGrid(
		t,
		"=importcell(\"u\")\n",
		echoFetcher{body: []byte("42\n"), accept: &accept},
		engine.DefaultLimits(),
	)
	assert.Equal(t, "42", cellAt(t, grid, 0, 0))
	assert.Equal(t, mediaCellWire, accept, "IMPORTCELL must request the cell media type")
}

func TestImportRowSpillsHorizontally(t *testing.T) {
	t.Parallel()

	var accept engine.MediaType
	grid := importGrid(
		t,
		"=importrow(\"u\")\n",
		echoFetcher{body: []byte("a\tb\tc\n"), accept: &accept},
		engine.DefaultLimits(),
	)
	assert.Equal(t, "a", cellAt(t, grid, 0, 0))
	assert.Equal(t, "b", cellAt(t, grid, 0, 1))
	assert.Equal(t, "c", cellAt(t, grid, 0, 2))
	assert.Equal(t, mediaRowWire, accept, "IMPORTROW must request the row media type")
}

func TestImportColumnSpillsVertically(t *testing.T) {
	t.Parallel()

	var accept engine.MediaType
	grid := importGrid(
		t,
		"=importcolumn(\"u\")\n",
		echoFetcher{body: []byte("x\ny\nz\n"), accept: &accept},
		engine.DefaultLimits(),
	)
	assert.Equal(t, "x", cellAt(t, grid, 0, 0))
	assert.Equal(t, "y", cellAt(t, grid, 1, 0))
	assert.Equal(t, "z", cellAt(t, grid, 2, 0))
	assert.Equal(t, mediaColumnWire, accept, "IMPORTCOLUMN must request the column media type")
}

func TestImportRangeSpillsRectangle(t *testing.T) {
	t.Parallel()

	var accept engine.MediaType
	grid := importGrid(
		t,
		"=importrange(\"u\")\n",
		echoFetcher{body: []byte("1\t2\n3\t4\n"), accept: &accept},
		engine.DefaultLimits(),
	)
	assert.Equal(t, "1", cellAt(t, grid, 0, 0))
	assert.Equal(t, "2", cellAt(t, grid, 0, 1))
	assert.Equal(t, "3", cellAt(t, grid, 1, 0))
	assert.Equal(t, "4", cellAt(t, grid, 1, 1))
	assert.Equal(t, mediaRangeWire, accept, "IMPORTRANGE must request the range media type")
}

func TestImportSheetSpillsLikeRange(t *testing.T) {
	t.Parallel()

	// For this engine chunk IMPORTSHEET spills like IMPORTRANGE; only the
	// requested Accept media type differs (the nested-grid rendering is deferred).
	var accept engine.MediaType
	grid := importGrid(
		t,
		"=importsheet(\"u\")\n",
		echoFetcher{body: []byte("1\t2\n3\t4\n"), accept: &accept},
		engine.DefaultLimits(),
	)
	assert.Equal(t, "1", cellAt(t, grid, 0, 0))
	assert.Equal(t, "4", cellAt(t, grid, 1, 1))
	assert.Equal(t, mediaSheetWire, accept, "IMPORTSHEET must request the sheet media type")
}

func TestImportShapeMismatches(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		src  string
		body string
	}{
		"cell with two cells":   {"=importcell(\"u\")\n", "1\t2\n"},
		"cell with two rows":    {"=importcell(\"u\")\n", "1\n2\n"},
		"row with two rows":     {"=importrow(\"u\")\n", "a\nb\n"},
		"column with wide row":  {"=importcolumn(\"u\")\n", "1\n2\t3\n"},
		"range with ragged row": {"=importrange(\"u\")\n", "1\t2\n3\n"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			grid := importGrid(t, tc.src, echoFetcher{body: []byte(tc.body)}, engine.DefaultLimits())
			assert.Equal(t, "#IMPORT!", cellAt(t, grid, 0, 0))
		})
	}
}

func TestImportOversizeRejected(t *testing.T) {
	t.Parallel()

	// A tiny cell budget rejects each spilling shape as #IMPORT! (oversize).
	tight := engine.Limits{ResultCells: 2, GridDim: 20_000, ResultBytes: 64 << 10}
	cases := map[string]struct {
		src  string
		body string
	}{
		"row":    {"=importrow(\"u\")\n", "a\tb\tc\n"},
		"column": {"=importcolumn(\"u\")\n", "a\nb\nc\n"},
		"range":  {"=importrange(\"u\")\n", "1\t2\n3\t4\n"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			grid := importGrid(t, tc.src, echoFetcher{body: []byte(tc.body)}, tight)
			assert.Equal(t, "#IMPORT!", cellAt(t, grid, 0, 0))
		})
	}
}
