package tsvt_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// TestParseRowSelector covers the row half of a view directive's value: a
// comma-separated list of 1-based inclusive spans, where a bare number is a
// span of one. Rows and columns are separate entry rules precisely so a key
// that wants rows rejects a value naming columns rather than accepting it and
// selecting nothing.
func TestParseRowSelector(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		text tsvt.SelectorText
		want []tsvt.RowSpan
	}{
		{"single row", "3", []tsvt.RowSpan{{First: 3, Last: 3}}},
		{"one span", "20-31", []tsvt.RowSpan{{First: 20, Last: 31}}},
		{"span and singleton", "20-31,40", []tsvt.RowSpan{{First: 20, Last: 31}, {First: 40, Last: 40}}},
		{"degenerate span", "7-7", []tsvt.RowSpan{{First: 7, Last: 7}}},
		{"padding is insignificant", "20 - 31 , 40", []tsvt.RowSpan{{First: 20, Last: 31}, {First: 40, Last: 40}}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := tsvt.ParseRowSelector(c.text)
			require.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}
}

// TestParseColSelector covers the column half: spans of A1 column letters,
// separated by `-` rather than `:` so that a whole-column form can never reach
// the expression language, which has no such reference (SPECIFICATION §4).
func TestParseColSelector(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		text tsvt.SelectorText
		want []tsvt.ColSpan
	}{
		{"single column", "N", []tsvt.ColSpan{{First: "N", Last: "N"}}},
		{"one span", "B-M", []tsvt.ColSpan{{First: "B", Last: "M"}}},
		{"list", "A,C-D", []tsvt.ColSpan{{First: "A", Last: "A"}, {First: "C", Last: "D"}}},
		{"multi-letter columns", "AA-AC", []tsvt.ColSpan{{First: "AA", Last: "AC"}}},
		{"span widening past Z", "Z-AA", []tsvt.ColSpan{{First: "Z", Last: "AA"}}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := tsvt.ParseColSelector(c.text)
			require.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}
}

// TestParseCount covers the count-valued keys (header-rows, freeze-cols): a
// plain non-negative integer, zero meaning "none declared".
func TestParseCount(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		text tsvt.SelectorText
		want tsvt.Count
	}{
		{"one", "1", 1},
		{"zero declares none", "0", 0},
		{"multi-digit", "12", 12},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := tsvt.ParseCount(c.text)
			require.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}
}

// TestParseSelectorErrors pins every rejection to its specific sentinel:
// ErrSyntax when the text is not the shape the grammar admits, ErrInvalidValue
// when it parses but cannot mean anything (a zero row, a backwards span). The
// axis-mismatch cases are the reason the two selectors are separate rules.
func TestParseSelectorErrors(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		call func(tsvt.SelectorText) error
		want error
		text tsvt.SelectorText
		name string
	}{
		{name: "columns where rows belong", text: "B-M", want: constants.ErrSyntax, call: rowSelectorErr},
		{name: "rows where columns belong", text: "20", want: constants.ErrSyntax, call: colSelectorErr},
		{name: "empty row selector", text: "", want: constants.ErrSyntax, call: rowSelectorErr},
		{name: "empty column selector", text: "", want: constants.ErrSyntax, call: colSelectorErr},
		{name: "dangling span", text: "20-", want: constants.ErrSyntax, call: rowSelectorErr},
		{name: "trailing comma", text: "20,", want: constants.ErrSyntax, call: rowSelectorErr},
		{name: "range colon is not a selector", text: "B:M", want: constants.ErrSyntax, call: colSelectorErr},
		{name: "lowercase column", text: "b-m", want: constants.ErrSyntax, call: colSelectorErr},
		{name: "trailing text", text: "20-31 rows", want: constants.ErrSyntax, call: rowSelectorErr},
		{name: "row zero", text: "0", want: constants.ErrInvalidValue, call: rowSelectorErr},
		{name: "row zero inside a span", text: "0-3", want: constants.ErrInvalidValue, call: rowSelectorErr},
		{name: "backwards row span", text: "31-20", want: constants.ErrInvalidValue, call: rowSelectorErr},
		{name: "backwards column span", text: "M-B", want: constants.ErrInvalidValue, call: colSelectorErr},
		{name: "backwards across letter widths", text: "AA-Z", want: constants.ErrInvalidValue, call: colSelectorErr},
		{name: "fractional row", text: "1.5", want: constants.ErrInvalidValue, call: rowSelectorErr},
		{name: "fractional span endpoint", text: "1-2.5", want: constants.ErrInvalidValue, call: rowSelectorErr},
		{name: "fractional count", text: "1.5", want: constants.ErrInvalidValue, call: countErr},
		{name: "empty count", text: "", want: constants.ErrSyntax, call: countErr},
		{name: "non-numeric count", text: "one", want: constants.ErrSyntax, call: countErr},
		{name: "count with a tail", text: "1 row", want: constants.ErrSyntax, call: countErr},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := c.call(c.text)
			require.Error(t, err)
			assert.True(t, errors.Is(err, c.want), "got %v, want %v", err, c.want)
		})
	}
}

// rowSelectorErr, colSelectorErr and countErr adapt each parser to the single
// error-returning shape the table drives.
func rowSelectorErr(s tsvt.SelectorText) error { _, err := tsvt.ParseRowSelector(s); return err }

func colSelectorErr(s tsvt.SelectorText) error { _, err := tsvt.ParseColSelector(s); return err }

func countErr(s tsvt.SelectorText) error { _, err := tsvt.ParseCount(s); return err }
