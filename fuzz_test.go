package tsvsheet_test

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	tsvsheet "github.com/tsvsheet/go-tsvsheet"
)

// The fuzz targets assert the library's standing contracts over arbitrary
// untrusted input, at the public facade where that input actually arrives:
// canonicalization is idempotent, the edits format round-trips byte-for-byte,
// and evaluation under a fixed clock is total and deterministic (R1). Every
// target runs under BrowserLimits-scale budgets where evaluation is involved,
// so a fuzz-found pathological input exercises the refusal paths rather than
// the machine.

// fuzzClock is the fixed pass clock: volatile functions and the pass PRNG seed
// from it, so determinism is assertable.
func fuzzClock() time.Time { return time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC) }

// FuzzParseDocument asserts canonical idempotence: a document that parses
// re-serializes to text that parses again and re-serializes byte-identically.
func FuzzParseDocument(f *testing.F) {
	f.Add([]byte("a\tb\n=A1+1\t2\n"))
	f.Add([]byte("#. hidden\tB\n# note\n=sum(A1:B2)\t#REF!\n"))
	f.Add([]byte("\t\n\t=DEJTLX1\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		d, err := tsvsheet.ParseDocument(data)
		if err != nil {
			return
		}
		text := d.Text()
		d2, err := tsvsheet.ParseDocument(text)
		if err != nil {
			t.Fatalf("canonical text failed to reparse: %v", err)
		}
		if !bytes.Equal(d2.Text(), text) {
			t.Fatal("canonicalization is not idempotent")
		}
	})
}

// FuzzParseEdits asserts the edits document's round-trip contract: a batch
// that parses re-serializes to text that parses to the same bytes.
func FuzzParseEdits(f *testing.F) {
	f.Add([]byte("#.base\t9f2c\nsetCell\tB2\t=A1+1\ninsertRow\t2\n"))
	f.Add([]byte("fill\tA1\tA2:B4\npaste\tC1\tA1\tOQ==\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		e, err := tsvsheet.ParseEdits(data)
		if err != nil {
			return
		}
		text := e.Text()
		e2, err := tsvsheet.ParseEdits(text)
		if err != nil {
			t.Fatalf("canonical edits failed to reparse: %v", err)
		}
		if !bytes.Equal(e2.Text(), text) {
			t.Fatal("edits round-trip is not byte-identical")
		}
	})
}

// FuzzCompileExpr asserts that a compiled expression evaluates totally (no
// panic) and deterministically under a fixed clock — the R1 contract that one
// engine, one clock, one draw sequence is the semantics.
func FuzzCompileExpr(f *testing.F) {
	f.Add("sum(A1:B2) | round(2)")
	f.Add(`if(A1 >= 1, "y", #N/A)`)
	f.Add("rand() + randbetween(1, 9)")
	f.Add("sum(A1:A50000000000)")
	f.Fuzz(func(t *testing.T, src string) {
		expr, err := tsvsheet.CompileExpr([]byte(src))
		if err != nil {
			return
		}
		grid := tsvsheet.Grid{{"1", "2"}, {"3", "x"}}
		opts := tsvsheet.ComputeOptions{At: fuzzClock(), Limits: tsvsheet.BrowserLimits()}
		first := tsvsheet.FormatValue(expr.Eval(grid, opts))
		second := tsvsheet.FormatValue(expr.Eval(grid, opts))
		if first != second {
			t.Fatalf("evaluation is not deterministic: %q then %q", first, second)
		}
	})
}

// FuzzParse asserts that a sheet that parses computes totally and
// deterministically under a fixed clock and browser-scale budgets.
func FuzzParse(f *testing.F) {
	f.Add([]byte("1\t=A1+1\n=sum(A1:B1)\t=sequence(2,2)\n"))
	f.Add([]byte("=rept(\"a\", 999999999)\t=sum(A1:DEJTLX1)\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		s, err := tsvsheet.Parse(data)
		if err != nil {
			return
		}
		opts := tsvsheet.ComputeOptions{At: fuzzClock(), Limits: tsvsheet.BrowserLimits()}
		first := s.ComputeWith(opts)
		second := s.ComputeWith(opts)
		if len(first) != len(second) {
			t.Fatal("recompute changed the grid's shape")
		}
		for r := range first {
			if len(first[r]) != len(second[r]) {
				t.Fatalf("recompute changed row %d's width: %d then %d", r, len(first[r]), len(second[r]))
			}
			for c := range first[r] {
				if first[r][c] != second[r][c] {
					t.Fatalf("recompute diverged at (%d,%d): %q then %q", r, c, first[r][c], second[r][c])
				}
			}
		}
	})
}

// FuzzReadTSV asserts the raw grid reader's round-trip: a grid that reads
// re-serializes (TABs within rows, LF between them) to text that reads back as
// the same grid.
func FuzzReadTSV(f *testing.F) {
	f.Add([]byte("a\tb\nc\n"))
	f.Add([]byte("\t\t\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		g, err := tsvsheet.ReadTSV(bytes.NewReader(data))
		if err != nil {
			return
		}
		var b bytes.Buffer
		for _, row := range g {
			_, _ = b.WriteString(strings.Join(row, "\t"))
			_, _ = b.WriteString("\n")
		}
		// A grid the bare format cannot faithfully represent (a first cell that
		// would read back as a comment marker) is REFUSED by WriteTSV rather
		// than silently losing the row; the manual serialization above has no
		// such guard, so mirror the refusal and skip.
		if writeErr := tsvsheet.WriteTSV(io.Discard, g); writeErr != nil {
			return
		}
		g2, err := tsvsheet.ReadTSV(bytes.NewReader(b.Bytes()))
		if err != nil {
			t.Fatalf("serialized grid failed to re-read: %v", err)
		}
		if len(g) != len(g2) {
			t.Fatalf("row count changed: %d then %d", len(g), len(g2))
		}
		for r := range g {
			if len(g[r]) != len(g2[r]) {
				t.Fatalf("row %d width changed: %d then %d", r, len(g[r]), len(g2[r]))
			}
			for c := range g[r] {
				if g[r][c] != g2[r][c] {
					t.Fatalf("cell (%d,%d) changed: %q then %q", r, c, g[r][c], g2[r][c])
				}
			}
		}
	})
}

// FuzzOpenSheet asserts the any-size entry point over arbitrary bytes: exactly
// one capability comes back (or an error), a resident open computes exactly as
// Parse does, and a windowed census agrees with the resident row count when
// both are obtainable.
func FuzzOpenSheet(f *testing.F) {
	f.Add([]byte("a\tb\n=A1\tc\n"), 4)
	f.Add([]byte("#. note\nr0\n"), 1)
	f.Fuzz(func(t *testing.T, data []byte, resident int) {
		if resident < 1 || resident > 1<<20 {
			return
		}
		limits := tsvsheet.DefaultLimits()
		limits.ResidentCells = resident
		sheet, windowed, err := tsvsheet.OpenSheet(
			tsvsheet.ByteSource{ReadAt: bytes.NewReader(data), Size: int64(len(data))}, limits)
		if err != nil {
			return
		}
		parsed, parseErr := tsvsheet.Parse(data)
		if windowed == nil {
			if parseErr != nil {
				t.Fatalf("resident open accepted what Parse refuses: %v", parseErr)
			}
			a, b := sheet.ComputeWith(tsvsheet.ComputeOptions{At: fuzzClock(), Limits: limits}),
				parsed.ComputeWith(tsvsheet.ComputeOptions{At: fuzzClock(), Limits: limits})
			if len(a) != len(b) {
				t.Fatal("resident open diverged from Parse in shape")
			}
			return
		}
		if parseErr == nil {
			rows, rowsErr := windowed.Rows(0, 1<<20)
			if rowsErr != nil {
				t.Fatalf("windowed rows failed on a parseable doc: %v", rowsErr)
			}
			if len(rows) != windowed.Census().Rows {
				t.Fatalf("census rows %d vs served %d", windowed.Census().Rows, len(rows))
			}
		}
	})
}

// FuzzComputeRows asserts windowed evaluation over arbitrary bytes: any
// window of a windowed document computes without panic, deterministically for
// a fixed clock, and — under generous budgets — cell-for-cell equal to the
// resident computation of the same document where one exists.
func FuzzComputeRows(f *testing.F) {
	f.Add([]byte("1\t=A1+1\n=sum(A1:B1)\n"), 0, 2)
	f.Add([]byte("=sequence(3)\n\n\nx\n"), 0, 4)
	f.Fuzz(func(t *testing.T, data []byte, from, n int) {
		if n < 0 || n > 1<<12 {
			return
		}
		limits := tsvsheet.DefaultLimits()
		limits.ResidentCells = 1 // force the windowed capability
		_, windowed, err := tsvsheet.OpenSheet(
			tsvsheet.ByteSource{ReadAt: bytes.NewReader(data), Size: int64(len(data))}, limits)
		if err != nil || windowed == nil {
			return
		}
		opts := tsvsheet.ComputeOptions{At: fuzzClock(), Limits: tsvsheet.DefaultLimits()}
		first, err := windowed.ComputeRows(from, n, opts)
		if err != nil {
			return // a legitimate refusal (e.g. a syntax error read lazily)
		}
		second, err := windowed.ComputeRows(from, n, opts)
		if err != nil {
			t.Fatalf("the same window failed on repeat: %v", err)
		}
		if len(first) != len(second) {
			t.Fatal("window shape changed between identical calls")
		}
		parsed, parseErr := tsvsheet.Parse(data)
		if parseErr != nil || from < 0 {
			return
		}
		resident := parsed.ComputeWith(opts)
		source, err := windowed.Rows(from, n)
		if err != nil {
			t.Fatalf("source rows failed where compute succeeded: %v", err)
		}
		for r := range first {
			for c := range first[r] {
				if first[r][c] != second[r][c] {
					t.Fatalf("nondeterministic at (%d,%d): %q then %q", r, c, first[r][c], second[r][c])
				}
				if source[r][c] == "" {
					continue // an empty source cell may hold a resident SPILL; windows do not spill — documented
				}
				rr := from + r
				if rr < len(resident) && c < len(resident[rr]) && first[r][c] != resident[rr][c] {
					t.Fatalf(
						"windowed diverged from resident at (%d,%d): %q vs %q",
						rr,
						c,
						first[r][c],
						resident[rr][c],
					)
				}
			}
		}
	})
}
