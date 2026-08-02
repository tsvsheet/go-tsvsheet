// Package tsvt's directive builder: the walk from a parsed directive value's
// parse tree to the immutable Item/Span AST, kept apart from the directive
// types and the parser entry point.
package tsvt

import (
	"strconv"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	grammar "github.com/tsvsheet/go-tsvsheet/internal/grammar"
)

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
	switch name := funcName(ctx.NAME().GetText()); name {
	case fnCount:
	case fnRange:
		// `range(40)` has a count's shape; say which spelling was meant.
		n := ctx.Offset().GetText()
		return Item{}, constants.ErrInvalidValue.With(nil, "item", ctx.GetText(),
			"hint", "a range takes colon spans: range("+n+":"+n+")")
	default:
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
	switch name := funcName(ctx.NAME().GetText()); name {
	case fnRange:
	default:
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
