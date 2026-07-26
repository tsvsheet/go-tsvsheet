package tsvt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// mustParse parses a directive value a test states as source.
func mustParse(t *testing.T, text tsvt.ValueText) tsvt.DirectiveValue {
	t.Helper()
	v, err := tsvt.ParseDirectiveValue(text)
	require.NoError(t, err)
	return v
}

// TestRenderRoundTrips proves a canonically written value comes back out
// byte-identical, so a sheet the engine has not edited keeps its own spelling.
// Grouping is part of that: `range(20:31, 40:40)` is one call of two spans and
// must not come back as two calls.
func TestRenderRoundTrips(t *testing.T) {
	t.Parallel()

	for _, text := range []tsvt.ValueText{
		"rows(range(20:31))",
		"rows(range(20:31, 40:40))",
		"rows(count(3))",
		"rows(count(-1))",
		"rows(range(20:-1))",
		"rows(count(2), count(-1))",
		"rows(range(1:3), count(2))",
		"cols(range(B:M))",
		"cols(range(A:A, Z:AA))",
		"cols(count(-1))",
		"cols(range(B:-1))",
	} {
		t.Run(string(text), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, text, mustParse(t, text).Render())
		})
	}
}

// TestRenderNormalisesSpacingOnly proves the one thing a rewrite may change
// about spelling: incidental padding. Everything structural — order, grouping,
// which form each item takes — survives untouched.
func TestRenderNormalisesSpacingOnly(t *testing.T) {
	t.Parallel()

	got := mustParse(t, "rows( range( 20 : 31 ) , count( 3 ) )").Render()
	assert.Equal(t, tsvt.ValueText("rows(range(20:31), count(3))"), got)
}

// TestInsertShiftsAbsoluteOnly is the heart of the anchoring-class rule: an
// insertion moves the spans that name fixed places and leaves alone everything
// anchored to an edge. It is why count(3) is not a spelling of range(1:3), and
// why a lint pass may not swap one for the other.
func TestInsertShiftsAbsoluteOnly(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		from tsvt.ValueText
		want tsvt.ValueText
		at   tsvt.Offset
	}{
		{name: "span below the insertion moves", from: "rows(range(5:9))", at: 1, want: "rows(range(6:10))"},
		{name: "span above it stands still", from: "rows(range(5:9))", at: 20, want: "rows(range(5:9))"},
		{name: "an insertion inside a span widens it", from: "rows(range(3:7))", at: 5, want: "rows(range(3:8))"},
		{name: "an insertion at the start pushes it down", from: "rows(range(3:7))", at: 3, want: "rows(range(4:8))"},
		{name: "a count never moves", from: "rows(count(3))", at: 1, want: "rows(count(3))"},
		{name: "a from-the-end count never moves", from: "rows(count(-1))", at: 1, want: "rows(count(-1))"},
		{name: "a from-the-end endpoint never moves", from: "rows(range(5:-1))", at: 1, want: "rows(range(6:-1))"},
		{name: "one list, both behaviours", from: "rows(count(3), range(5:9))", at: 1, want: "rows(count(3), range(6:10))"},
		{name: "columns shift on their own axis", from: "cols(range(B:D))", at: 1, want: "cols(range(C:E))"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			v := mustParse(t, c.from)
			got := v.Insert(tsvt.Insertion{Axis: v.Axis, At: c.at})
			assert.Equal(t, c.want, got.Render())
		})
	}
}

// TestInsertIgnoresTheOtherAxis proves a row insert leaves a column directive
// alone and the reverse, since the axes are independent.
func TestInsertIgnoresTheOtherAxis(t *testing.T) {
	t.Parallel()

	cols := mustParse(t, "cols(range(B:D))")
	assert.Equal(t, tsvt.ValueText("cols(range(B:D))"),
		cols.Insert(tsvt.Insertion{Axis: tsvt.AxisRow, At: 1}).Render())

	rows := mustParse(t, "rows(range(2:4))")
	assert.Equal(t, tsvt.ValueText("rows(range(2:4))"),
		rows.Delete(tsvt.Deletion{Axis: tsvt.AxisCol, At: 1}).Render())
}

// TestDeleteShiftsAndDrops covers removal, where the two endpoints move by
// different rules: a span starting at the removed position starts at whatever
// moved up into it, a span ending there ends one earlier, and a span that named
// only the removed position is dropped rather than left backwards.
func TestDeleteShiftsAndDrops(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		from tsvt.ValueText
		want tsvt.ValueText
		at   tsvt.Offset
	}{
		{name: "span below the deletion moves up", from: "rows(range(5:9))", at: 1, want: "rows(range(4:8))"},
		{name: "span above it stands still", from: "rows(range(5:9))", at: 20, want: "rows(range(5:9))"},
		{name: "a deletion inside a span narrows it", from: "rows(range(3:7))", at: 5, want: "rows(range(3:6))"},
		{name: "deleting the last row of a span narrows it", from: "rows(range(3:7))", at: 7, want: "rows(range(3:6))"},
		{name: "deleting the first row of a span narrows it", from: "rows(range(3:7))", at: 3, want: "rows(range(3:6))"},
		{name: "a count never moves", from: "rows(count(3))", at: 1, want: "rows(count(3))"},
		{name: "columns delete on their own axis", from: "cols(range(C:E))", at: 1, want: "cols(range(B:D))"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			v := mustParse(t, c.from)
			got := v.Delete(tsvt.Deletion{Axis: v.Axis, At: c.at})
			assert.Equal(t, c.want, got.Render())
		})
	}
}

// TestDeleteDropsAnEmptiedSpan covers the case that cannot be written down: a
// one-row span whose row is removed. The span goes, its neighbours stay, and an
// item left with no spans goes with it.
func TestDeleteDropsAnEmptiedSpan(t *testing.T) {
	t.Parallel()

	both := mustParse(t, "rows(range(5:5, 8:9))").Delete(tsvt.Deletion{Axis: tsvt.AxisRow, At: 5})
	assert.Equal(t, tsvt.ValueText("rows(range(7:8))"), both.Render(),
		"the emptied span goes; the one beside it moves up")

	only := mustParse(t, "rows(range(5:5), count(2))").Delete(tsvt.Deletion{Axis: tsvt.AxisRow, At: 5})
	assert.Equal(t, tsvt.ValueText("rows(count(2))"), only.Render(),
		"an item left with no spans goes too, and the count is untouched")
}

// TestEditsPreserveGroupingAndOrder proves the rule a rewrite must not break:
// endpoints move, and nothing else does. Overlapping spans are not merged and
// items are not reordered, because canonical means one spelling per value, not
// a normalised set.
func TestEditsPreserveGroupingAndOrder(t *testing.T) {
	t.Parallel()

	v := mustParse(t, "rows(range(2:3, 3:5), count(1))")
	got := v.Insert(tsvt.Insertion{Axis: tsvt.AxisRow, At: 1})
	assert.Equal(t, tsvt.ValueText("rows(range(3:4, 4:6), count(1))"), got.Render())
}
