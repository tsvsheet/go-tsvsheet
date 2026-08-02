// TSV serialization for the sheet engine: reading a .tsvt grid into a Grid of
// raw cell strings and writing a computed Grid back out. The A1 reference
// resolution and formula evaluation live in model.go and eval.go; this file is
// just the tab-separated line format. (The package doc is in model.go.)
package engine

import (
	"bufio"
	"bytes"
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
// whitespace except the legacy one — which is why it was superseded, a space
// being indistinguishable from the TAB that separates fields.
const (
	shebangMarker   = "#!"
	directiveMarker = "#."
	legacyMarker    = "# "
)

// WouldStartCommentLine reports whether text placed in a row's first cell
// would make the serialized line a comment — which would delete that row from
// the grid on the next read, silently shifting every address below it. An edit
// that would do this is refused rather than written, because the language has
// no escape for a leading marker (SPECIFICATION §3); the shebang form is
// refused in any row, since a row can become line 1 under a later edit.
func WouldStartCommentLine(at Address, text CellText) bool {
	if at.Col != 0 {
		return false
	}
	line := string(text)
	return strings.HasPrefix(line, shebangMarker) ||
		strings.HasPrefix(line, directiveMarker) ||
		strings.HasPrefix(line, legacyMarker)
}

// CellText is a cell's source text as an editor supplies it.
type CellText string

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
	scanner.Split(scanLine)

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
// cannot exhaust memory silently. The bound applies per CR/LF-terminated
// segment: since scanLine treats a lone CR as a terminator, an input whose
// only breaks are CRs now scans as bounded rows where it was formerly one
// over-long token — a deliberate relaxation; a terminator-free run over the
// bound is still refused.
const maxLineBytes = 1 << 20

// scanLine is the bufio.SplitFunc for .tsvt lines: LF, CRLF, and a lone CR
// each terminate a line — the same normalization ParseBlock applies to
// clipboard text. bufio's default split strips only a CR sitting before an LF,
// so a lone CR survived into cell text, where no TSV delimiter may live: the
// canonical serialization wrote it back mid-line and the round-trip broke
// (FuzzParseDocument's "\r\r" parsed to a cell holding a bare CR).
var scanLine bufio.SplitFunc = func(data []byte, isAtEOF bool) (advance int, token []byte, err error) {
	i := bytes.IndexAny(data, "\r\n")
	if i < 0 {
		if isAtEOF && len(data) > 0 {
			return len(data), data, nil
		}
		return 0, nil, nil
	}
	if data[i] == '\n' {
		return i + 1, data[:i], nil
	}
	return crLine(data, terminatorAt(i), isFinal(isAtEOF))
}

// terminatorAt is the index of a line terminator within a scan buffer.
type terminatorAt int

// isFinal reports whether the scanner has reached the end of its input.
type isFinal bool

// crLine terminates a line at a CR, consuming a following LF as part of the
// same CRLF terminator; mid-stream, with the CR as the last buffered byte, it
// asks for one more byte to tell a lone CR from a split CRLF.
func crLine(data []byte, at terminatorAt, isEOF isFinal) (advance int, token []byte, err error) {
	if int(at)+1 < len(data) {
		if data[at+1] == '\n' {
			return int(at) + 2, data[:at], nil
		}
		return int(at) + 1, data[:at], nil
	}
	if isEOF {
		return int(at) + 1, data[:at], nil
	}
	return 0, nil, nil
}

// WriteTSV writes the grid as tab-separated rows, each terminated by a newline.
// A write failure surfaces as constants.ErrWriteFile. Callers wanting buffering
// pass a bufio.Writer; WriteTSV writes each row directly so a write error is
// reported at its source.
//
// A row whose first cell would serialize to a comment-reading line (a leading
// `#!`, `#.`, or `# ` — WouldStartCommentLine's rule) is refused as
// constants.ErrInvalidValue rather than written: the language has no escape
// for a leading marker, so the emitted line would read back as a comment and
// the row would silently vanish on the next read — data loss dressed as
// output. (FuzzReadTSV found the shape: a document may legally hold `#!x` as
// data below its comment lines, but the bare-grid format cannot represent it.)
func WriteTSV(w io.Writer, g Grid) error {
	for r, row := range g {
		if len(row) > 0 && WouldStartCommentLine(Address{Row: r}, CellText(row[0])) {
			return constants.ErrInvalidValue.With(
				nil,
				"row would read back as a comment line",
				Address{Row: r}.String(),
			)
		}
		if _, err := io.WriteString(w, strings.Join(row, tab)+newline); err != nil {
			return constants.ErrWriteFile.With(err)
		}
	}
	return nil
}
