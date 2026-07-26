// View directives: the `#.` lines that declare what a viewport does with the
// grid — which rows and columns are hidden, which carry headers, which stay
// anchored while the rest scrolls (SPECIFICATION §3).
//
// A directive is a projection of the grid, never an input to it: nothing here
// is reachable from a formula, and deleting every directive leaves the data
// byte-identical and computing identically. The line is a TSV row, so a key
// and its value are ordinary fields; only the value is a language, parsed by
// internal/tsvt against the shared grammar.
package engine

import (
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// Key is a view-directive key: which class of view state a value configures.
type Key int

// The three keys, one per class — a projection, a structure declaration, and a
// viewport hint. A grid has exactly two axes, and the axis lives in the value.
const (
	KeyHide Key = iota
	KeyHeader
	KeyFreeze
)

// keyNames is the whole key vocabulary. A first field outside it is prose.
var keyNames = map[string]Key{"hide": KeyHide, "header": KeyHeader, "freeze": KeyFreeze}

// markerLen is the length of the `#.` opening a directive or comment line; the
// rest of the line is ordinary TSV fields.
const markerLen = 2

// fieldText is one TAB-delimited field of a directive line, as written.
type fieldText string

// Directive is one key/value pair read from a `#.` line, carrying the physical
// line it came from so a diagnostic can name it.
type Directive struct {
	Value tsvt.DirectiveValue
	Key   Key
	At    LineNumber
}

// Extent is the grid's size. Edge-anchored items — `count(…)` and a `-1`
// endpoint — resolve against it, which is why a view is derived rather than
// stored.
type Extent struct {
	Rows int
	Cols int
}

// axisSize is how many rows or columns the grid has on the axis being
// resolved: the extent an edge-anchored item counts back from.
type axisSize int

// position is a resolved 1-based place on an axis, after every edge-anchored
// endpoint has been substituted against the extent.
type position int

// Selection is a set of 1-based positions on one axis: the rows a directive
// hides, the columns it freezes.
type Selection map[int]bool

// View is what a viewport does with the grid: sets of positions, never counts,
// so a host renders it without re-deriving anything.
type View struct {
	HiddenRows Selection
	HiddenCols Selection
	HeaderRows Selection
	HeaderCols Selection
	FreezeRows Selection
	FreezeCols Selection
}

// DirectivesOf reads the directives from a document's physical lines. A line
// is a directive only when its first field names a known key; anything else is
// prose, ignored in silence, so `#.a note` costs nothing while `#.hide` alone
// is reported.
func DirectivesOf(lines []SourceLine) ([]Directive, []Diagnostic) {
	var (
		found []Directive
		diags []Diagnostic
	)
	for i, line := range lines {
		at := LineNumber(i + 1)
		if !strings.HasPrefix(string(line), directiveMarker) {
			continue
		}
		lineFound, lineDiags := directivesOfLine(line, at)
		found = append(found, lineFound...)
		diags = append(diags, lineDiags...)
	}
	return found, diags
}

// directivesOfLine splits one `#.` line into its key/value pairs. Packed pairs
// are accepted — `#.freeze<TAB>rows(count(1))<TAB>header<TAB>rows(count(1))` —
// so reading is strict alternation over the fields.
func directivesOfLine(line SourceLine, at LineNumber) ([]Directive, []Diagnostic) {
	fields := strings.Split(string(line)[markerLen:], tab)
	if _, known := keyNames[strings.TrimSpace(fields[0])]; !known {
		return nil, nil
	}
	if len(fields)%2 != 0 {
		return nil, []Diagnostic{lineDiag(at, findingText("a directive pairs each key with a value: "+
			"#."+strings.TrimSpace(fields[0])+"<TAB>rows(count(1))"))}
	}
	return pairsOfLine(fields, at)
}

// pairsOfLine converts each key/value pair, collecting a diagnostic per pair
// that cannot be read rather than abandoning the whole line.
func pairsOfLine(fields []string, at LineNumber) ([]Directive, []Diagnostic) {
	var (
		found []Directive
		diags []Diagnostic
	)
	for i := 0; i+1 < len(fields); i += 2 {
		d, err := directiveOf(fieldText(fields[i]), fieldText(fields[i+1]), at)
		if err != nil {
			diags = append(diags, lineDiag(at, findingText(err.Error())))
			continue
		}
		found = append(found, d)
	}
	return found, diags
}

// directiveOf reads one key/value pair. An unknown key on a line that already
// opened with a known one is a mistake worth naming, not prose.
func directiveOf(key, value fieldText, at LineNumber) (Directive, error) {
	k, known := keyNames[strings.TrimSpace(string(key))]
	if !known {
		return Directive{}, constants.ErrInvalidValue.With(nil,
			"key", strings.TrimSpace(string(key)), "hint", "the keys are hide, header and freeze")
	}
	parsed, err := tsvt.ParseDirectiveValue(tsvt.ValueText(strings.TrimSpace(string(value))))
	if err != nil {
		return Directive{}, err
	}
	if k == KeyFreeze {
		if err := freezeTouchesEdge(parsed); err != nil {
			return Directive{}, err
		}
	}
	return Directive{Key: k, Value: parsed, At: at}, nil
}

// freezeTouchesEdge enforces the one constraint a viewport imposes: a pane is
// anchored to an edge, so every freeze item must reach one. A count always
// does; a range must start at the first position or end at the last.
func freezeTouchesEdge(v tsvt.DirectiveValue) error {
	for _, item := range v.Items {
		if item.Kind == tsvt.ItemCount || item.First == 1 || item.Last == -1 {
			continue
		}
		return constants.ErrInvalidValue.With(nil, "hint",
			"a freeze pane anchors to an edge: count(2), count(-1), range(1:2) or range(-2:-1)")
	}
	return nil
}

// ResolveView derives the view from directives and the grid's extent. Items
// union; an item naming positions the grid does not have selects nothing,
// because ambiguity in a view resolves toward showing more data, never less.
func ResolveView(ds []Directive, ext Extent) (View, []Diagnostic) {
	view := newView()
	var diags []Diagnostic
	for _, d := range ds {
		into, size := view.target(d, ext)
		for _, item := range d.Value.Items {
			first, last, isForwards := resolveItem(item, size)
			if !isForwards {
				diags = append(diags, lineDiag(d.At, "this selects nothing: it runs backwards once resolved"))
				continue
			}
			selectSpan(into, first, last, size)
		}
	}
	return view, diags
}

// newView allocates every selection so a caller never distinguishes "no
// directive" from "nothing selected".
func newView() View {
	return View{
		HiddenRows: Selection{}, HiddenCols: Selection{},
		HeaderRows: Selection{}, HeaderCols: Selection{},
		FreezeRows: Selection{}, FreezeCols: Selection{},
	}
}

// target picks the selection a directive writes into and the extent its
// edge-anchored items resolve against.
func (v View) target(d Directive, ext Extent) (Selection, axisSize) {
	rows := d.Value.Axis == tsvt.AxisRow
	byKey := map[Key][2]Selection{
		KeyHide:   {v.HiddenRows, v.HiddenCols},
		KeyHeader: {v.HeaderRows, v.HeaderCols},
		KeyFreeze: {v.FreezeRows, v.FreezeCols},
	}
	if rows {
		return byKey[d.Key][0], axisSize(ext.Rows)
	}
	return byKey[d.Key][1], axisSize(ext.Cols)
}

// resolveItem turns one item into an inclusive 1-based span against the
// extent, reporting false when it runs backwards once the edge-anchored
// endpoints are substituted.
func resolveItem(item tsvt.Item, size axisSize) (first, last position, isForwards bool) {
	if item.Kind == tsvt.ItemCount {
		return resolveCount(item.First, size)
	}
	first, last = resolvePos(item.First, size), resolvePos(item.Last, size)
	return first, last, first <= last
}

// resolveCount turns `count(n)` into a span at the near edge for a positive n
// and at the far edge for a negative one.
func resolveCount(n tsvt.Offset, size axisSize) (first, last position, isForwards bool) {
	if n > 0 {
		return 1, position(n), true
	}
	return position(size) + position(n) + 1, position(size), true
}

// resolvePos substitutes an endpoint: positive positions stand as written,
// negative ones count back from the last.
func resolvePos(o tsvt.Offset, size axisSize) position {
	if o < 0 {
		return position(size) + position(o) + 1
	}
	return position(o)
}

// selectSpan adds a resolved span to a selection, clamped to the grid — a span
// naming rows the sheet does not have simply selects fewer.
func selectSpan(into Selection, first, last position, size axisSize) {
	if first < 1 {
		first = 1
	}
	if last > position(size) {
		last = position(size)
	}
	for i := first; i <= last; i++ {
		into[int(i)] = true
	}
}

// findingText is what a directive diagnostic says — always naming the spelling
// it wants, so the message teaches the language rather than only refusing it.
type findingText string

// lineDiag reports a finding about a directive line rather than a cell.
func lineDiag(at LineNumber, message findingText) Diagnostic {
	return Diagnostic{Line: int(at), Message: string(message)}
}
