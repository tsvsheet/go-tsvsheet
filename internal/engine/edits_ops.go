// The op table: one parser per wire op.
//
// edits.go owns the envelope — how a batch is framed, revision-checked and
// folded over a document. This file owns what each line MEANS: the operand
// grammar of every op, and the table that is the edit surface itself. The two
// change for different reasons, which is why they are separate files: adding an
// op touches only this one, and changing how a batch is applied touches only
// the other.
package engine

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
)

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
