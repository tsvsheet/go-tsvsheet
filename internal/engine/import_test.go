package engine_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// The content-typed import media wire strings (ADR 0006 §2). Hardcoded here as
// the black-box contract each IMPORT* function must send as its Accept header.
const (
	mediaSheetWire  engine.MediaType = "application/vnd.tsvsheet+tsv"
	mediaCellWire   engine.MediaType = "application/vnd.tsvsheet.cell+tsv"
	mediaRowWire    engine.MediaType = "application/vnd.tsvsheet.row+tsv"
	mediaColumnWire engine.MediaType = "application/vnd.tsvsheet.column+tsv"
	mediaRangeWire  engine.MediaType = "application/vnd.tsvsheet.range+tsv"
)

// echoFetcher answers every fetch with body and a Content-Type that echoes the
// requested Accept, so the handshake always matches; it captures the Accept it
// was sent so a test can assert each function requests the correct media type.
type echoFetcher struct {
	accept *engine.MediaType
	body   []byte
}

func (f echoFetcher) Fetch(_ engine.ImportURL, accept engine.MediaType) (engine.FetchResult, error) {
	if f.accept != nil {
		*f.accept = accept
	}
	return engine.FetchResult{Body: f.body, ContentType: accept}, nil
}

// fixedFetcher answers with a fixed result and error, for the failure paths
// (handshake mismatch, transport error).
type fixedFetcher struct {
	err    error
	result engine.FetchResult
}

func (f fixedFetcher) Fetch(_ engine.ImportURL, _ engine.MediaType) (engine.FetchResult, error) {
	return f.result, f.err
}

// importGrid parses src and computes it with the injected Fetcher and Limits.
func importGrid(t *testing.T, src string, f engine.Fetcher, limits engine.Limits) engine.Grid {
	t.Helper()
	s, err := engine.Parse([]byte(src))
	require.NoError(t, err)
	return s.ComputeWith(engine.ComputeOptions{At: time.Now(), Fetcher: f, Limits: limits})
}

func TestHasImports(t *testing.T) {
	t.Parallel()

	assert.True(t, parse(t, "=importcell(\"https://x.example/v\")\n").HasImports())
	assert.True(t, parse(t, "=importsheet(\"https://x.example/v\")\n").HasImports())
	assert.False(t, parse(t, "=sum(A1:A2)\n").HasImports()) // a call, but not an import
	assert.False(t, parse(t, "plain\n").HasImports())       // no formula at all
}

func TestImportDisabledYieldsImportError(t *testing.T) {
	t.Parallel()

	for _, src := range []string{
		"=importcell(\"https://x.example/v\")\n",
		"=importrow(\"https://x.example/v\")\n",
		"=importcolumn(\"https://x.example/v\")\n",
		"=importrange(\"https://x.example/v\")\n",
		"=importsheet(\"https://x.example/v\")\n",
	} {
		assert.Equal(t, "#IMPORT!", cellAt(t, compute(t, src), 0, 0), "no Fetcher injected: %s must be #IMPORT!", src)
	}
}

func TestImportErrorLiteralPropagates(t *testing.T) {
	t.Parallel()

	// A cell literally holding #IMPORT! round-trips as an error value and
	// propagates through a reference (isErrorCode recognizes it).
	assert.Equal(t, "#IMPORT!", cellAt(t, compute(t, "#IMPORT!\t=A1\n"), 0, 1))
}

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

func TestImportLeadingEqualsStaysLiteral(t *testing.T) {
	t.Parallel()

	// A values-only import never compiles a cell: a leading `=` is literal text.
	grid := importGrid(t, "=importcell(\"u\")\n", echoFetcher{body: []byte("=A1\n")}, engine.DefaultLimits())
	assert.Equal(t, "=A1", cellAt(t, grid, 0, 0))
}

func TestImportHandshakeMismatch(t *testing.T) {
	t.Parallel()

	// The server declares the cell type for an IMPORTROW request: a mismatch.
	f := fixedFetcher{result: engine.FetchResult{Body: []byte("a\tb\n"), ContentType: mediaCellWire}}
	grid := importGrid(t, "=importrow(\"u\")\n", f, engine.DefaultLimits())
	assert.Equal(t, "#IMPORT!", cellAt(t, grid, 0, 0))
}

func TestImportFetchError(t *testing.T) {
	t.Parallel()

	f := fixedFetcher{err: constants.ErrReadInput}
	grid := importGrid(t, "=importcell(\"u\")\n", f, engine.DefaultLimits())
	assert.Equal(t, "#IMPORT!", cellAt(t, grid, 0, 0))
}

func TestImportTSVParseError(t *testing.T) {
	t.Parallel()

	// A single line longer than the scanner's 1 MiB token cap makes ReadTSV fail.
	body := bytes.Repeat([]byte("a"), (1<<20)+1)
	f := echoFetcher{body: body}
	grid := importGrid(t, "=importcell(\"u\")\n", f, engine.DefaultLimits())
	assert.Equal(t, "#IMPORT!", cellAt(t, grid, 0, 0))
}

func TestImportEmptyBody(t *testing.T) {
	t.Parallel()

	f := echoFetcher{body: []byte("")}
	grid := importGrid(t, "=importcell(\"u\")\n", f, engine.DefaultLimits())
	assert.Equal(t, "#IMPORT!", cellAt(t, grid, 0, 0))
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

func TestImportWrongArity(t *testing.T) {
	t.Parallel()

	for _, src := range []string{
		"=importcell()\n",
		"=importcell(\"a\", \"b\")\n",
	} {
		grid := importGrid(t, src, echoFetcher{body: []byte("1\n")}, engine.DefaultLimits())
		assert.Equal(t, "#VALUE!", cellAt(t, grid, 0, 0), "wrong arity is #VALUE!: %s", src)
	}
}

func TestImportAcceptHeaderNegotiates(t *testing.T) {
	t.Parallel()

	// The Accept helper advertises the vendor type preferred with the standard
	// tabular types at descending quality — the header frontends must send.
	assert.Equal(
		t,
		"application/vnd.tsvsheet.cell+tsv, text/tab-separated-values;q=0.9, text/csv;q=0.8",
		mediaCellWire.Accept(),
	)
}

func TestImportGenericTabularAccepted(t *testing.T) {
	t.Parallel()

	// The standard tabular types satisfy the handshake for every function, and
	// parameters/case on the Content-Type never break it.
	cases := map[string]struct {
		src         string
		contentType engine.MediaType
		body        string
		want        [][]string
	}{
		"tsv for importsheet": {
			"=importsheet(\"u\")\n", "text/tab-separated-values", "1\t2\n3\t4\n",
			[][]string{{"1", "2"}, {"3", "4"}},
		},
		"tsv with charset param": {
			"=importcell(\"u\")\n", "text/tab-separated-values; charset=utf-8", "42\n",
			[][]string{{"42"}},
		},
		"vendor with charset param": {
			"=importcell(\"u\")\n", "application/vnd.tsvsheet.cell+tsv; charset=utf-8", "42\n",
			[][]string{{"42"}},
		},
		"csv uppercase": {
			"=importrow(\"u\")\n", "TEXT/CSV", "a,b,c\n",
			[][]string{{"a", "b", "c"}},
		},
		"csv for importrange": {
			"=importrange(\"u\")\n", "text/csv", "1,2\n3,4\n",
			[][]string{{"1", "2"}, {"3", "4"}},
		},
		"csv quoted comma stays one cell": {
			"=importcell(\"u\")\n", "text/csv", "\"a, b\"\n",
			[][]string{{"a, b"}},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			f := fixedFetcher{result: engine.FetchResult{Body: []byte(tc.body), ContentType: tc.contentType}}
			grid := importGrid(t, tc.src, f, engine.DefaultLimits())
			for r, row := range tc.want {
				for c, want := range row {
					assert.Equal(t, want, cellAt(t, grid, r, c))
				}
			}
		})
	}
}

func TestImportGenericTabularRefusals(t *testing.T) {
	t.Parallel()

	// Non-tabular types stay refused; a malformed csv body and a generic body
	// failing the requested shape are #IMPORT! — never a salvage.
	cases := map[string]struct {
		src         string
		contentType engine.MediaType
		body        string
	}{
		"text/plain refused":     {"=importcell(\"u\")\n", "text/plain", "42\n"},
		"text/html refused":      {"=importsheet(\"u\")\n", "text/html; charset=utf-8", "<table></table>\n"},
		"json refused":           {"=importrange(\"u\")\n", "application/json", "[[1,2]]\n"},
		"malformed csv":          {"=importcell(\"u\")\n", "text/csv", "\"a\n"},
		"csv shape still strict": {"=importcell(\"u\")\n", "text/csv", "1,2\n"},
		"tsv shape still strict": {"=importrow(\"u\")\n", "text/tab-separated-values", "a\nb\n"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			f := fixedFetcher{result: engine.FetchResult{Body: []byte(tc.body), ContentType: tc.contentType}}
			grid := importGrid(t, tc.src, f, engine.DefaultLimits())
			assert.Equal(t, "#IMPORT!", cellAt(t, grid, 0, 0))
		})
	}
}

func TestImportCSVLeadingEqualsStaysLiteral(t *testing.T) {
	t.Parallel()

	// Values-only holds on the csv path too: a leading `=` is literal text.
	f := fixedFetcher{result: engine.FetchResult{Body: []byte("=A1\n"), ContentType: "text/csv"}}
	grid := importGrid(t, "=importcell(\"u\")\n", f, engine.DefaultLimits())
	assert.Equal(t, "=A1", cellAt(t, grid, 0, 0))
}

func TestImportErrorValuedURLPropagates(t *testing.T) {
	t.Parallel()

	// The URL argument evaluates to #DIV/0!, which propagates unchanged — the
	// fetch never happens, so the result is the argument's error, not #IMPORT!.
	grid := importGrid(t, "=importcell(1/0)\n", echoFetcher{body: []byte("1\n")}, engine.DefaultLimits())
	assert.Equal(t, "#DIV/0!", cellAt(t, grid, 0, 0))
}
