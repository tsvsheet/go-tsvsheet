// TSV serialization for the sheet engine: reading a .tsvt grid into a Grid of
// raw cell strings and writing a computed Grid back out. The A1 reference
// resolution and formula evaluation live in model.go and eval.go; this file is
// just the tab-separated line format. (The package doc is in model.go.)
package engine

import (
	"bufio"
	"io"
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
)

// tab is the single field separator; newline terminates a row.
const (
	tab     = "\t"
	newline = "\n"
)

// Grid is a rectangular value grid indexed [row][col], 0-based. Cells are raw
// strings: a literal's own text on input, or a formula cell's computed value
// after ComputeAt.
type Grid [][]string

// ReadTSV reads a tab-separated value grid. Rows are newline-separated; a
// trailing newline does not add an empty row. Full-line comments are skipped
// and do not occupy a grid row, per IsCommentLine: a leading `#!` on the first
// line (a shebang, so a .tsvt can be `chmod +x` and run via
// `#!/usr/bin/env tsvsheet`), any `#.` directive-or-comment line, and any
// legacy `# ` hash-space line. An error-value cell like `#N/A` is data, not a
// comment. A read failure surfaces as ErrReadInput.
func ReadTSV(r io.Reader) (Grid, error) {
	grid := Grid{}
	err := scanLines(r, func(text string, isComment bool) {
		if !isComment {
			grid = append(grid, strings.Split(text, tab))
		}
	})
	if err != nil {
		return nil, err
	}
	return grid, nil
}

// LineNumber is a 1-based physical line position in a .tsvt source file, as
// opposed to a grid row: comment lines occupy a line but no row.
type LineNumber int

// SourceLine is one physical line of .tsvt source, newline already stripped.
type SourceLine string

// Line markers of SPECIFICATION §3. shebangMarker opens the first line only;
// directiveMarker opens a directive or comment line anywhere; legacyMarker is
// the superseded hash-space comment form, still accepted so that sheets and
// share links written before the change keep parsing. No marker involves
// whitespace except the legacy one — which is exactly why it was superseded, a
// space being indistinguishable from the TAB that separates fields.
const (
	shebangMarker   = "#!"
	directiveMarker = "#."
	legacyMarker    = "# "
)

// IsCommentLine reports whether the line at 1-based position at is one the grid
// skips: a first-line `#!` shebang, a `#.` directive-or-comment line, or a
// legacy `# ` hash-space comment. Everything else is data — including `#N/A`
// (hash then a letter) and `#<TAB>x` (hash then a TAB), so a mistyped marker
// shows up as a visible row rather than silently shifting A1 addresses.
//
// This is the single definition of the rule; frontends call it rather than
// re-testing the prefixes themselves.
func IsCommentLine(at LineNumber, text SourceLine) bool {
	line := string(text)
	if at == 1 && strings.HasPrefix(line, shebangMarker) {
		return true
	}
	return strings.HasPrefix(line, directiveMarker) || strings.HasPrefix(line, legacyMarker)
}

// scanLines reads r line by line, calling fn with each line's text and whether
// IsCommentLine classifies it as a comment — the lines ReadTSV skips and
// Document preserves. A read failure surfaces as ErrReadInput.
func scanLines(r io.Reader, fn func(text string, isComment bool)) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxLineBytes)

	lineNum := LineNumber(0)
	for scanner.Scan() {
		lineNum++
		text := scanner.Text()
		fn(text, IsCommentLine(lineNum, SourceLine(text)))
	}
	if err := scanner.Err(); err != nil {
		return constants.ErrReadInput.With(err)
	}
	return nil
}

// maxLineBytes bounds a single scanned row (1 MiB) so a pathological input
// cannot exhaust memory silently.
const maxLineBytes = 1 << 20

// WriteTSV writes the grid as tab-separated rows, each terminated by a newline.
// A write failure surfaces as constants.ErrWriteFile. Callers wanting buffering
// pass a bufio.Writer; WriteTSV writes each row directly so a write error is
// reported at its source.
func WriteTSV(w io.Writer, g Grid) error {
	for _, row := range g {
		if _, err := io.WriteString(w, strings.Join(row, tab)+newline); err != nil {
			return constants.ErrWriteFile.With(err)
		}
	}
	return nil
}
