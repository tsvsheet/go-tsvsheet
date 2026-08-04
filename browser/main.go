//go:build js && wasm

// Command browser exposes the tsvsheet engine to the browser as a set of STATELESS
// functions: the caller holds the .tsvt source, and each call parses it, applies
// one immutable engine operation, and returns the result as a JSON string. There
// is no server and no filesystem — SHEET(...) / "file"!A1 resolve to #REF! — but
// every other function, including the clock functions TODAY/NOW/ISNOW, works
// against the browser's own clock. A relative IMPORT* resolves against the seed
// store the page loads with seedData (see seed.go); an absolute URL is refused,
// so a sheet arriving in a shared link can never cause a network request.
//
// go-tsvsheet's release CI builds this into the versioned tsvsheet.wasm asset
// that browser consumers (the docs playground, tsvsheet.js) pin. It is
// js/wasm-tagged, so it is invisible to the host quality gate.
package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"syscall/js"
	"time"

	tsvsheet "github.com/tsvsheet/go-tsvsheet"
)

func main() {
	obj := js.Global().Get("Object").New()
	obj.Set("compute", guarded(compute))
	obj.Set("setCell", guarded(setCell))
	obj.Set("insertRow", guarded(edit(tsvsheet.Document.InsertRow)))
	obj.Set("deleteRow", guarded(edit(tsvsheet.Document.DeleteRow)))
	obj.Set("insertCol", guarded(edit(tsvsheet.Document.InsertCol)))
	obj.Set("deleteCol", guarded(edit(tsvsheet.Document.DeleteCol)))
	obj.Set("duplicateRow", guarded(edit(tsvsheet.Document.DuplicateRow)))
	obj.Set("duplicateCol", guarded(edit(tsvsheet.Document.DuplicateCol)))
	obj.Set("fill", guarded(fill))
	obj.Set("paste", guarded(paste))
	obj.Set("pasteInto", guarded(pasteInto))
	obj.Set("references", guarded(references))
	obj.Set("explain", guarded(explain))
	obj.Set("seedData", guarded(seedData))
	// The windowed capability (021): every UI pages an over-budget document
	// the same way, so the browser gets the same census and window reads the
	// terminal pager uses.
	obj.Set("census", guarded(census))
	obj.Set("windowRows", guarded(windowRows))
	obj.Set("windowComputed", guarded(windowComputed))
	js.Global().Set("tsvsheet", obj)
	select {} // run until the page unloads
}

// guarded wraps a binding so a panic (a short-arity call's index error, an
// unexpected runtime fault) becomes an {"error": …} result instead of killing
// the Go runtime — a dead wasm engine takes the page's whole session with it.
// (A fatal runtime abort like out-of-memory is still unrecoverable; the
// engine bounds its allocations so no single call can reach one.)
func guarded(fn func(js.Value, []js.Value) any) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) (out any) {
		defer func() {
			if r := recover(); r != nil {
				b, _ := json.Marshal(map[string]string{
					"error": fmt.Sprintf("engine call failed: %v", r),
				})
				out = string(b)
			}
		}()
		return fn(this, args)
	})
}

// view is the render model returned to JS after any operation: the computed
// grid, the (possibly edited) source, the canonical source text (the one
// sanctioned serialization — comment and shebang lines preserved), static
// diagnostics, and the sheet's volatility — both the bare "is anything live"
// flag and one cadence spec per volatile(…) call, so the page can recompute at
// the soonest cadence the sheet asks for rather than a flat interval — plus the
// sheet's own view: the rows and columns its `#.` directives hide, head, or
// freeze, as sorted 1-based positions a page can render without deriving
// anything itself.
type view struct {
	Computed    [][]string            `json:"computed"`
	Source      [][]string            `json:"source"`
	Text        string                `json:"text"`
	Diagnostics []tsvsheet.Diagnostic `json:"diagnostics"`
	Hidden      axes                  `json:"hidden"`
	Headers     axes                  `json:"headers"`
	Frozen      axes                  `json:"frozen"`
	Volatile    bool                  `json:"volatile"`
	Schedules   []string              `json:"schedules"`
}

// axes is one declaration's rows and columns, each a sorted list of 1-based
// positions. Sorted because a JSON object's key order is not a contract and a
// page renders in order.
type axes struct {
	Rows []int `json:"rows"`
	Cols []int `json:"cols"`
}

// positions turns a selection into the sorted list a page can iterate.
func positions(sel tsvsheet.Selection) []int {
	out := make([]int, 0, len(sel))
	for at := range sel {
		out = append(out, at)
	}
	sort.Ints(out)
	return out
}

// computeOptions is the environment every browser compute and trace runs
// under: the tighter browser limits, the page's clock, and the seed-backed
// Fetcher, so a relative IMPORT* resolves from the store the page loaded.
func computeOptions() tsvsheet.ComputeOptions {
	return tsvsheet.ComputeOptions{
		At:      time.Now(),
		Limits:  tsvsheet.BrowserLimits(),
		Tick:    browserTick,
		Fetcher: seedFetcher{},
	}
}

// browserTick is the recompute-pass ordinal, advanced on every render so
// tick()/frame() animate under the page's periodic recompute. The wasm bridge is
// single-threaded (the JS event loop), so a package counter is safe.
var browserTick tsvsheet.Tick

// render computes a document's sheet under the tighter browser limits and its
// own clock, and gathers the read model. Text and the grids come from the same
// Document, so the view always describes exactly the text it carries.
func render(doc tsvsheet.Document) view {
	browserTick++
	sheet := doc.Sheet()
	opts := computeOptions()
	declared, viewDiags := doc.View()
	return view{
		Computed:    sheet.ComputeWith(opts),
		Source:      sheet.Source(),
		Text:        string(doc.Text()),
		Diagnostics: append(viewDiags, tsvsheet.Check(sheet)...),
		Hidden:      axes{Rows: positions(declared.HiddenRows), Cols: positions(declared.HiddenCols)},
		Headers:     axes{Rows: positions(declared.HeaderRows), Cols: positions(declared.HeaderCols)},
		Frozen:      axes{Rows: positions(declared.FreezeRows), Cols: positions(declared.FreezeCols)},
		Volatile:    sheet.IsVolatile(),
		Schedules:   sheet.VolatileSchedules(),
	}
}

// result marshals v (or a {"error": …} object on failure) to a JSON string.
func result(v any, err error) any {
	if err != nil {
		v = map[string]string{"error": err.Error()}
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// addr builds a cell address from the (row, col) JS integer arguments.
func addr(row, col js.Value) tsvsheet.Address {
	return tsvsheet.Address{Row: row.Int(), Col: col.Int()}
}

// parse is the shared first step of every function: the source is args[0],
// parsed with its line layout retained so serialization preserves comments.
// The parse is bounded by BrowserLimits (spec 018): a document over the
// browser's resident budget refuses as ErrDocTooLarge instead of
// materializing a tab-killing allocation.
func parse(args []js.Value) (tsvsheet.Document, error) {
	return tsvsheet.ParseDocumentWith([]byte(args[0].String()), tsvsheet.BrowserLimits())
}

// compute parses and renders the source (args: source).
func compute(_ js.Value, args []js.Value) any {
	doc, err := parse(args)
	if err != nil {
		return result(nil, err)
	}
	return result(render(doc), nil)
}

// setCell replaces one cell and re-renders (args: source, row, col, text).
func setCell(_ js.Value, args []js.Value) any {
	doc, err := parse(args)
	if err != nil {
		return result(nil, err)
	}
	updated, err := doc.SetCell(addr(args[1], args[2]), args[3].String(), tsvsheet.BrowserLimits())
	if err != nil {
		return result(nil, err)
	}
	return result(render(updated), nil)
}

// edit adapts an immutable structural operation into a JS function that parses,
// applies it at the given cell, and re-renders (args: source, row, col).
func edit(op func(tsvsheet.Document, tsvsheet.Address) tsvsheet.Document) func(js.Value, []js.Value) any {
	return func(_ js.Value, args []js.Value) any {
		doc, err := parse(args)
		if err != nil {
			return result(nil, err)
		}
		return result(render(op(doc, addr(args[1], args[2]))), nil)
	}
}

// fill copies one cell across a target span with fill semantics — unpinned
// references shift by each target's offset, `$`-pinned coordinates hold — and
// re-renders (args: source, fromRow, fromCol, loRow, loCol, hiRow, hiCol).
func fill(_ js.Value, args []js.Value) any {
	doc, err := parse(args)
	if err != nil {
		return result(nil, err)
	}
	to := tsvsheet.Span{From: addr(args[3], args[4]), To: addr(args[5], args[6])}
	return result(render(doc.Fill(addr(args[1], args[2]), to)), nil)
}

// paste places a clipboard block with its top-left at the target, each formula
// rebased by target−origin with fill semantics — unpinned references shift,
// `$`-pinned coordinates hold — and re-renders (args: source, atRow, atCol,
// originRow, originCol, blockText). blockText is decoded by ParseBlock: every
// line is data, so pasted text can never smuggle in a directive or comment.
func paste(_ js.Value, args []js.Value) any {
	doc, err := parse(args)
	if err != nil {
		return result(nil, err)
	}
	updated, err := doc.Paste(addr(args[1], args[2]), addr(args[3], args[4]),
		tsvsheet.ParseBlock(tsvsheet.BlockText(args[5].String())), tsvsheet.BrowserLimits())
	if err != nil {
		return result(nil, err)
	}
	return result(render(updated), nil)
}

// pasteInto places a clipboard block over the target span: an exactly
// block-divisible span TILES (each tile rebased to its own position — one
// copied row spread-pastes onto every selected row), any other span pastes
// once at its top-left (args: source, loRow, loCol, hiRow, hiCol, originRow,
// originCol, blockText).
func pasteInto(_ js.Value, args []js.Value) any {
	doc, err := parse(args)
	if err != nil {
		return result(nil, err)
	}
	target := tsvsheet.Span{From: addr(args[1], args[2]), To: addr(args[3], args[4])}
	updated, err := doc.PasteInto(target, addr(args[5], args[6]),
		tsvsheet.ParseBlock(tsvsheet.BlockText(args[7].String())), tsvsheet.BrowserLimits())
	if err != nil {
		return result(nil, err)
	}
	return result(render(updated), nil)
}

// references returns a cell's precedents and dependents (args: source, row, col).
func references(_ js.Value, args []js.Value) any {
	doc, err := parse(args)
	if err != nil {
		return result(nil, err)
	}
	at := addr(args[1], args[2])
	return result(map[string]any{
		"precedents": doc.Sheet().Precedents(at),
		"dependents": doc.Sheet().Dependents(at),
	}, nil)
}

// explain traces how a cell was produced (args: source, row, col).
func explain(_ js.Value, args []js.Value) any {
	doc, err := parse(args)
	if err != nil {
		return result(nil, err)
	}
	trace, err := tsvsheet.ExplainWith(doc.Sheet(), addr(args[1], args[2]), computeOptions())
	if err != nil {
		return result(nil, err)
	}
	return result(trace, nil)
}
