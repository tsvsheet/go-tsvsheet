package tsvt

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
)

// TestParseDirectiveValue covers the shape a view directive's value takes: an
// axis call over a list of items, each item itself a named call. Both item
// forms are always written out — neither is a default — so a reader never has
// to know which one a bare number would have meant.
func TestParseDirectiveValue(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		text ValueText
		want DirectiveValue
	}{
		{
			name: "column range",
			text: "cols(range(B:M))",
			want: DirectiveValue{
				Axis:  AxisCol,
				Items: []Item{{Kind: ItemRange, Spans: []Span{{First: 2, Last: 13}}}},
			},
		},
		{
			name: "row range",
			text: "rows(range(20:31))",
			want: DirectiveValue{
				Axis:  AxisRow,
				Items: []Item{{Kind: ItemRange, Spans: []Span{{First: 20, Last: 31}}}},
			},
		},
		{
			name: "single row still spells its span",
			text: "rows(range(40:40))",
			want: DirectiveValue{
				Axis:  AxisRow,
				Items: []Item{{Kind: ItemRange, Spans: []Span{{First: 40, Last: 40}}}},
			},
		},
		{
			name: "variadic range unions its spans",
			text: "rows(range(20:31, 40:40))",
			want: DirectiveValue{
				Axis: AxisRow,
				Items: []Item{
					{Kind: ItemRange, Spans: []Span{{First: 20, Last: 31}, {First: 40, Last: 40}}},
				},
			},
		},
		{
			name: "count of the first rows",
			text: "rows(count(3))",
			want: DirectiveValue{
				Axis:  AxisRow,
				Items: []Item{{Kind: ItemCount, Count: 3}},
			},
		},
		{
			name: "count from the end",
			text: "rows(count(-1))",
			want: DirectiveValue{
				Axis:  AxisRow,
				Items: []Item{{Kind: ItemCount, Count: -1}},
			},
		},
		{
			name: "range to the last row",
			text: "rows(range(20:-1))",
			want: DirectiveValue{
				Axis:  AxisRow,
				Items: []Item{{Kind: ItemRange, Spans: []Span{{First: 20, Last: -1}}}},
			},
		},
		{
			name: "the last row alone",
			text: "rows(range(-1:-1))",
			want: DirectiveValue{
				Axis:  AxisRow,
				Items: []Item{{Kind: ItemRange, Spans: []Span{{First: -1, Last: -1}}}},
			},
		},
		{
			name: "mixed items in one list",
			text: "rows(count(2), count(-1))",
			want: DirectiveValue{
				Axis: AxisRow,
				Items: []Item{
					{Kind: ItemCount, Count: 2},
					{Kind: ItemCount, Count: -1},
				},
			},
		},
		{
			name: "columns to the last are numeric from the end",
			text: "cols(range(B:-1))",
			want: DirectiveValue{
				Axis:  AxisCol,
				Items: []Item{{Kind: ItemRange, Spans: []Span{{First: 2, Last: -1}}}},
			},
		},
		{
			name: "multi-letter columns",
			text: "cols(range(Z:AA))",
			want: DirectiveValue{
				Axis:  AxisCol,
				Items: []Item{{Kind: ItemRange, Spans: []Span{{First: 26, Last: 27}}}},
			},
		},
		{
			name: "padding is insignificant",
			text: "rows( range( 20 : 31 ) , count( 3 ) )",
			want: DirectiveValue{
				Axis: AxisRow,
				Items: []Item{
					{Kind: ItemRange, Spans: []Span{{First: 20, Last: 31}}},
					{Kind: ItemCount, Count: 3},
				},
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseDirectiveValue(c.text)
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
		text ValueText
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
			_, err := ParseDirectiveValue(c.text)
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
		text ValueText
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
			_, err := ParseDirectiveValue(c.text)
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

	for _, text := range []ValueText{
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
			_, err := ParseDirectiveValue(text)
			require.Error(t, err)
			assert.True(t, errors.Is(err, constants.ErrInvalidValue), "got %v", err)
		})
	}
}

// TestItemKind_RangeAndCountAreNeverInterchangeable pins the distinction the
// ItemKind doc rests on: a range is absolute and shifts under a structural
// edit, a count is anchored to an edge and re-resolves against it. If the two
// collapsed, `count(3)` would become a spelling of `range(1:3)` and a header
// declaration would silently stop tracking the top of the grid.
func TestItemKind_RangeAndCountAreNeverInterchangeable(t *testing.T) {
	t.Parallel()

	spanning, err := ParseDirectiveValue("rows(range(1:3))")
	require.NoError(t, err)
	counting, err := ParseDirectiveValue("rows(count(3))")
	require.NoError(t, err)

	require.Len(t, spanning.Items, 1)
	require.Len(t, counting.Items, 1)
	assert.Equal(t, ItemRange, spanning.Items[0].Kind)
	assert.Equal(t, ItemCount, counting.Items[0].Kind)
	assert.NotEqual(t, spanning.Items[0].Kind, counting.Items[0].Kind)

	// The behavioural half of "never interchangeable". Both forms select rows
	// 1-3 before the edit; after one insertion at the top they no longer agree,
	// because the absolute span moves with its rows while the edge-anchored
	// count keeps covering the top of the grid.
	edit := Insertion{Axis: AxisRow, At: 1, Size: 10}
	assert.Equal(t, ValueText("rows(range(2:4))"), spanning.Insert(edit).Render())
	assert.Equal(t, ValueText("rows(count(4))"), counting.Insert(edit).Render())
}

// TestItemRange_IsNeverADefaultSpelling pins why both forms are always written
// out: a bare number is refused rather than guessed at, so it can never be read
// as the wrong one of the two.
func TestItemRange_IsNeverADefaultSpelling(t *testing.T) {
	t.Parallel()

	for _, bare := range []ValueText{"rows(3)", "cols(2)", "rows(1:3)"} {
		_, err := ParseDirectiveValue(bare)
		assert.Error(t, err, string(bare)+" must not be guessed into a range or a count")
	}
}

// TestAdviseSyntax_NamesTheSpellingTheLanguageWanted pins the claim that the
// two refusals a reader hits most carry their fix. A syntax error that only
// says "syntax error" leaves an author guessing at a grammar they cannot see.
func TestAdviseSyntax_NamesTheSpellingTheLanguageWanted(t *testing.T) {
	t.Parallel()

	cases := map[ValueText]string{
		"rows(3)":            "count(3)",
		"rows(range(20,31))": "20:31",
	}
	for value, wanted := range cases {
		_, err := ParseDirectiveValue(value)
		require.Error(t, err, string(value))
		assert.Contains(t, err.Error(), wanted, "the refusal names the spelling that works")
	}
}
