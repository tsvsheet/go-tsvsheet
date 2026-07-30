package tsvsheet_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet"
)

// TestFacadeRoundTrip exercises every public wrapper the facade adds over
// internal/engine, asserting the surface round-trips through the root package:
// a caller parses, computes, edits, inspects, and (de)serializes a sheet
// without ever naming the internal engine.
func TestFacadeRoundTrip(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	// ReadTSV -> Grid, WriteTSV -> bytes: a literal grid round-trips.
	grid, err := tsvsheet.ReadTSV(strings.NewReader("a\tb\n1\t2\n"))
	require.NoError(t, err)
	want.Equal(tsvsheet.Grid{{"a", "b"}, {"1", "2"}}, grid)

	var buf bytes.Buffer
	require.NoError(t, tsvsheet.WriteTSV(&buf, grid))
	want.Equal("a\tb\n1\t2\n", buf.String())

	// A document declares its own view, resolved against its own extent.
	viewDoc, err := tsvsheet.ParseDocument([]byte("#.header\trows(count(1))\n#.hide\tcols(range(B:B))\nn\tq\nx\t1\n"))
	require.NoError(t, err)
	view, viewDiags := viewDoc.View()
	want.Empty(viewDiags)
	want.Equal(tsvsheet.Extent{Rows: 2, Cols: 2}, viewDoc.Extent())
	want.Equal(tsvsheet.Selection{1: true}, view.HeaderRows)
	want.Equal(tsvsheet.Selection{2: true}, view.HiddenCols)

	directives, directiveDiags := viewDoc.Directives()
	want.Empty(directiveDiags)
	want.Len(directives, 2)
	want.Equal(tsvsheet.KeyHeader, directives[0].Key)

	// IsCommentLine exposes the marker family a frontend needs to map document
	// lines to grid rows: `#.` anywhere, `#!` on line 1, legacy `# `, and
	// nothing else — a hash before a TAB is data.
	want.True(tsvsheet.IsCommentLine(2, tsvsheet.SourceLine("#.hide-cols\tB-M")))
	want.True(tsvsheet.IsCommentLine(1, tsvsheet.SourceLine("#!/usr/bin/env tsvsheet")))
	want.True(tsvsheet.IsCommentLine(2, tsvsheet.SourceLine("# legacy note")))
	want.False(tsvsheet.IsCommentLine(2, tsvsheet.SourceLine("#\tnot a comment")))
	want.False(tsvsheet.IsCommentLine(2, tsvsheet.SourceLine("#N/A\t=A2")))

	// Parse -> Compute: a formula cell computes in place.
	sheet, err := tsvsheet.Parse([]byte("2\t3\t=A1 + B1\n"))
	require.NoError(t, err)
	want.Equal("5", sheet.Compute()[0][2])

	// ParseAddress resolves A1 notation into an Address.
	addr, err := tsvsheet.ParseAddress(tsvsheet.AddressText("C1"))
	require.NoError(t, err)
	want.Equal(tsvsheet.Address{Row: 0, Col: 2}, addr)

	// Explain traces the computed cell back to its formula and inputs.
	trace, err := tsvsheet.Explain(sheet, addr)
	require.NoError(t, err)
	want.Equal("5", trace.Value)
	want.NotEmpty(trace.Formula)

	// Check reports an unknown-function diagnostic (which computes to #NAME?).
	bad, err := tsvsheet.Parse([]byte("=bogus(A1)\n"))
	require.NoError(t, err)
	want.NotEmpty(tsvsheet.Check(bad))
	want.Equal(string(tsvsheet.ErrName), bad.Compute()[0][0])
}

// TestFacadeParseBlockPaste drives the clipboard surface through the facade:
// pasted text decodes with ParseBlock (every line data, CRLF tolerated) and
// lands through Document.Paste with the copy's references rebased.
func TestFacadeParseBlockPaste(t *testing.T) {
	t.Parallel()

	blk := tsvsheet.ParseBlock("=A1*2\r\n")
	assert.Equal(t, tsvsheet.Grid{{"=A1*2"}}, blk)

	doc, err := tsvsheet.ParseDocument([]byte("10\t=A1*2\n20\n"))
	require.NoError(t, err)
	pasted, err := doc.Paste(
		tsvsheet.Address{Row: 1, Col: 1}, tsvsheet.Address{Row: 0, Col: 1},
		blk, tsvsheet.DefaultLimits(),
	)
	require.NoError(t, err)
	assert.Equal(t, "10\t=A1*2\n20\t=A2 * 2\n", string(pasted.Text()))
}

// TestFacadeLimits verifies the two injected ceilings are distinct: the browser
// build is tighter than the default so untrusted sheets are bounded harder.
func TestFacadeLimits(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	def := tsvsheet.DefaultLimits()
	browser := tsvsheet.BrowserLimits()

	want.NotEqual(def, browser)
	want.Less(browser.ResultCells, def.ResultCells)
	want.Less(browser.GridDim, def.GridDim)
}

// tracingFetcher answers one source and refuses every other, reporting the URL
// it "resolved" to — the shape a real frontend fetcher has, where a relative
// source is resolved against an operator-supplied base the engine never sees.
type tracingFetcher struct{}

func (tracingFetcher) Fetch(url tsvsheet.ImportURL, _ tsvsheet.MediaType) (tsvsheet.FetchResult, error) {
	if url != "balances.tsv" {
		return tsvsheet.FetchResult{}, errors.New("import reference escapes the data base")
	}
	return tsvsheet.FetchResult{
		ContentType: "text/tab-separated-values",
		URL:         "https://data.example.com/team/balances.tsv",
		Body:        []byte("310000\n"),
	}, nil
}

func explainCell(t *testing.T, src string) tsvsheet.Trace {
	t.Helper()

	sheet, err := tsvsheet.Parse([]byte(src))
	require.NoError(t, err)
	at, err := tsvsheet.ParseAddress("A1")
	require.NoError(t, err)
	trace, err := tsvsheet.ExplainWith(sheet, at, tsvsheet.ComputeOptions{Fetcher: tracingFetcher{}})
	require.NoError(t, err)
	return trace
}

func TestExplainWith_ReportsTheResolvedImportURL(t *testing.T) {
	t.Parallel()

	trace := explainCell(t, "=importcell(\"balances.tsv\")\n")

	assert.Equal(t, "310000", trace.Value, "the import resolves, where plain Explain would be #IMPORT!")
	require.Len(t, trace.Imports, 1)
	assert.Equal(t, "balances.tsv", trace.Imports[0].Source, "the source exactly as the sheet wrote it")
	assert.Equal(t, "https://data.example.com/team/balances.tsv", trace.Imports[0].URL, "where it actually went")
	assert.Empty(t, trace.Imports[0].Error)
}

func TestExplainWith_ReportsWhyAnImportFailed(t *testing.T) {
	t.Parallel()

	trace := explainCell(t, "=importcell(\"../../admin/keys.tsv\")\n")

	assert.Equal(t, "#IMPORT!", trace.Value, "the grid stays opaque")
	require.Len(t, trace.Imports, 1)
	assert.Equal(t, "import reference escapes the data base", trace.Imports[0].Error,
		"the trace is the only place the specific failure is visible")
	assert.Empty(t, trace.Imports[0].URL)
}

func TestExplainWith_ReportsEveryImportInTheFormula(t *testing.T) {
	t.Parallel()

	trace := explainCell(t, "=iferror(importcell(\"balances.tsv\"), importcell(\"missing.tsv\"))\n")

	require.Len(t, trace.Imports, 2, "both calls are described, not just the one that was evaluated")
	assert.Equal(t, "balances.tsv", trace.Imports[0].Source)
	assert.Equal(t, "missing.tsv", trace.Imports[1].Source)
	assert.NotEmpty(t, trace.Imports[1].Error)
}

func TestExplainWith_NoImportsInTheFormula(t *testing.T) {
	t.Parallel()

	trace := explainCell(t, "=1+1\n")

	assert.Equal(t, "2", trace.Value)
	assert.Empty(t, trace.Imports)
}

func TestExplainWith_WrongArityIsNotTraced(t *testing.T) {
	t.Parallel()

	// A malformed IMPORT* is #VALUE! from the evaluator; there is no single
	// source to report, so the trace carries no import note.
	trace := explainCell(t, "=importcell(\"a.tsv\", \"b.tsv\")\n")

	assert.Empty(t, trace.Imports)
}

func TestExplain_WithoutAFetcherTracesNoImports(t *testing.T) {
	t.Parallel()

	sheet, err := tsvsheet.Parse([]byte("=importcell(\"balances.tsv\")\n"))
	require.NoError(t, err)
	at, err := tsvsheet.ParseAddress("A1")
	require.NoError(t, err)

	trace, err := tsvsheet.Explain(sheet, at)
	require.NoError(t, err)
	assert.Equal(t, "#IMPORT!", trace.Value)
	assert.Empty(t, trace.Imports, "no fetcher means nothing to report")
}

func TestExplainWith_MissingCell(t *testing.T) {
	t.Parallel()

	sheet, err := tsvsheet.Parse([]byte("1\n"))
	require.NoError(t, err)
	at, err := tsvsheet.ParseAddress("Z99")
	require.NoError(t, err)

	_, err = tsvsheet.ExplainWith(sheet, at, tsvsheet.ComputeOptions{})
	require.Error(t, err)
}
