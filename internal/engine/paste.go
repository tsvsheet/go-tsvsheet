// Paste and the clipboard block: the paste half of the copy-time reference
// model (spec 010). A copied block travels as plain TSV text — ParseBlock is
// the one definition of how that text becomes a grid again — and Paste places
// it with every formula rebased by the single delta target−origin, using the
// same AST transform Fill applies per target (fill.go). Like the rest of the
// family, everything re-serializes through the canonical renderer.
package engine

import (
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
)

// BlockText is clipboard-block source text: TAB-separated cells on
// newline-separated rows, as a copied range travels through a clipboard.
type BlockText string

// ParseBlock reads a clipboard block: CRLF and lone CR normalize to LF,
// exactly one trailing newline is ignored, rows split on LF and cells on TAB.
// Every line is data — a clipboard block has no comment or directive
// semantics, unlike a .tsvt file. An empty text is a single empty cell (the
// TSV serialization of one empty cell IS the empty string), so pasting it
// clears its target.
func ParseBlock(text BlockText) Grid {
	normalized := strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(string(text))
	trimmed := strings.TrimSuffix(normalized, "\n")
	rows := strings.Split(trimmed, "\n")
	out := make(Grid, len(rows))
	for r, row := range rows {
		out[r] = strings.Split(row, tab)
	}
	return out
}

// Paste returns a new sheet with block — a grid of raw cell texts, as copied
// with its top-left at origin — placed with its top-left at at: every formula
// rebases by the single delta at−origin with Fill's semantics (unpinned
// coordinates shift, `$`-pinned hold, a coordinate rebased off the grid
// renders #REF!, cross-sheet references copy unshifted); a literal copies
// verbatim; an empty block cell clears its target, and a ragged block pads its
// short rows with empty cells, so a paste always overwrites its whole
// rectangular footprint. The grid grows to fit, bounded by limits as Set is. Paste
// is atomic: a malformed formula is a syntax error naming its target cell and
// the sheet is unchanged. An empty block is a no-op.
func (s Sheet) Paste(at, origin Address, block Grid, limits Limits) (Sheet, error) {
	if err := pasteBounds(at, origin, block, limits); err != nil {
		return Sheet{}, err
	}
	d := offset{rows: at.Row - origin.Row, cols: at.Col - origin.Col}
	width := colIndex(widestBlockRow(block))
	cells := mapRows(s.cells, cloneRow)
	for r, row := range block {
		var err error
		if cells, err = pasteRow(cells, Address{Row: at.Row + r, Col: at.Col}, padded(row, width), d); err != nil {
			return Sheet{}, err
		}
	}
	return Sheet{cells: cells}, nil
}

// padded extends a short block row to the block's full width with empty
// cells, so a ragged block still overwrites its whole rectangular footprint —
// the footprint pasteBounds already reserves. Stale data never survives under
// a pasted rectangle.
func padded(row []string, width colIndex) []string {
	if len(row) == int(width) {
		return row
	}
	out := make([]string, width)
	copy(out, row)
	return out
}

// pasteRow places one block row from its leftmost target, parsing each text at
// its target position and rebasing it by the block's delta.
func pasteRow(cells [][]cell, from Address, row []string, d offset) ([][]cell, error) {
	for c, text := range row {
		target := Address{Row: from.Row, Col: from.Col + c}
		if WouldStartCommentLine(target, CellText(text)) {
			return nil, constants.ErrCommentCell.With(nil, "cell", target.String())
		}
		parsed, err := parseCell(textVal(text), rowIndex(target.Row), colIndex(target.Col))
		if err != nil {
			return nil, err
		}
		cells = place(cells, target, rebaseCell(parsed, d))
	}
	return cells, nil
}

// pasteBounds rejects a paste whose corners are negative or whose footprint
// reaches the grid limit, mirroring Set's validation: a content edit rejects
// bad input rather than silently no-opping.
func pasteBounds(at, origin Address, block Grid, limits Limits) error {
	if at.Row < 0 || at.Col < 0 {
		return constants.ErrInvalidValue.With(nil, "address", at.String())
	}
	if origin.Row < 0 || origin.Col < 0 {
		return constants.ErrInvalidValue.With(nil, "origin", origin.String())
	}
	if at.Row+len(block) > limits.GridDim || at.Col+widestBlockRow(block) > limits.GridDim {
		return constants.ErrInvalidValue.With(nil, "paste exceeds the grid limit", at.String())
	}
	return nil
}

// widestBlockRow is the widest row of a block, the columns its footprint spans.
func widestBlockRow(block Grid) int {
	w := 0
	for _, row := range block {
		w = max(w, len(row))
	}
	return w
}
