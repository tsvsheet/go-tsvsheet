package tsvt_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// TestParseDirectiveValue covers the shape a view directive's value takes: an
// axis call over a list of items, each item itself a named call. Both item
// forms are always written out — neither is a default — so a reader never has
// to know which one a bare number would have meant.
func TestParseDirectiveValue(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		text tsvt.ValueText
		want tsvt.DirectiveValue
	}{
		{
			name: "column range",
			text: "cols(range(B:M))",
			want: tsvt.DirectiveValue{
				Axis:  tsvt.AxisCol,
				Items: []tsvt.Item{{Kind: tsvt.ItemRange, Spans: []tsvt.Span{{First: 2, Last: 13}}}},
			},
		},
		{
			name: "row range",
			text: "rows(range(20:31))",
			want: tsvt.DirectiveValue{
				Axis:  tsvt.AxisRow,
				Items: []tsvt.Item{{Kind: tsvt.ItemRange, Spans: []tsvt.Span{{First: 20, Last: 31}}}},
			},
		},
		{
			name: "single row still spells its span",
			text: "rows(range(40:40))",
			want: tsvt.DirectiveValue{
				Axis:  tsvt.AxisRow,
				Items: []tsvt.Item{{Kind: tsvt.ItemRange, Spans: []tsvt.Span{{First: 40, Last: 40}}}},
			},
		},
		{
			name: "variadic range unions its spans",
			text: "rows(range(20:31, 40:40))",
			want: tsvt.DirectiveValue{
				Axis: tsvt.AxisRow,
				Items: []tsvt.Item{
					{Kind: tsvt.ItemRange, Spans: []tsvt.Span{{First: 20, Last: 31}, {First: 40, Last: 40}}},
				},
			},
		},
		{
			name: "count of the first rows",
			text: "rows(count(3))",
			want: tsvt.DirectiveValue{
				Axis:  tsvt.AxisRow,
				Items: []tsvt.Item{{Kind: tsvt.ItemCount, Count: 3}},
			},
		},
		{
			name: "count from the end",
			text: "rows(count(-1))",
			want: tsvt.DirectiveValue{
				Axis:  tsvt.AxisRow,
				Items: []tsvt.Item{{Kind: tsvt.ItemCount, Count: -1}},
			},
		},
		{
			name: "range to the last row",
			text: "rows(range(20:-1))",
			want: tsvt.DirectiveValue{
				Axis:  tsvt.AxisRow,
				Items: []tsvt.Item{{Kind: tsvt.ItemRange, Spans: []tsvt.Span{{First: 20, Last: -1}}}},
			},
		},
		{
			name: "the last row alone",
			text: "rows(range(-1:-1))",
			want: tsvt.DirectiveValue{
				Axis:  tsvt.AxisRow,
				Items: []tsvt.Item{{Kind: tsvt.ItemRange, Spans: []tsvt.Span{{First: -1, Last: -1}}}},
			},
		},
		{
			name: "mixed items in one list",
			text: "rows(count(2), count(-1))",
			want: tsvt.DirectiveValue{
				Axis: tsvt.AxisRow,
				Items: []tsvt.Item{
					{Kind: tsvt.ItemCount, Count: 2},
					{Kind: tsvt.ItemCount, Count: -1},
				},
			},
		},
		{
			name: "columns to the last are numeric from the end",
			text: "cols(range(B:-1))",
			want: tsvt.DirectiveValue{
				Axis:  tsvt.AxisCol,
				Items: []tsvt.Item{{Kind: tsvt.ItemRange, Spans: []tsvt.Span{{First: 2, Last: -1}}}},
			},
		},
		{
			name: "multi-letter columns",
			text: "cols(range(Z:AA))",
			want: tsvt.DirectiveValue{
				Axis:  tsvt.AxisCol,
				Items: []tsvt.Item{{Kind: tsvt.ItemRange, Spans: []tsvt.Span{{First: 26, Last: 27}}}},
			},
		},
		{
			name: "padding is insignificant",
			text: "rows( range( 20 : 31 ) , count( 3 ) )",
			want: tsvt.DirectiveValue{
				Axis: tsvt.AxisRow,
				Items: []tsvt.Item{
					{Kind: tsvt.ItemRange, Spans: []tsvt.Span{{First: 20, Last: 31}}},
					{Kind: tsvt.ItemCount, Count: 3},
				},
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := tsvt.ParseDirectiveValue(c.text)
			require.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}
}

// TestParseDirectiveValueRefusals pins the three refusals the language exists
// to make — a bare item, a bare endpoint, and a comma standing in for a span —
// plus the unknown-name cases. Each refused text would otherwise read two ways,
// and each message names the spelling it wants, so the error teaches rather
// than scolds.
func TestParseDirectiveValueRefusals(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		want error
		name string
		text tsvt.ValueText
		says string
	}{
		{name: "bare item", text: "rows(3)", says: "range(3:3)", want: constants.ErrSyntax},
		{name: "bare column item", text: "cols(B)", says: "range", want: constants.ErrSyntax},
		{name: "bare endpoint", text: "rows(range(40))", says: "range(40:40)", want: constants.ErrInvalidValue},
		{name: "comma for a span", text: "rows(range(20,31))", says: "20:31", want: constants.ErrSyntax},
		{name: "unknown axis", text: "row(range(1:1))", says: "rows", want: constants.ErrInvalidValue},
		{name: "unknown item", text: "rows(span(1:1))", says: "range", want: constants.ErrInvalidValue},
		{name: "unknown count-shaped item", text: "rows(first(3))", says: "count(3)", want: constants.ErrInvalidValue},
		{name: "count is not an axis", text: "count(3)", says: "rows", want: constants.ErrSyntax},
		{name: "empty list", text: "rows()", want: constants.ErrSyntax},
		{name: "no value at all", text: "", want: constants.ErrSyntax},
		{name: "trailing text", text: "rows(count(3)) please", want: constants.ErrSyntax},
		{name: "cell reference is not admitted", text: "rows(range(A1:A5))", want: constants.ErrSyntax},
		{name: "a formula is not a directive", text: "=sum(A1:A3)", want: constants.ErrSyntax},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := tsvt.ParseDirectiveValue(c.text)
			require.Error(t, err)
			assert.True(t, errors.Is(err, c.want), "got %v, want %v", err, c.want)
			if c.says != "" {
				assert.Contains(t, err.Error(), c.says, "the diagnostic must name the spelling it wants")
			}
		})
	}
}

// TestParseDirectiveValueAxisMismatch covers why the axis is checked against
// its items: a value naming the wrong axis is refused rather than accepted and
// silently selecting nothing. A negative offset is axis-neutral — it counts
// from the end on either axis — so both admit it.
func TestParseDirectiveValueAxisMismatch(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		text tsvt.ValueText
		ok   bool
	}{
		{name: "columns where rows belong", text: "rows(range(B:M))", ok: false},
		{name: "rows where columns belong", text: "cols(range(20:31))", ok: false},
		{name: "mixed endpoints in one span", text: "rows(range(1:C))", ok: false},
		{name: "negative offsets suit either axis", text: "cols(count(-1))", ok: true},
		{name: "negative endpoints suit either axis", text: "cols(range(-2:-1))", ok: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := tsvt.ParseDirectiveValue(c.text)
			if c.ok {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.True(t, errors.Is(err, constants.ErrInvalidValue), "got %v", err)
		})
	}
}

// TestParseDirectiveValueRejectsMeaningless covers values that parse but cannot
// mean anything: a span whose endpoints descend, and a count of zero, which
// selects nothing and is likelier a mistake than an intent — a directive that
// declares nothing is written by omitting the line.
func TestParseDirectiveValueRejectsMeaningless(t *testing.T) {
	t.Parallel()

	for _, text := range []tsvt.ValueText{
		"rows(range(31:20))",
		"cols(range(M:B))",
		"cols(range(AA:Z))",
		"rows(range(-1:-3))",
		"rows(count(0))",
		"rows(range(0:3))",
		"rows(count(1.5))",
		"rows(count(-0))",
		"rows(range(-0:3))",
		"rows(range(1.5:3))",
	} {
		t.Run(string(text), func(t *testing.T) {
			t.Parallel()
			_, err := tsvt.ParseDirectiveValue(text)
			require.Error(t, err)
			assert.True(t, errors.Is(err, constants.ErrInvalidValue), "got %v", err)
		})
	}
}
