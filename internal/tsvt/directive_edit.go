// Rendering and shifting a view-directive value: the two operations a
// structural edit needs (SPECIFICATION §3, work order 009).
//
// A rewrite shifts endpoints and NOTHING else. Items keep their order and
// their grouping — `range(2:3, 3:5)` stays one call of two spans rather than
// collapsing to `range(2:5)` — because canonical means one spelling per value,
// not a normalised set: merging would rewrite an authorial choice and make the
// diff say more than the edit did.
package tsvt

import (
	"strconv"
	"strings"
)

// axisNames spells each axis back out, so a rewritten line reads exactly as an
// author would have written it.
var axisNames = map[Axis]funcName{AxisRow: fnRows, AxisCol: fnCols}

// Render writes a directive value back as source text. Parsing and rendering
// round-trip: a value the engine has not edited comes out byte-identical to a
// canonically written original.
func (v DirectiveValue) Render() ValueText {
	items := make([]string, 0, len(v.Items))
	for _, item := range v.Items {
		items = append(items, item.render(v.Axis))
	}
	return ValueText(string(axisNames[v.Axis]) + "(" + strings.Join(items, ", ") + ")")
}

// render writes one item — a count, or a range with every span it was written
// with, in the order it was written.
func (i Item) render(axis Axis) string {
	if i.Kind == ItemCount {
		return string(fnCount) + "(" + i.Count.render(axis) + ")"
	}
	spans := make([]string, 0, len(i.Spans))
	for _, s := range i.Spans {
		spans = append(spans, s.First.render(axis)+":"+s.Last.render(axis))
	}
	return string(fnRange) + "(" + strings.Join(spans, ", ") + ")"
}

// render writes one offset: a column letter on the column axis, a number on
// the row axis, and a negative number on either — a from-the-end offset is
// axis-neutral, since both axes have an end.
func (o Offset) render(axis Axis) string {
	if o < 0 || axis == AxisRow {
		return strconv.Itoa(int(o))
	}
	return columnLetters(o)
}

// columnLetters is the inverse of columnOffset: 1 → A, 27 → AA.
func columnLetters(o Offset) string {
	var out []byte
	for n := int(o); n > 0; n = (n - 1) / 26 {
		out = append([]byte{byte('A' + (n-1)%26)}, out...)
	}
	return string(out)
}

// Insertion is a structural edit that adds one row or column before a 1-based
// position, on one axis. Size is the axis length before the edit, which a
// count anchored to the far end needs in order to know whether the insertion
// fell inside it.
type Insertion struct {
	At   Offset
	Size Offset
	Axis Axis
}

// Deletion is a structural edit that removes the row or column at a 1-based
// position, on one axis. Size is the axis length before the edit, for the same
// reason as Insertion.
type Deletion struct {
	At   Offset
	Size Offset
	Axis Axis
}

// Insert returns the value as it stands after a row or column is inserted.
// Absolute endpoints at or after the insertion move down; edge-anchored ones —
// every count, and any negative offset — are untouched, because they
// re-resolve against the grid rather than naming a fixed place. That divergence
// is exactly why `count(3)` is not a spelling of `range(1:3)`.
func (v DirectiveValue) Insert(edit Insertion) DirectiveValue {
	if edit.Axis != v.Axis {
		return v
	}
	return v.mapItems(
		func(s Span) (Span, bool) {
			return Span{First: shiftFor(s.First, edit.At, 1), Last: shiftFor(s.Last, edit.At, 1)}, true
		},
		func(n Offset) (Offset, bool) { return growCount(n, edit.At, edit.Size, 1) },
	)
}

// Delete returns the value as it stands after a row or column is removed. A
// span that named only the removed position is dropped: it described something
// the grid no longer has, and keeping it would leave a backwards span that
// cannot be written down.
func (v DirectiveValue) Delete(edit Deletion) DirectiveValue {
	if edit.Axis != v.Axis {
		return v
	}
	return v.mapItems(
		func(s Span) (Span, bool) {
			// The two endpoints move by different rules: a span starting AT the
			// removed position now starts at whatever moved up into it, while a
			// span ending there ends one earlier. When that empties the span it
			// named only the removed position, so it goes.
			first, last := deleteFirst(s.First, edit.At), deleteLast(s.Last, edit.At)
			if first > 0 && last > 0 && last < first {
				return Span{}, false
			}
			return Span{First: first, Last: last}, true
		},
		func(n Offset) (Offset, bool) { return growCount(n, edit.At, edit.Size, -1) },
	)
}

// mapItems rebuilds a value with every item transformed — spans by one rule,
// counts by another — dropping whatever no longer describes anything. Item
// order and grouping are preserved throughout.
func (v DirectiveValue) mapItems(span func(Span) (Span, bool), count func(Offset) (Offset, bool)) DirectiveValue {
	items := make([]Item, 0, len(v.Items))
	for _, item := range v.Items {
		if moved, keep := item.mapItem(span, count); keep {
			items = append(items, moved)
		}
	}
	return DirectiveValue{Axis: v.Axis, Items: items}
}

// mapItem transforms one item, reporting false when nothing survives: a range
// whose every span named only removed positions, or a count reduced to nothing.
func (i Item) mapItem(span func(Span) (Span, bool), count func(Offset) (Offset, bool)) (Item, bool) {
	if i.Kind == ItemCount {
		n, keep := count(i.Count)
		return Item{Kind: ItemCount, Count: n}, keep
	}
	spans := make([]Span, 0, len(i.Spans))
	for _, s := range i.Spans {
		if moved, keep := span(s); keep {
			spans = append(spans, moved)
		}
	}
	return Item{Kind: ItemRange, Spans: spans}, len(spans) > 0
}

// growCount widens or narrows a count when the edit fell INSIDE the block it
// names — inserting a row within a three-row header makes it a four-row header,
// and deleting one makes it two. An edit outside the block leaves it alone,
// since a count re-resolves against its edge either way. A count reduced to
// nothing describes nothing and is dropped.
func growCount(n, at, size, by Offset) (Offset, bool) {
	if !countContains(n, at, size) {
		return n, true
	}
	if n > 0 {
		return n + by, n+by > 0
	}
	return n - by, n-by < 0
}

// countContains reports whether an edit position falls inside the block a count
// names: the first n positions, or the last -n of them.
func countContains(n, at, size Offset) bool {
	if n > 0 {
		return at <= n
	}
	return at >= size+n+1
}

// shiftFor moves an absolute endpoint that sits at or after an insertion,
// leaving from-the-end offsets alone.
func shiftFor(o, at, by Offset) Offset {
	if o < 0 || o < at {
		return o
	}
	return o + by
}

// deleteFirst moves a span's start after a removal. A start after the removed
// position moves back; a start AT it stays, because the next row has moved up
// into that place. From-the-end offsets are untouched.
func deleteFirst(o, at Offset) Offset {
	if o < 0 || o <= at {
		return o
	}
	return o - 1
}

// deleteLast moves a span's end after a removal. An end at or after the removed
// position moves back, since the span lost a member.
func deleteLast(o, at Offset) Offset {
	if o < 0 || o < at {
		return o
	}
	return o - 1
}
