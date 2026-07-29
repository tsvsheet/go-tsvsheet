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
//
// URL is the location the fetcher actually reached, which is not always the
// source the sheet named: a relative source is resolved against the operator's
// data base by the fetcher, so the engine cannot know it. Purely informational —
// it feeds EXPLAIN so an author can see where a value came from, and nothing in
// the compute path reads it. A fetcher that leaves it empty is valid.
type FetchResult struct {
	ContentType MediaType
	URL         ImportURL
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
