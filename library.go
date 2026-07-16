// Package tsvsheet is the engine for the tsvsheet single-file spreadsheet: a
// .tsvt is a single TAB-separated grid whose cells are literal values or
// =formulas that address other cells in A1 notation (B2, D2:D5), computed in
// place.
//
// The package parses a grid (Parse, ReadTSV), computes it with an Excel- and
// Google-Sheets-faithful expression evaluator that carries error values
// (#REF!, #DIV/0!, #CIRC!, …) through a dependency-ordered, memoized pass
// (Compute, ComputeWith), and inspects the result (Check diagnostics, Explain
// traces) before rendering it back to TSV (WriteTSV). Formula compilation
// reuses the grammar repo's ANTLR-generated expression parser through the
// internal/tsvt seam; no ANTLR type escapes into the public surface.
//
// The engine is filesystem- and network-free by construction: cross-sheet
// embedding (SHEET/INPUT/OUTPUT) and imports (IMPORT*) resolve only through the
// Loader and Fetcher a caller injects, and every allocation is bounded by an
// injected Limits ceiling. Errors returned to callers are the errs.Const
// sentinels re-exported from errors.go, matchable with errors.Is.
package tsvsheet
