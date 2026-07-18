package tsvsheet_test

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tsvsheet "github.com/tsvsheet/go-tsvsheet"
)

// exprGrid is the differential corpus grid: A1=2 (number), B1=x (string),
// C1=3.5 (number); A2 empty, B2=7 (number), C2=#N/A (error literal).
func exprGrid() tsvsheet.Grid {
	return tsvsheet.Grid{
		{"2", "x", "3.5"},
		{"", "7", "#N/A"},
	}
}

// exprClock is the fixed pass clock every expression test injects, so volatile
// functions are deterministic and comparable across both evaluation routes.
func exprClock() time.Time {
	return time.Date(2026, 7, 18, 12, 30, 0, 0, time.UTC)
}

// sheetHolding renders grid g as .tsvt source — via the engine's own WriteTSV
// serialization — with `=e` appended as the sole cell of a new bottom row: the
// sheet the differential compares the seam against.
func sheetHolding(t *testing.T, g tsvsheet.Grid, e string) []byte {
	t.Helper()
	var b strings.Builder
	require.NoError(t, tsvsheet.WriteTSV(&b, g))
	return []byte(b.String() + "=" + e + "\n")
}

// computedCell computes the sheet holding `=e` below grid g with opts and
// returns the formula cell's rendered text — what WriteTSV would emit for it.
func computedCell(t *testing.T, g tsvsheet.Grid, e string, opts tsvsheet.ComputeOptions) string {
	t.Helper()
	sheet, err := tsvsheet.Parse(sheetHolding(t, g, e))
	require.NoError(t, err)
	out := sheet.ComputeWith(opts)
	require.Greater(t, len(out), len(g))
	return out[len(g)][0]
}

// assertDifferential asserts the two evaluation routes agree AND that both
// produce the contract's expected canonical text — so the test verifies the
// specified value, not merely that the routes share a (possibly wrong) output.
func assertDifferential(t *testing.T, g tsvsheet.Grid, e, want string, opts tsvsheet.ComputeOptions) {
	t.Helper()
	expr, err := tsvsheet.CompileExpr([]byte(e))
	require.NoError(t, err)
	seam := tsvsheet.FormatValue(expr.Eval(g, opts))
	cell := computedCell(t, g, e, opts)
	assert.Equal(t, want, seam, "seam value for %q", e)
	assert.Equal(t, cell, seam, "seam and sheet diverge for %q", e)
}

// TestExprDifferential runs the acceptance corpus: for expressions spanning
// every operand kind, every operator, error values, and volatile functions,
// CompileExpr(e).Eval(g, opts) formats byte-identically to the value a sheet
// holding `=e` over the same grid computes — and both match the contract's
// expected text.
func TestExprDifferential(t *testing.T) {
	t.Parallel()
	corpus := []struct{ expr, want string }{
		// Literal operands: number, string, boolean, error.
		{`1.5`, "1.5"},
		{`"hi"`, "hi"},
		{`TRUE`, "TRUE"},
		{`#DIV/0!`, "#DIV/0!"},
		// References: number, string, empty, error, out-of-grid, bare range.
		{`A1`, "2"},
		{`B1`, "x"},
		{`A2`, ""},
		{`C2`, "#N/A"},
		{`Z99`, "#REF!"},
		{`A0`, "#REF!"},
		{`A1:C2`, "#VALUE!"},
		// Unary, percent, power, arithmetic, concatenation.
		{`-A1`, "-2"},
		{`+C1`, "3.5"},
		{`50%`, "0.5"},
		{`A1%`, "0.02"},
		{`2^3`, "8"},
		{`10/4`, "2.5"},
		{`1/0`, "#DIV/0!"},
		{`A1*C1`, "7"},
		{`A1+C1`, "5.5"},
		{`C1-A1`, "1.5"},
		{`"a"&"b"`, "ab"},
		{`A1&B1`, "2x"},
		// Comparisons, both outcomes each.
		{`A1=2`, "TRUE"},
		{`A1<>2`, "FALSE"},
		{`A1<3`, "TRUE"},
		{`A1<=1`, "FALSE"},
		{`A1>1`, "TRUE"},
		{`A1>=3`, "FALSE"},
		// Calls, pipe sugar, ranges as arguments, conditionals.
		{`sum(A1,C1)`, "5.5"},
		{`sum(A1:A2)`, "2"},
		{`A1 | sum()`, "2"},
		{`if(A1>1,"big","small")`, "big"},
		// Error-value production and propagation.
		{`"x"+1`, "#VALUE!"},
		{`C2+1`, "#N/A"},
		{`bogus(1)`, "#NAME?"},
		// Volatile functions read the injected clock.
		{`today()`, "2026-07-18"},
		{`now()`, "2026-07-18 12:30:00"},
		// Dynamic arrays reduce to their scalar-context (top-left) value.
		{`sequence(2,2)`, "1"},
		{`transpose(A1:C1)`, "2"},
		// Gating: a zero Loader disables SHEET, a nil Fetcher disables IMPORT*.
		{`sheet("lib.tsvt")`, "#REF!"},
		{`importcell("data:cell")`, "#IMPORT!"},
	}
	for _, tc := range corpus {
		t.Run(tc.expr, func(t *testing.T) {
			t.Parallel()
			assertDifferential(t, exprGrid(), tc.expr, tc.want, tsvsheet.ComputeOptions{At: exprClock()})
		})
	}
}

// stubLoader resolves every reference to the fixed doubling sub-sheet
// `=output(input(1)*2)`, mirroring what a Compute pass would load.
func stubLoader(t *testing.T) tsvsheet.Loader {
	t.Helper()
	sub, err := tsvsheet.Parse([]byte("=output(input(1)*2)\n"))
	require.NoError(t, err)
	return func(_, ref tsvsheet.Path) (tsvsheet.Sheet, tsvsheet.Path, error) {
		return sub, ref, nil
	}
}

// echoFetcher answers every import with the requested media type and a fixed
// body, so the handshake succeeds identically on both evaluation routes.
type echoFetcher struct {
	body string
}

// Fetch implements tsvsheet.Fetcher.
func (f echoFetcher) Fetch(_ tsvsheet.ImportURL, accept tsvsheet.MediaType) (tsvsheet.FetchResult, error) {
	return tsvsheet.FetchResult{ContentType: accept, Body: []byte(f.body)}, nil
}

// TestExprDifferentialInjected verifies the seam honors the injected
// collaborators exactly as Compute does: an enabled Loader embeds SHEET, an
// enabled Fetcher serves IMPORT*, and tightened Limits bound an array result.
func TestExprDifferentialInjected(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, expr, want string
		opts             tsvsheet.ComputeOptions
	}{
		{
			name: "loader embeds SHEET",
			expr: `sheet("lib.tsvt", 21)`,
			want: "42",
			opts: tsvsheet.ComputeOptions{At: exprClock(), Loader: stubLoader(t), Base: "main.tsvt"},
		},
		{
			name: "fetcher serves IMPORTCELL",
			expr: `importcell("https://example.test/cell")`,
			want: "42",
			opts: tsvsheet.ComputeOptions{At: exprClock(), Fetcher: echoFetcher{body: "42"}},
		},
		{
			name: "limits bound an array result",
			expr: `sequence(2,2)`,
			want: "#VALUE!",
			opts: tsvsheet.ComputeOptions{
				At:     exprClock(),
				Limits: tsvsheet.Limits{ResultCells: 2, GridDim: 100, ResultBytes: 100},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertDifferential(t, exprGrid(), tc.expr, tc.want, tc.opts)
		})
	}
}

// TestCompileExprSyntaxError asserts every malformed input is the ErrSyntax
// sentinel, matchable with errors.Is, and that a parser-reported failure
// carries its line and column detail.
func TestCompileExprSyntaxError(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	for _, malformed := range []string{"sum(", ")", "1 +", `"unterminated`, "1 2"} {
		_, err := tsvsheet.CompileExpr([]byte(malformed))
		require.Error(t, err, "input %q", malformed)
		want.True(errors.Is(err, tsvsheet.ErrSyntax), "input %q: %v", malformed, err)
	}

	_, err := tsvsheet.CompileExpr([]byte("sum("))
	require.ErrorIs(t, err, tsvsheet.ErrSyntax)
	want.Contains(err.Error(), "line")
	want.Contains(err.Error(), "column")
}

// TestExprCompileOnceEvalMany verifies R3: one compiled expression evaluates
// against many different grids — concurrently, from a single immutable value —
// with each evaluation resolving references into its own grid.
func TestExprCompileOnceEvalMany(t *testing.T) {
	t.Parallel()

	expr, err := tsvsheet.CompileExpr([]byte("A1*2"))
	require.NoError(t, err)

	var wg sync.WaitGroup
	results := make([]string, 8)
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g := tsvsheet.Grid{{strconv.Itoa(i + 1)}}
			results[i] = tsvsheet.FormatValue(expr.Eval(g, tsvsheet.ComputeOptions{At: exprClock()}))
		}()
	}
	wg.Wait()
	for i, got := range results {
		assert.Equal(t, strconv.Itoa((i+1)*2), got)
	}
}

// TestFormatValueScalarContext verifies FormatValue's scalar-context reduction
// directly at the facade: a 2-D array formats as its top-left value, and an
// empty grid evaluates literals with no grid at all.
func TestFormatValueScalarContext(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	expr, err := tsvsheet.CompileExpr([]byte("sequence(3,2)"))
	require.NoError(t, err)
	want.Equal("1", tsvsheet.FormatValue(expr.Eval(nil, tsvsheet.ComputeOptions{At: exprClock()})))

	lit, err := tsvsheet.CompileExpr([]byte(`"free-standing"`))
	require.NoError(t, err)
	want.Equal("free-standing", tsvsheet.FormatValue(lit.Eval(nil, tsvsheet.ComputeOptions{})))
}
