// Materialization: the one scan-and-read path every document size shares
// (spec 016). A source is indexed with the engine's own single definitions —
// scanLine terminates lines, IsCommentLine classifies them — and its grid rows
// are parsed into cells through the strided block reader. A fully materialized
// sheet is simply this path with an unbounded cache; the windowed and lazy
// modes to come are the same path under policy, never a second implementation.
package engine

import (
	"errors"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/index"
)

// allCells is the cache capacity that keeps every block resident — full
// materialization.
const allCells index.CellCount = 1 << 62

// indexOptions wires the engine's line and comment rules into the index
// package, which is deliberately language-blind.
func indexOptions() index.Options {
	return index.Options{
		Split: scanLine,
		IsComment: func(line index.LineNumber, text []byte) bool {
			return IsCommentLine(LineNumber(line), SourceLine(text))
		},
		MaxLineBytes: maxLineBytes,
	}
}

// materializeSheet reads every grid row of an indexed source into a resident
// Sheet. A formula syntax error surfaces as the familiar cell-naming
// ErrSyntax; a scan failure keeps Parse's documented ErrReadInput (pinned by
// TestParseMapsScanFailuresToErrReadInput).
func materializeSheet(src sheetSource) (Sheet, error) {
	rows, err := src.reader.ReadRows(0, index.RowCount(src.ix.Census().Rows))
	if err != nil {
		return Sheet{}, readFailure(err)
	}
	if rows == nil {
		rows = [][]cell{}
	}
	return Sheet{cells: rows}, nil
}

// sheetSource is an indexed byte source: the census and the cell reader over
// one io.ReaderAt.
type sheetSource struct {
	reader *index.Reader[[]cell]
	ix     index.Index
}

// newSheetSource scans src once and stands a cell reader over it, its cache
// bounded to capacity cells: full materialization passes allCells, the
// windowed capability passes its resident budget — the adversary's finding
// that an unbounded windowed cache re-materializes the whole file under
// scrolling.
func newSheetSource(src readerAtSize, capacity index.CellCount) (sheetSource, error) {
	opts := indexOptions()
	// The scan also marks candidate declaration rows (any line carrying the
	// meta marker's bytes), capped at the same capacity that bounds the cache:
	// one ceiling (R16), and the windowed bindings pass reads only these rows.
	// A false candidate — the marker inside a string literal — parses to no
	// clause and binds nothing.
	opts.Mark, opts.MaxMarks = []byte(metaMarker), capacity
	ix, err := index.Scan(src.at, src.size, opts)
	if err != nil {
		return sheetSource{}, readFailure(err)
	}
	reader := index.NewReader(src.at, src.size, ix, indexOptions(), capacity, parseRowCells, identityRow)
	return sheetSource{ix: ix, reader: reader}, nil
}

// readerAtSize pairs a byte source with its length.
type readerAtSize struct {
	at   readAtSource
	size index.SourceSize
}

// readAtSource is the byte-source seam: a file, a spooled stream, or in-memory
// bytes.
type readAtSource = interface {
	ReadAt(p []byte, off int64) (int, error)
}

// parseRowCells parses one grid row's fields into cells, naming the cell on a
// formula syntax error — sheet parsing's standing contract (pinned by
// TestParse_SyntaxErrorNamesCell).
func parseRowCells(row index.GridRow, fields []string) ([]cell, error) {
	out := make([]cell, len(fields))
	for c, text := range fields {
		parsed, err := parseCell(textVal(text), rowIndex(row), colIndex(c))
		if err != nil {
			return nil, err
		}
		out[c] = parsed
	}
	return out, nil
}

// identityRow passes a parsed row through: cells are immutable, sharing
// nothing a caller may write.
func identityRow(row []cell) []cell { return row }

// readFailure maps an index scan failure to Parse's documented ErrReadInput;
// any other error (a syntax error naming its cell) is already the contract's
// own and passes verbatim.
func readFailure(err error) error {
	if errors.Is(err, index.ErrScan) {
		return constants.ErrReadInput.With(err)
	}
	return err
}
