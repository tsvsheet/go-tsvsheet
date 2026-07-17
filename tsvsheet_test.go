package tsvsheet_test

import (
	"bytes"
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
