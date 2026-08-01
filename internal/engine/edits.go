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
	"encoding/base64"
	"encoding/hex"
	"strconv"
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

// editArgs is an op's expected argument count.
type editArgs int

// editParser parses one op's argument fields into its application.
type editParser func(at editLine, args []string) (editApply, error)

// editParsers maps each wire op name to its parser. The op set is exactly the
// Document edit surface; move ops are deliberately absent (cut/move semantics
// await their own ruling).
var editParsers = map[string]editParser{
	"setCell":      parseSetCell,
	"insertRow":    rowOp(Document.InsertRow),
	"deleteRow":    rowOp(Document.DeleteRow),
	"duplicateRow": rowOp(Document.DuplicateRow),
	"insertCol":    colOp(Document.InsertCol),
	"deleteCol":    colOp(Document.DeleteCol),
	"duplicateCol": colOp(Document.DuplicateCol),
	"fill":         parseFill,
	"paste":        parsePaste,
}

// editArity returns nil when args has exactly want fields, the arity sentinel
// otherwise.
func editArity(at editLine, args []string, want editArgs) error {
	if len(args) == int(want) {
		return nil
	}
	return constants.ErrEditsArity.With(nil, "line", int(at), "want", int(want), "got", len(args))
}

// parseEditAddress parses an A1 cell address argument.
func parseEditAddress(at editLine, s editField) (Address, error) {
	addr, err := ParseAddress(AddressText(s))
	if err != nil {
		return Address{}, constants.ErrEditsAddress.With(err, "line", int(at), "address", string(s))
	}
	return addr, nil
}

// parseEditRow parses a 1-based row-number argument into the Address form the
// row methods take.
func parseEditRow(at editLine, s editField) (Address, error) {
	row, err := strconv.Atoi(string(s))
	if err != nil || row < 1 {
		return Address{}, constants.ErrEditsAddress.With(err, "line", int(at), "row", string(s))
	}
	return Address{Row: row - 1}, nil
}

// parseEditCol parses a column-letters argument into the Address form the
// column methods take.
func parseEditCol(at editLine, s editField) (Address, error) {
	letters, digits := splitLetters(AddressText(s))
	if letters == "" || digits != "" || len(letters) != len(s) {
		return Address{}, constants.ErrEditsAddress.With(nil, "line", int(at), "column", string(s))
	}
	return Address{Col: lettersToIndex(columnLetters(letters))}, nil
}

// parseEditSpan parses a rectangle argument: `A1:B9`, or a single cell as the
// one-cell rectangle.
func parseEditSpan(at editLine, s editField) (Span, error) {
	fromText, toText, isRange := strings.Cut(string(s), ":")
	from, err := parseEditAddress(at, editField(fromText))
	if err != nil {
		return Span{}, err
	}
	if !isRange {
		return Span{From: from, To: from}, nil
	}
	to, err := parseEditAddress(at, editField(toText))
	if err != nil {
		return Span{}, err
	}
	return Span{From: from, To: to}, nil
}

// rowOp builds the parser for a one-argument row op delegating to method.
func rowOp(method func(Document, Address) Document) editParser {
	return func(at editLine, args []string) (editApply, error) {
		if err := editArity(at, args, 1); err != nil {
			return nil, err
		}
		row, err := parseEditRow(at, editField(args[0]))
		if err != nil {
			return nil, err
		}
		return func(d Document, _ Limits) (Document, error) { return method(d, row), nil }, nil
	}
}

// colOp builds the parser for a one-argument column op delegating to method.
func colOp(method func(Document, Address) Document) editParser {
	return func(at editLine, args []string) (editApply, error) {
		if err := editArity(at, args, 1); err != nil {
			return nil, err
		}
		col, err := parseEditCol(at, editField(args[0]))
		if err != nil {
			return nil, err
		}
		return func(d Document, _ Limits) (Document, error) { return method(d, col), nil }, nil
	}
}

// parseSetCell parses `setCell <address> <text>`; empty text clears the cell,
// and a leading `=` is a formula, verbatim.
func parseSetCell(at editLine, args []string) (editApply, error) {
	if err := editArity(at, args, 2); err != nil {
		return nil, err
	}
	addr, err := parseEditAddress(at, editField(args[0]))
	if err != nil {
		return nil, err
	}
	text := args[1]
	return func(d Document, limits Limits) (Document, error) { return d.SetCell(addr, text, limits) }, nil
}

// parseFill parses `fill <source> <target-span>` (Document.Fill semantics).
func parseFill(at editLine, args []string) (editApply, error) {
	if err := editArity(at, args, 2); err != nil {
		return nil, err
	}
	from, err := parseEditAddress(at, editField(args[0]))
	if err != nil {
		return nil, err
	}
	to, err := parseEditSpan(at, editField(args[1]))
	if err != nil {
		return nil, err
	}
	return func(d Document, limits Limits) (Document, error) {
		// Document.Fill takes no limits, so the bound is enforced here —
		// otherwise one short line could grow the grid without ceiling.
		if err := withinGrid(to, limits); err != nil {
			return Document{}, err
		}
		return d.Fill(from, to), nil
	}, nil
}

// withinGrid refuses a target rectangle reaching beyond the grid limit,
// mirroring the validation Set applies to a single address.
func withinGrid(to Span, limits Limits) error {
	corner := Address{Row: max(to.From.Row, to.To.Row), Col: max(to.From.Col, to.To.Col)}
	if corner.Row >= limits.GridDim || corner.Col >= limits.GridDim {
		return constants.ErrInvalidValue.With(nil, "span exceeds the grid limit", corner.String())
	}
	return nil
}

// parsePaste parses `paste <target> <origin> <base64-block>` — the block is
// base64 because a TSV block is the one argument that can itself contain TABs
// and newlines (Document.Paste semantics).
func parsePaste(at editLine, args []string) (editApply, error) {
	if err := editArity(at, args, 3); err != nil {
		return nil, err
	}
	target, err := parseEditAddress(at, editField(args[0]))
	if err != nil {
		return nil, err
	}
	origin, err := parseEditAddress(at, editField(args[1]))
	if err != nil {
		return nil, err
	}
	decoded, err := base64.StdEncoding.DecodeString(args[2])
	if err != nil {
		return nil, constants.ErrEditsBlock.With(err, "line", int(at))
	}
	block := ParseBlock(BlockText(decoded))
	return func(d Document, limits Limits) (Document, error) { return d.Paste(target, origin, block, limits) }, nil
}
