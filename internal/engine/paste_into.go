// Pasting into a selection.
//
// paste.go places a block once, at a point. This file answers the other
// question a spreadsheet asks of a clipboard: what happens when the target is a
// SELECTION rather than a cell. The block tiles the span when it divides it
// exactly, lands once at the top-left when it does not, and the span itself is
// bounded — the work and the allocation are O(area) while the request is four
// integers, so an unbounded selection would be an out-of-memory the caller has
// no way to catch.
package engine

import "github.com/tsvsheet/go-tsvsheet/internal/constants"

// PasteInto returns a new sheet with block placed over the normalized target
// span. When the span's rows AND columns are exact multiples of the block's
// dimensions, the block TILES the span — each tile rebased by its own delta
// (tile position − origin), exactly as separate pastes would rebase — which
// is how one copied row spread-pastes onto every selected row. Any other
// span places the block once at the span's top-left (Paste semantics), so a
// selection that does not fit the block never half-fills. Atomic: an error
// in any tile leaves the sheet unchanged.
func (s Sheet) PasteInto(target Span, origin Address, block Grid, limits Limits) (Sheet, error) {
	span := normalized(target)
	if !tilesSpan(span, rowIndex(len(block)), colIndex(widestBlockRow(block))) {
		return s.Paste(span.From, origin, block, limits)
	}
	return s.pasteTiles(span, origin, block, limits)
}

// pasteTiles fills an exactly-divisible span tile by tile over ONE copy of
// the grid — a thousand-cell selection is one allocation, not a thousand.
// Atomicity note: every per-cell refusal (a syntax error, the comment-marker
// guard) depends only on the block's text and the target's column class, and
// the leftmost/topmost tile is always visited first — so with today's
// refusal set a failure can only fire in the FIRST tile that can fail at
// all; the sheet copy still makes the later-tile guarantee structural.
func (s Sheet) pasteTiles(span Span, origin Address, block Grid, limits Limits) (Sheet, error) {
	if err := spanBounds(span, origin, limits); err != nil {
		return Sheet{}, err
	}
	width := colIndex(widestBlockRow(block))
	cells := mapRows(s.grid(), cloneRow)
	for r := span.From.Row; r <= span.To.Row; r += len(block) {
		for c := span.From.Col; c <= span.To.Col; c += int(width) {
			var err error
			if cells, err = pasteBlockInto(cells, Address{Row: r, Col: c}, origin, block, width); err != nil {
				return Sheet{}, err
			}
		}
	}
	return Sheet{cells: cells}, nil
}

// tilesSpan reports whether a normalized span is an exact grid of
// rows×cols-sized tiles. An empty block never tiles (Paste no-ops it).
func tilesSpan(span Span, rows rowIndex, cols colIndex) bool {
	if rows == 0 || cols == 0 {
		return false
	}
	spanRows := rowIndex(span.To.Row - span.From.Row + 1)
	spanCols := colIndex(span.To.Col - span.From.Col + 1)
	return spanRows%rows == 0 && spanCols%cols == 0
}

// spanBounds rejects a tiled paste whose span has negative corners, reaches
// the grid limit on either axis, or whose AREA exceeds the result-cells
// ceiling — the work and the allocation are O(area) while the request is four
// integers, so an unbounded area is an out-of-memory the caller cannot catch
// (the wasm runtime aborts fatally, killing the engine for the page's life).
func spanBounds(span Span, origin Address, limits Limits) error {
	if span.From.Row < 0 || span.From.Col < 0 {
		return constants.ErrInvalidValue.With(nil, "address", span.From.String())
	}
	if origin.Row < 0 || origin.Col < 0 {
		return constants.ErrInvalidValue.With(nil, "origin", origin.String())
	}
	if span.To.Row >= limits.GridDim || span.To.Col >= limits.GridDim {
		return constants.ErrInvalidValue.With(nil, "paste exceeds the grid limit", span.To.String())
	}
	// Dimension-first via tooManyCells, never the raw product: two large
	// dimensions multiply past the int ceiling and wrap negative, passing the
	// comparison — the exact trap overCellBudget's comment documents.
	rows := resultDim(span.To.Row - span.From.Row + 1)
	cols := resultDim(span.To.Col - span.From.Col + 1)
	if limits.tooManyCells(rows, cols) {
		return constants.ErrInvalidValue.With(nil, "paste span exceeds the cell limit", span.To.String())
	}
	return nil
}
