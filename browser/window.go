//go:build js && wasm

// The windowed capability at the browser boundary (spec 021): a document past
// the resident budget is view/compute-only in EVERY frontend, paged the same
// way, so the browser reads viewport windows exactly as the terminal pager
// does (R18).
//
// One honest bound, stated because the mechanism cannot match the terminal's:
// the source arrives here as a JS string, so the document is already resident
// before these functions see it — backing a ReadAt with Blob.slice would need
// asynchronous reads, and the engine's source seam is synchronous. What
// windowing saves in a browser is the materialization and the render, which
// is what actually ends a tab; it does not save the read.
package main

import (
	"bytes"
	"syscall/js"

	tsvsheet "github.com/tsvsheet/go-tsvsheet"
)

// sourceOf wraps args[0] as an any-size byte source.
func sourceOf(args []js.Value) tsvsheet.ByteSource {
	src := []byte(args[0].String())
	return tsvsheet.ByteSource{ReadAt: bytes.NewReader(src), Size: int64(len(src))}
}

// censusView is what a frontend needs to decide and to describe: the totals,
// and whether this document is over the budget that makes it view-only.
type censusView struct {
	Rows       int   `json:"rows"`
	Cols       int   `json:"cols"`
	Cells      int64 `json:"cells"`
	Formulas   int64 `json:"formulas"`
	Budget     int64 `json:"budget"`
	IsWindowed bool  `json:"isWindowed"`
}

// census reports a source's totals from one scan, materializing nothing
// (args: source). It is the pre-flight a frontend runs to choose between
// editing and paging — the same census, and the same ceiling, the terminal
// consults.
func census(_ js.Value, args []js.Value) any {
	limits := tsvsheet.BrowserLimits()
	counted, err := tsvsheet.Census(sourceOf(args))
	if err != nil {
		return result(nil, err)
	}
	return result(censusView{
		Rows:       counted.Rows,
		Cols:       counted.MaxWidth,
		Cells:      counted.Cells,
		Formulas:   counted.Formulas,
		Budget:     limits.EffectiveResidentCells(),
		IsWindowed: counted.Cells > limits.EffectiveResidentCells(),
	}, nil)
}

// windowRows returns the source texts of one viewport (args: source, from, n)
// — literals and formulas as written, nothing computed, clipped to the grid.
func windowRows(_ js.Value, args []js.Value) any {
	windowed, err := openWindowed(args)
	if err != nil {
		return result(nil, err)
	}
	rows, err := windowed.Rows(args[1].Int(), args[2].Int())
	if err != nil {
		return result(nil, err)
	}
	return result(rows, nil)
}

// windowComputed evaluates one viewport (args: source, from, n): the window's
// own formulas computed, their dependencies read lazily, all of it under the
// browser's budgets.
func windowComputed(_ js.Value, args []js.Value) any {
	windowed, err := openWindowed(args)
	if err != nil {
		return result(nil, err)
	}
	browserTick++
	grid, err := windowed.ComputeRows(args[1].Int(), args[2].Int(), computeOptions())
	if err != nil {
		return result(nil, err)
	}
	return result(grid, nil)
}

// openWindowed opens the source as a windowed document. A document within the
// budget has no windowed capability — it is meant to be edited, not paged — so
// asking for a window on one is a caller error rather than a silent empty
// answer.
func openWindowed(args []js.Value) (*tsvsheet.WindowedSheet, error) {
	_, windowed, err := tsvsheet.OpenSheet(sourceOf(args), tsvsheet.BrowserLimits())
	if err != nil {
		return nil, err
	}
	if windowed == nil {
		return nil, tsvsheet.ErrInvalidValue.With(nil, "message", "document is within the budget: read it whole")
	}
	return windowed, nil
}
