// View-directive values: the axis call that follows a key on a `#.` line
// (SPECIFICATION §3). The LINE is a TSV row split by the host, exactly like a
// grid row; only a VALUE is parsed here, through the grammar's own
// directiveValue entry rule, so no reader splits directive text by hand and
// every implementation refuses the same inputs.
//
// The value is written in the shape of a formula but is NOT the expression
// language: a cell reference or a volatile call cannot appear, so value-driven
// hiding cannot arrive through a directive while that design is still open.
package tsvt

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	grammar "github.com/tsvsheet/go-tsvsheet/internal/grammar"
)

// ValueText is the text of one directive value: `rows(range(20:31), count(3))`.
type ValueText string

// Axis names which way a directive value selects.
type Axis int

// The two axes a grid has, and the only two it will ever have.
const (
	AxisRow Axis = iota
	AxisCol
)

// ItemKind distinguishes the two item forms. They are never interchangeable:
// a range is absolute and shifts under a structural edit, a count is anchored
// to an edge and re-resolves against it.
type ItemKind int

// The item forms, both always written out — neither is a default, so a bare
// number can never be read as the wrong one.
const (
	ItemRange ItemKind = iota
	ItemCount
)

// Offset is a position on either axis. A positive value counts from the start
// (row 1, column A = 1); a negative value counts from the end, the convention
// isnow established across this family, so -1 is the last row or column.
type Offset int

// Span is one inclusive run inside a range call.
type Span struct {
	First Offset
	Last  Offset
}

// Item is one selector in a directive value: a range call carrying one or more
// spans (`range(20:31, 40:40)`), or a count at an edge (`count(3)`). The spans
// of one call stay together so a rewrite can shift endpoints without regrouping
// what the author wrote.
type Item struct {
	Spans []Span
	Count Offset
	Kind  ItemKind
}

// DirectiveValue is a parsed axis call: which way it selects, and the items it
// unions.
type DirectiveValue struct {
	Items []Item
	Axis  Axis
}

// funcName is a directive function's name as written — `rows`, `range`, or
// whatever a reader typed instead.
type funcName string

// hintText is the spelling a diagnostic asks for, so the error teaches the
// language rather than only refusing it.
type hintText string

// Directive function names. They resolve in this scope only — `=rows(...)` in
// a cell is #NAME?, because a view is not something a formula may observe.
const (
	fnRows  funcName = "rows"
	fnCols  funcName = "cols"
	fnRange funcName = "range"
	fnCount funcName = "count"
)

// ParseDirectiveValue parses one `#.` directive value. Text the grammar does
// not admit is constants.ErrSyntax, carrying the spelling it wanted; text that
// parses but cannot mean anything — an unknown function, an axis whose items
// name the other axis, a descending span, a zero count — is
// constants.ErrInvalidValue.
func ParseDirectiveValue(src ValueText) (DirectiveValue, error) {
	parser, sink := newValueParser(src)
	tree := parser.DirectiveValue()
	if sink.err != nil {
		return DirectiveValue{}, adviseSyntax(src, sink.err)
	}
	axis, err := axisOf(funcName(tree.AxisCall().NAME().GetText()))
	if err != nil {
		return DirectiveValue{}, err
	}
	items, err := buildItems(tree.AxisCall().ItemList().AllItem(), axis)
	if err != nil {
		return DirectiveValue{}, err
	}
	return DirectiveValue{Axis: axis, Items: items}, nil
}

// newValueParser builds a parser over one directive value with the default
// error listeners replaced by a collector, so a syntax error becomes
// constants.ErrSyntax instead of a message printed to stderr.
func newValueParser(src ValueText) (*grammar.TsvsheetParser, *errorSink) {
	collector := &errorCollector{sink: &errorSink{}}

	lexer := grammar.NewTsvsheetLexer(antlr.NewInputStream(string(src)))
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(collector)

	parser := grammar.NewTsvsheetParser(antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel))
	parser.RemoveErrorListeners()
	parser.AddErrorListener(collector)

	return parser, collector.sink
}

// adviseSyntax attaches the spelling the language wanted to a syntax error,
// because the two refusals a reader hits most — a bare item and a bare or
// comma-separated endpoint — are exactly the ones whose fix is not obvious.
func adviseSyntax(src ValueText, err error) error {
	// The parser's own position and token names are dropped deliberately: a
	// directive value is one short field, so quoting it beside the spelling it
	// should have had says more than "mismatched input '3' expecting NAME".
	if head, _, found := strings.Cut(string(src), "("); found && !isAxisName(funcName(head)) {
		return constants.ErrSyntax.With(nil, "value", string(src),
			"hint", "a directive value selects an axis: rows(…) or cols(…)")
	}
	for _, advice := range syntaxAdvice {
		if strings.Contains(string(src), advice.when) {
			return constants.ErrSyntax.With(nil, "value", string(src), "hint", string(advice.say))
		}
	}
	return constants.ErrSyntax.With(err, "value", string(src))
}

// advice pairs a substring of the offending value with the spelling to suggest,
// so the two refusals a reader hits most often carry their own fix.
type advice struct {
	when string
	say  hintText
}

// syntaxAdvice is scanned in order: the more specific suggestion first.
var syntaxAdvice = []advice{
	{when: string(fnRange) + "(", say: "a range takes colon spans: range(20:31), and range(3:3) for one"},
	{when: "(", say: "every item is a call: range(3:3) for one row or column, count(3) for the first three"},
}

// isAxisName reports whether a leading call name is one of the axes, so a
// syntax error can say which spelling the value should have opened with.
func isAxisName(name funcName) bool {
	trimmed := funcName(strings.TrimSpace(string(name)))
	return trimmed == fnRows || trimmed == fnCols
}

// axisOf resolves an axis function name, naming both spellings when it is not
// one of them — an unknown name is a mistake worth a suggestion, not a bare
// failure.
func axisOf(name funcName) (Axis, error) {
	switch name {
	case fnRows:
		return AxisRow, nil
	case fnCols:
		return AxisCol, nil
	}
	return 0, constants.ErrInvalidValue.With(nil,
		"function", string(name), "hint", "a directive value selects an axis: rows(…) or cols(…)")
}

// unknownItem reports a function name that is not one of the item forms,
// naming both so the message teaches the language.
func unknownItem(name funcName) error {
	return constants.ErrInvalidValue.With(nil,
		"function", string(name), "hint", "an item is range(20:31) or count(3)")
}

// wrongAxis reports an endpoint written for the other axis, with the spelling
// the axis wanted.
func wrongAxis(text endpointText, hint hintText) error {
	return constants.ErrInvalidValue.With(nil, "endpoint", string(text), "hint", string(hint))
}
