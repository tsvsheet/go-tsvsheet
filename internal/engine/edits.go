// The sheet edit language (work order 011): an edits document is a TSV stream
// of semantic operations applied to a Document as a pure left fold. Comment
// lines are metadata (`#.base`) or prose; every data line is one op mirroring
// a Document method, its arguments verbatim TSV fields in the language's own
// address forms. An Edits value round-trips verbatim: Text returns the exact
// bytes parsed.
package engine

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
)

// RevisionHex is a document's content address: the lowercase-hex SHA-256 of
// its canonical Text bytes. Equal revision ⇔ byte-equal document.
type RevisionHex string

// Revision content-addresses d.
func Revision(d Document) RevisionHex {
	sum := sha256.Sum256(d.Text())
	return RevisionHex(hex.EncodeToString(sum[:]))
}

// editLine is a 1-based line number in an edits document.
type editLine int

// editApply is one parsed op's application, closed over its arguments.
type editApply func(Document, Limits) (Document, error)

// editOp is one parsed operation and the line it came from.
type editOp struct {
	apply editApply
	at    editLine
}

// Edits is a parsed edits document. The zero value is empty. Edits is
// immutable: parsing returns a value whose Text is the exact input.
type Edits struct {
	raw  []byte
	base RevisionHex
	ops  []editOp
}

// Base returns the `#.base` revision the document was authored against, or ""
// when it names none (an unconditional batch). A batch naming several bases
// keeps the last, matching the metadata rule that a repeated key is a
// re-declaration; position carries no meaning, so a base line below the ops
// still governs the whole batch.
func (e Edits) Base() RevisionHex { return e.base }

// Len is the number of operations (comment and metadata lines are not ops).
func (e Edits) Len() int { return len(e.ops) }

// Text returns the edits document's exact source bytes.
func (e Edits) Text() []byte { return bytes.Clone(e.raw) }

// ParseEdits reads an edits document: `#.` and legacy comment lines carry
// metadata or prose, blank lines are prose, and every other line is one op.
// Any malformed line is a specific sentinel carrying its 1-based line number.
func ParseEdits(src []byte) (Edits, error) {
	edits := Edits{raw: bytes.Clone(src)}
	for i, text := range splitEditLines(src) {
		base, op, err := parseEditLine(editLine(i+1), SourceLine(text))
		if err != nil {
			return Edits{}, err
		}
		if base != "" {
			edits.base = base
		}
		if op.apply != nil {
			edits.ops = append(edits.ops, op)
		}
	}
	return edits, nil
}

// ParseEditsWith is ParseEdits under a residency budget (spec 018): a batch
// whose line count (each line is at most one op) exceeds Limits' effective
// resident ceiling refuses
// with ErrDocTooLarge before any op parses — each op touches at least one
// cell, so a batch larger than the cells a document may hold is over budget
// by construction, and no caller pays an unbounded op-slice allocation it did
// not raise its budget to accept.
func ParseEditsWith(src []byte, limits Limits) (Edits, error) {
	budget := limits.EffectiveResidentCells()
	if lines := int64(len(splitEditLines(src))); lines > budget {
		return Edits{}, constants.ErrDocTooLarge.With(nil, "ops", lines, "budget", budget)
	}
	return ParseEdits(src)
}

// Apply left-folds e's ops over d. When e names a base revision that is not
// d's, the whole batch is refused (constants.ErrEditsBase); when any op is
// refused, the whole batch is rejected with constants.ErrEditsApply wrapping
// the cause and naming the op's line — d is immutable, so nothing is ever
// partially applied. Nothing outside (d, e, limits) influences the result.
func Apply(d Document, e Edits, limits Limits) (Document, error) {
	if e.base != "" && e.base != Revision(d) {
		return Document{}, constants.ErrEditsBase.With(nil, "base", string(e.base), "document", string(Revision(d)))
	}
	out := d
	for _, op := range e.ops {
		next, err := op.apply(out, limits)
		if err != nil {
			return Document{}, constants.ErrEditsApply.With(err, "line", int(op.at))
		}
		out = next
	}
	return out, nil
}

// splitEditLines splits src into lines, without a trailing empty element when
// src ends in a newline. A CRLF terminator loses its CR: the grid's own reader
// strips it, so a CR carried into a cell would store text that does not
// survive a re-read — the document would stop being a fixed point of
// parse∘serialize and its content address would name bytes other than the ones
// on disk.
func splitEditLines(src []byte) []string {
	if len(src) == 0 {
		return nil
	}
	lines := strings.Split(string(src), newline)
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for i, line := range lines {
		lines[i] = strings.TrimSuffix(line, carriageReturn)
	}
	return lines
}

// carriageReturn is the CR a CRLF line terminator leaves behind.
const carriageReturn = "\r"

// parseEditLine classifies one line: comment and blank lines yield at most a
// base revision, data lines yield an op.
func parseEditLine(at editLine, text SourceLine) (RevisionHex, editOp, error) {
	if text == "" || IsCommentLine(LineNumber(at), text) {
		return editBase(text), editOp{}, nil
	}
	fields := strings.Split(string(text), tab)
	parse, ok := editParsers[fields[0]]
	if !ok {
		return "", editOp{}, constants.ErrEditsOp.With(nil, "line", int(at), "op", fields[0])
	}
	apply, err := parse(at, fields[1:])
	if err != nil {
		return "", editOp{}, err
	}
	return "", editOp{at: at, apply: apply}, nil
}

// editBase extracts the revision from a `#.base` metadata line ("" otherwise).
func editBase(text SourceLine) RevisionHex {
	fields := strings.Split(string(text), tab)
	if fields[0] != "#."+editKeyBase || len(fields) < 2 {
		return ""
	}
	return RevisionHex(fields[1])
}

// editKeyBase is the metadata key naming the revision a batch was authored
// against.
const editKeyBase = "base"

// editField is one op-argument field of an edits line, verbatim.
type editField string
