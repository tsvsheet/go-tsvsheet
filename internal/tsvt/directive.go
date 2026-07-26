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
	"strconv"
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
	if head, _, found := strings.Cut(string(src), "("); found && !isAxisName(funcName(head)) {
		return constants.ErrSyntax.With(err, "value", string(src),
			"hint", "a directive value selects an axis: rows(…) or cols(…)")
	}
	for _, advice := range syntaxAdvice {
		if strings.Contains(string(src), advice.when) {
			return constants.ErrSyntax.With(err, "value", string(src), "hint", string(advice.say))
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

// buildItems converts each parsed item, checking it against the axis it was
// written under so that a value naming the wrong axis is refused rather than
// accepted and silently selecting nothing.
func buildItems(parsed []grammar.IItemContext, axis Axis) ([]Item, error) {
	items := make([]Item, 0, len(parsed))
	for _, ctx := range parsed {
		built, err := buildItem(ctx, axis)
		if err != nil {
			return nil, err
		}
		items = append(items, built)
	}
	return items, nil
}

// buildItem converts one item — a count, or a range with all of its spans.
func buildItem(ctx grammar.IItemContext, axis Axis) (Item, error) {
	if call := ctx.CountCall(); call != nil {
		return buildCount(call)
	}
	return buildRange(ctx.RangeCall(), axis)
}

// buildCount converts `count(n)`. Zero selects nothing, which is likelier a
// mistake than an intent: a directive that declares nothing is written by
// omitting the line.
func buildCount(ctx grammar.ICountCallContext) (Item, error) {
	if name := funcName(ctx.NAME().GetText()); name != fnCount {
		if name == fnRange {
			// `range(40)` has a count's shape; say which spelling was meant.
			n := ctx.Offset().GetText()
			return Item{}, constants.ErrInvalidValue.With(nil, "item", ctx.GetText(),
				"hint", "a range takes colon spans: range("+n+":"+n+")")
		}
		return Item{}, unknownItem(name)
	}
	n, err := offsetOf(ctx.Offset())
	if err != nil {
		return Item{}, err
	}
	return Item{Kind: ItemCount, Count: n}, nil
}

// buildRange converts `range(a:b, …)`, validating every span against the axis
// and for direction.
func buildRange(ctx grammar.IRangeCallContext, axis Axis) (Item, error) {
	if name := funcName(ctx.NAME().GetText()); name != fnRange {
		return Item{}, unknownItem(name)
	}
	parsed := ctx.AllSpan()
	spans := make([]Span, 0, len(parsed))
	for _, ctx := range parsed {
		span, err := buildSpan(ctx, axis)
		if err != nil {
			return Item{}, err
		}
		spans = append(spans, span)
	}
	return Item{Kind: ItemRange, Spans: spans}, nil
}

// buildSpan converts one `a:b` span, rejecting endpoints that name the other
// axis and spans that run backwards as written.
func buildSpan(ctx grammar.ISpanContext, axis Axis) (Span, error) {
	ends := ctx.AllEndpoint()
	first, err := endpointOf(ends[0], axis)
	if err != nil {
		return Span{}, err
	}
	last, err := endpointOf(ends[1], axis)
	if err != nil {
		return Span{}, err
	}
	if !ascending(first, last) {
		return Span{}, constants.ErrInvalidValue.With(nil,
			"span", ctx.GetText(), "hint", "a span runs forwards: range(20:31)")
	}
	return Span{First: first, Last: last}, nil
}

// ascending reports whether a span runs forwards. Two offsets compare directly
// only when they are anchored to the same end; a positive start with a
// negative end always runs forwards, since it ends at or before the last.
func ascending(first, last Offset) bool {
	if (first < 0) != (last < 0) {
		return first > 0
	}
	return first <= last
}

// endpointOf converts one span endpoint. A column letter suits only the column
// axis and a plain number only the row axis, but a NEGATIVE number suits
// either: it counts from the end, which both axes have.
func endpointOf(ctx grammar.IEndpointContext, axis Axis) (Offset, error) {
	if col := ctx.COL(); col != nil {
		if axis != AxisCol {
			return 0, wrongAxis(endpointText(ctx.GetText()), "rows are numbered: range(20:31)")
		}
		return columnOffset(endpointText(ctx.GetText())), nil
	}
	n, err := offsetOf(ctx.Offset())
	if err != nil || n < 0 {
		return n, err
	}
	if axis == AxisCol {
		return 0, wrongAxis(endpointText(ctx.GetText()), "columns are lettered: range(B:M)")
	}
	return n, nil
}

// offsetOf converts a signed number, rejecting the fractional form the NUMBER
// token also admits and a zero position (both axes are 1-based).
func offsetOf(ctx grammar.IOffsetContext) (Offset, error) {
	n, err := strconv.Atoi(ctx.NUMBER().GetText())
	if err != nil {
		return 0, constants.ErrInvalidValue.With(err, "number", ctx.GetText())
	}
	if n == 0 {
		// Zero is no position on either axis, and `-0` is not one either — a
		// negative offset counts back from the last, where -1 is already the end.
		return 0, constants.ErrInvalidValue.With(nil,
			"position", ctx.GetText(), "hint", "rows and columns are numbered from 1, or from -1 at the end")
	}
	if ctx.DASH() != nil {
		return Offset(-n), nil
	}
	return Offset(n), nil
}

// endpointText is one endpoint as written — `20`, `-1`, or `B` — carried into
// a diagnostic so the reader sees the part that was wrong.
type endpointText string

// columnOffset converts A1 column letters to a 1-based position, so that the
// column axis compares and shifts exactly like the row axis.
func columnOffset(letters endpointText) Offset {
	n := 0
	for _, r := range string(letters) {
		n = n*26 + int(r-'A') + 1
	}
	return Offset(n)
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
