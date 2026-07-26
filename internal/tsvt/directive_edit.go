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
// position, on one axis.
type Insertion struct {
	At   Offset
	Axis Axis
}

// Deletion is a structural edit that removes the row or column at a 1-based
// position, on one axis.
type Deletion struct {
	At   Offset
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
	return v.mapSpans(func(s Span) (Span, bool) {
		return Span{First: shiftFor(s.First, edit.At, 1), Last: shiftFor(s.Last, edit.At, 1)}, true
	})
}

// Delete returns the value as it stands after a row or column is removed. A
// span that named only the removed position is dropped: it described something
// the grid no longer has, and keeping it would leave a backwards span that
// cannot be written down.
func (v DirectiveValue) Delete(edit Deletion) DirectiveValue {
	if edit.Axis != v.Axis {
		return v
	}
	return v.mapSpans(func(s Span) (Span, bool) {
		// The two endpoints move by different rules: a span starting AT the
		// removed position now starts at whatever moved up into it, while a
		// span ending there ends one earlier. When that empties the span it
		// named only the removed position, so it goes.
		first, last := deleteFirst(s.First, edit.At), deleteLast(s.Last, edit.At)
		if first > 0 && last > 0 && last < first {
			return Span{}, false
		}
		return Span{First: first, Last: last}, true
	})
}

// mapSpans rebuilds a value with each span transformed, dropping the spans the
// transform rejects and any item left with none. Counts pass through untouched,
// and item order is preserved throughout.
func (v DirectiveValue) mapSpans(f func(Span) (Span, bool)) DirectiveValue {
	items := make([]Item, 0, len(v.Items))
	for _, item := range v.Items {
		if moved, keep := item.mapSpans(f); keep {
			items = append(items, moved)
		}
	}
	return DirectiveValue{Axis: v.Axis, Items: items}
}

// mapSpans transforms one item's spans, reporting false when nothing survives:
// a range whose every span named only removed positions has nothing left to
// say. A count is anchored to an edge, so it always survives untouched.
func (i Item) mapSpans(f func(Span) (Span, bool)) (Item, bool) {
	if i.Kind == ItemCount {
		return i, true
	}
	spans := make([]Span, 0, len(i.Spans))
	for _, s := range i.Spans {
		if moved, keep := f(s); keep {
			spans = append(spans, moved)
		}
	}
	return Item{Kind: ItemRange, Spans: spans}, len(spans) > 0
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
