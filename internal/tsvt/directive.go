// View-directive values: the row/column selectors and counts that follow a key
// on a `#.` line (SPECIFICATION §3). The LINE is a TSV row split by the host,
// exactly like a grid row; only a VALUE is parsed here, through the grammar's
// own rowSelector/colSelector/countValue entry rules, so no reader splits
// selector text by hand and every implementation rejects the same inputs.
package tsvt

import (
	"strconv"

	"github.com/antlr4-go/antlr/v4"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	grammar "github.com/tsvsheet/go-tsvsheet/internal/grammar"
)

// SelectorText is the text of one directive value: `20-31,40`, `B-M`, or `1`.
type SelectorText string

// RowNumber is a 1-based physical row of the grid, as a directive writes it —
// the same numbering the row gutter shows, not the engine's 0-based index.
type RowNumber int

// ColumnLetters is an A1 column name (`B`, `AA`). Directive values carry the
// letters rather than an index so the conversion stays in one place, beside
// the address parsing that already owns it.
type ColumnLetters string

// Count is a non-negative directive count: how many header or frozen rows or
// columns a sheet declares. Zero declares none.
type Count int

// RowSpan is an inclusive run of rows. A bare `40` parses as First == Last.
type RowSpan struct {
	First RowNumber
	Last  RowNumber
}

// ColSpan is an inclusive run of columns. A bare `N` parses as First == Last.
type ColSpan struct {
	First ColumnLetters
	Last  ColumnLetters
}

// ParseRowSelector parses a row-axis directive value — a comma-separated list
// of 1-based inclusive spans (`20-31,40`). Text the grammar does not admit is
// constants.ErrSyntax; a span that parses but cannot mean anything (row zero,
// a backwards run) is constants.ErrInvalidValue.
func ParseRowSelector(src SelectorText) ([]RowSpan, error) {
	parser, sink := newValueParser(src)
	tree := parser.RowSelector()
	if sink.err != nil {
		return nil, sink.err
	}
	spans := make([]RowSpan, 0, len(tree.AllRowSpan()))
	for _, ctx := range tree.AllRowSpan() {
		span, err := buildRowSpan(ctx)
		if err != nil {
			return nil, err
		}
		spans = append(spans, span)
	}
	return spans, nil
}

// ParseColSelector parses a column-axis directive value — a comma-separated
// list of inclusive column-letter spans (`B-M`, `A,C-D`). Errors follow
// ParseRowSelector.
func ParseColSelector(src SelectorText) ([]ColSpan, error) {
	parser, sink := newValueParser(src)
	tree := parser.ColSelector()
	if sink.err != nil {
		return nil, sink.err
	}
	spans := make([]ColSpan, 0, len(tree.AllColSpan()))
	for _, ctx := range tree.AllColSpan() {
		span, err := buildColSpan(ctx)
		if err != nil {
			return nil, err
		}
		spans = append(spans, span)
	}
	return spans, nil
}

// ParseCount parses a count-valued directive value (`1`). A non-integer such
// as `1.5` lexes as a number but cannot count rows, so it is
// constants.ErrInvalidValue rather than a syntax error.
func ParseCount(src SelectorText) (Count, error) {
	parser, sink := newValueParser(src)
	tree := parser.CountValue()
	if sink.err != nil {
		return 0, sink.err
	}
	n, err := wholeNumber(numberText(tree.NUMBER().GetText()))
	return Count(n), err
}

// newValueParser builds a parser over one directive value with the default
// error listeners replaced by a collector, so a syntax error becomes
// constants.ErrSyntax instead of a message printed to stderr.
func newValueParser(src SelectorText) (*grammar.TsvsheetParser, *errorSink) {
	collector := &errorCollector{sink: &errorSink{}}

	lexer := grammar.NewTsvsheetLexer(antlr.NewInputStream(string(src)))
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(collector)

	parser := grammar.NewTsvsheetParser(antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel))
	parser.RemoveErrorListeners()
	parser.AddErrorListener(collector)

	return parser, collector.sink
}

// buildRowSpan converts one parsed row span, rejecting row zero (rows are
// 1-based) and a run that ends before it starts.
func buildRowSpan(ctx grammar.IRowSpanContext) (RowSpan, error) {
	numbers := ctx.AllNUMBER()
	first, err := wholeNumber(numberText(numbers[0].GetText()))
	if err != nil {
		return RowSpan{}, err
	}
	last := first
	if len(numbers) > 1 {
		if last, err = wholeNumber(numberText(numbers[1].GetText())); err != nil {
			return RowSpan{}, err
		}
	}
	if first < 1 || last < first {
		return RowSpan{}, constants.ErrInvalidValue.With(nil, "selector", ctx.GetText())
	}
	return RowSpan{First: RowNumber(first), Last: RowNumber(last)}, nil
}

// buildColSpan converts one parsed column span, rejecting a run whose columns
// are in descending order (`M-B`).
func buildColSpan(ctx grammar.IColSpanContext) (ColSpan, error) {
	letters := ctx.AllCOL()
	first := ColumnLetters(letters[0].GetText())
	last := first
	if len(letters) > 1 {
		last = ColumnLetters(letters[1].GetText())
	}
	if !ordered(first, last) {
		return ColSpan{}, constants.ErrInvalidValue.With(nil, "selector", ctx.GetText())
	}
	return ColSpan{First: first, Last: last}, nil
}

// ordered reports whether first names a column at or before last. A1 column
// names sort by length first (`Z` precedes `AA`), then lexically.
func ordered(first, last ColumnLetters) bool {
	if len(first) != len(last) {
		return len(first) < len(last)
	}
	return first <= last
}

// numberText is the source text of a NUMBER token — `20`, or the `1.5` the
// token also admits and a count cannot mean.
type numberText string

// wholeNumber converts a NUMBER token's text to an integer, rejecting the
// fractional form the token also admits.
func wholeNumber(text numberText) (int, error) {
	n, err := strconv.Atoi(string(text))
	if err != nil {
		return 0, constants.ErrInvalidValue.With(err, "number", string(text))
	}
	return n, nil
}
