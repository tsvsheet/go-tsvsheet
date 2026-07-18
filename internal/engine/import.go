package engine

import (
	"bytes"
	"encoding/csv"
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// MediaType is a content-typed import's RFC 6838 media type — the Accept header
// an IMPORT* function requests, which the response Content-Type must match.
type MediaType string

// The content-typed import media types (ADR 0006 §2): the request Accept header
// each IMPORT* function sends, which the response Content-Type must match. The
// RFC 6838 vendor tree with a hierarchical subtype for granularity and the +tsv
// structured-syntax suffix.
const (
	mediaSheet  MediaType = "application/vnd.tsvsheet+tsv"
	mediaCell   MediaType = "application/vnd.tsvsheet.cell+tsv"
	mediaRow    MediaType = "application/vnd.tsvsheet.row+tsv"
	mediaColumn MediaType = "application/vnd.tsvsheet.column+tsv"
	mediaRange  MediaType = "application/vnd.tsvsheet.range+tsv"
)

// The standard tabular media types admitted alongside the vendor types (ADR
// 0010 §1): an endpoint deliberately publishing TSV or CSV — a Google Sheets
// export, a data portal, a CI artifact — speaks the tabular lingua franca
// instead of the vendor protocol. Every other base type is refused.
const (
	mediaTSV MediaType = "text/tab-separated-values"
	mediaCSV MediaType = "text/csv"
)

// Accept is the negotiation list an IMPORT* request sends for this vendor media
// type: the vendor type preferred, the standard tabular types admitted with
// descending quality (ADR 0010 §1). Frontends set it as the Accept header.
func (m MediaType) Accept() string {
	return string(m) + ", " + string(mediaTSV) + ";q=0.9, " + string(mediaCSV) + ";q=0.8"
}

// importMedia maps each lowercase import function name to the media type it
// requests — the name is the content type (ADR 0006 §2).
var importMedia = map[string]MediaType{
	"importcell":   mediaCell,
	"importrow":    mediaRow,
	"importcolumn": mediaColumn,
	"importrange":  mediaRange,
	"importsheet":  mediaSheet,
}

// ImportURL is the location an IMPORT* function fetches — the (already
// evaluated) string value of its single argument.
type ImportURL string

// FetchResult is a Fetcher's response: the raw body and the media type the
// server declared, which must match the requested Accept for the handshake to
// succeed (ADR 0006 §2).
type FetchResult struct {
	ContentType MediaType
	Body        []byte
}

// Fetcher retrieves the content-typed import at url, sending accept as the
// requested media type. The engine holds only this interface; the concrete
// net/http fetcher, allowlist, and caching are injected by a frontend. A nil
// Fetcher disables imports (every IMPORT* is #IMPORT!).
type Fetcher interface {
	Fetch(url ImportURL, accept MediaType) (FetchResult, error)
}

// isImportName reports whether name (already lowercased) is an import function.
func isImportName(name funcName) boolResult {
	_, ok := importMedia[string(name)]
	return boolResult(ok)
}

// HasImports reports whether any formula calls an IMPORT* function, so a
// frontend can offer a manual "refresh imports" control. Imports are NOT
// clock-volatile and are deliberately absent from IsVolatile — they must never
// ride the isnow refresh ticker (ADR 0006 §6).
func (s Sheet) HasImports() bool {
	found := false
	s.eachFormula(func(at Address) {
		walkCalls(s.cells[at.Row][at.Col].formula, func(call tsvt.Call) {
			if isImportName(funcName(strings.ToLower(call.Name))) {
				found = true
			}
		})
	})
	return found
}

// evalImport dispatches the five IMPORT* functions (ADR 0006 §4): each takes a
// single URL argument, requests its media type, and — on a matching handshake —
// parses the response as a values-only grid of the function's shape. ok is false
// for any non-import name; a wrong arity is #VALUE!, an error-valued URL
// propagates, and every fetch/handshake/parse failure is #IMPORT!.
func (r resolver) evalImport(name funcName, args []tsvt.Expr) (Value, boolResult) {
	media, ok := importMedia[string(name)]
	if !ok {
		return Value{}, false
	}
	if len(args) != 1 {
		return errorValue(ErrValue), true
	}
	url := r.eval(args[0])
	if url.isError() {
		return url, true
	}
	return r.fetchImport(ImportURL(url.String()), media), true
}

// fetchImport fetches url and parses the response into the import's value. A nil
// Fetcher (the plain Compute path, or a frontend that has not enabled imports)
// disables imports; a fetch error or a Content-Type outside the accept set —
// the requested vendor type or a standard tabular type — is #IMPORT! (ADR 0006
// §2, §4; ADR 0010 §1).
func (r resolver) fetchImport(url ImportURL, media MediaType) Value {
	if r.comp.fetcher == nil {
		return errorValue(ErrImport)
	}
	res, err := r.comp.fetcher.Fetch(url, media)
	if err != nil {
		return errorValue(ErrImport)
	}
	received := res.ContentType.base()
	if !acceptable(received, media) {
		return errorValue(ErrImport)
	}
	return parseImport(res.Body, received, media, r.comp.limits)
}

// base strips any media-type parameters (`; charset=utf-8`) and normalizes
// case, so handshake matching is against the base type alone (ADR 0010 §1).
func (m MediaType) base() MediaType {
	head, _, _ := strings.Cut(string(m), ";")
	return MediaType(strings.ToLower(strings.TrimSpace(head)))
}

// acceptable reports whether a received base Content-Type satisfies the
// handshake for the requested vendor media type: the vendor type itself or one
// of the standard tabular types (ADR 0010 §1).
func acceptable(received, media MediaType) boolResult {
	return received == media || received == mediaTSV || received == mediaCSV
}

// parseImport parses a fetched body as a values-only grid — by the reader the
// received base media type selects (ADR 0010 §3) — and shapes it to the
// requested media type. A read failure or an empty grid is #IMPORT!.
func parseImport(body []byte, received, media MediaType, limits Limits) Value {
	grid, err := readImport(body, received)
	if err != nil {
		return errorValue(ErrImport)
	}
	cells := importCells(grid)
	if len(cells) == 0 {
		return errorValue(ErrImport)
	}
	return shapeImport(cells, media, limits)
}

// readImport reads a fetched body into a raw grid: a text/csv body via RFC 4180
// (ragged rows tolerated here — shapeImport enforces rectangularity), anything
// else via the engine's TSV reader (ADR 0010 §3).
func readImport(body []byte, received MediaType) (Grid, error) {
	if received != mediaCSV {
		return ReadTSV(bytes.NewReader(body))
	}
	reader := csv.NewReader(bytes.NewReader(body))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, constants.ErrReadInput.With(err)
	}
	return records, nil
}

// importCells converts a fetched TSV grid to a value grid, VALUES ONLY: each
// cell parses as a literal, so a leading `=` stays literal text and is never
// compiled as a formula (ADR 0006 §3).
func importCells(grid Grid) [][]Value {
	cells := make([][]Value, 0, len(grid))
	for _, row := range grid {
		values := make([]Value, 0, len(row))
		for _, cell := range row {
			values = append(values, value(textVal(cell)))
		}
		cells = append(cells, values)
	}
	return cells
}

// shapeImport enforces the shape each import media type requires and returns the
// scalar (cell) or spilling array (row/column/range/sheet) result; a shape or
// size mismatch is #IMPORT!, never a salvage (ADR 0006 §4).
func shapeImport(cells [][]Value, media MediaType, limits Limits) Value {
	switch media {
	case mediaCell:
		return importScalar(cells)
	case mediaRow:
		return importRow(cells, limits)
	case mediaColumn:
		return importColumn(cells, limits)
	default: // mediaRange, mediaSheet
		return importRange(cells, limits)
	}
}

// importScalar shapes IMPORTCELL: the grid must be exactly one row of one cell,
// returned as that scalar value.
func importScalar(cells [][]Value) Value {
	if len(cells) != 1 || len(cells[0]) != 1 {
		return errorValue(ErrImport)
	}
	return cells[0][0]
}

// importRow shapes IMPORTROW: exactly one row (of one or more columns), returned
// as a 1×N array that spills horizontally.
func importRow(cells [][]Value, limits Limits) Value {
	if len(cells) != 1 {
		return errorValue(ErrImport)
	}
	row := cells[0]
	if oversize(limits, 1, resultDim(len(row))) {
		return errorValue(ErrImport)
	}
	return arrayValue([][]Value{row})
}

// importColumn shapes IMPORTCOLUMN: one or more rows, each exactly one cell,
// returned as an N×1 array that spills vertically.
func importColumn(cells [][]Value, limits Limits) Value {
	if !allWidth(cells, 1) {
		return errorValue(ErrImport)
	}
	if oversize(limits, resultDim(len(cells)), 1) {
		return errorValue(ErrImport)
	}
	return arrayValue(cells)
}

// importRange shapes IMPORTRANGE and IMPORTSHEET: a non-empty rectangular grid
// (every row the same width), returned as an R×C array that spills. For this
// engine chunk IMPORTSHEET behaves like IMPORTRANGE (a spilling values grid);
// the "nested grid inside one cell" rendering distinction is deferred to the
// frontend chunk — only the requested Accept media type (the handshake) differs.
func importRange(cells [][]Value, limits Limits) Value {
	width := resultDim(len(cells[0]))
	if !allWidth(cells, width) {
		return errorValue(ErrImport)
	}
	if oversize(limits, resultDim(len(cells)), width) {
		return errorValue(ErrImport)
	}
	return arrayValue(cells)
}

// allWidth reports whether every row of cells has exactly width columns.
func allWidth(cells [][]Value, width resultDim) boolResult {
	for _, row := range cells {
		if resultDim(len(row)) != width {
			return false
		}
	}
	return true
}

// oversize reports whether a rows×cols import result exceeds the injected cell
// budget (ADR 0006 §4) — an over-large response is #IMPORT!.
func oversize(limits Limits, rows, cols resultDim) boolResult {
	return boolResult(limits.tooManyCells(rows, cols))
}
