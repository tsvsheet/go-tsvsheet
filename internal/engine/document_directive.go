// Directives on a document: reading the view a sheet declares, and keeping the
// declarations true through a structural edit (SPECIFICATION §3, work order
// 009).
//
// Only a directive line whose value actually moved is rewritten. Every other
// line — prose, a comment, a directive the edit did not touch — comes back
// byte-identical, because a rewrite must say exactly what the edit did and no
// more.
package engine

import (
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// Directives reads the view directives this document declares, with the
// diagnostics for any that cannot be read.
func (d Document) Directives() ([]Directive, []Diagnostic) {
	return DirectivesOf(d.sourceLines())
}

// View resolves this document's directives against its own extent: which rows
// and columns are hidden, carry headers, or stay anchored.
func (d Document) View() (View, []Diagnostic) {
	ds, diags := d.Directives()
	view, resolved := ResolveView(ds, d.Extent())
	return view, append(diags, resolved...)
}

// Extent reports the grid's size, which edge-anchored items resolve against.
func (d Document) Extent() Extent {
	rows := len(d.sheet.cells)
	cols := 0
	if rows > 0 {
		cols = len(d.sheet.cells[0])
	}
	return Extent{Rows: rows, Cols: cols}
}

// sourceLines lists the document's physical lines: comments verbatim, grid rows
// as their serialized text, in layout order — the same order a reader sees.
func (d Document) sourceLines() []SourceLine {
	rows := d.sheet.Source()
	lines := make([]SourceLine, 0, len(d.layout))
	next := rowIndex(0)
	for _, line := range d.layout {
		if !line.isRow {
			lines = append(lines, SourceLine(line.comment))
			continue
		}
		lines = append(lines, SourceLine(strings.Join(rows[next], tab)))
		next++
	}
	return lines
}

// shiftDirectives rewrites the document's directive lines through one
// structural edit, leaving every other line untouched.
func (d Document) shiftDirectives(edit directiveEdit) Document {
	layout := make([]docLine, len(d.layout))
	copy(layout, d.layout)
	for i, line := range layout {
		if line.isRow || !strings.HasPrefix(line.comment, directiveMarker) {
			continue
		}
		moved := shiftDirectiveLine(lineText(line.comment), edit)
		layout[i] = docLine{comment: string(moved)}
	}
	return Document{sheet: d.sheet, layout: layout}
}

// lineText is one physical line of a document, as written.
type lineText string

// directiveEdit is one structural edit expressed for a directive value: an
// insertion or a deletion of one position on one axis.
type directiveEdit struct {
	at        tsvt.Offset
	size      tsvt.Offset
	axis      tsvt.Axis
	isRemoval bool
}

// apply moves one value through the edit.
func (e directiveEdit) apply(v tsvt.DirectiveValue) tsvt.DirectiveValue {
	if e.isRemoval {
		return v.Delete(tsvt.Deletion{Axis: e.axis, At: e.at, Size: e.size})
	}
	return v.Insert(tsvt.Insertion{Axis: e.axis, At: e.at, Size: e.size})
}

// shiftDirectiveLine rewrites the values on one directive line. A line that
// cannot be read is left exactly as written — an edit is no moment to discard
// someone's text — and so is a value the edit did not move.
func shiftDirectiveLine(text lineText, edit directiveEdit) lineText {
	fields := strings.Split(string(text)[markerLen:], tab)
	if _, known := keyNames[strings.TrimSpace(fields[0])]; !known || len(fields)%2 != 0 {
		return text
	}
	moved := make([]string, len(fields))
	copy(moved, fields)
	for i := 1; i < len(fields); i += 2 {
		moved[i] = string(shiftDirectiveField(fieldText(fields[i]), edit))
	}
	return lineText(directiveMarker + strings.Join(moved, tab))
}

// shiftDirectiveField moves one value, keeping the original text when the value
// cannot be parsed or when the edit left it alone — so an untouched line stays
// byte-identical rather than being re-rendered into canonical form.
func shiftDirectiveField(text fieldText, edit directiveEdit) fieldText {
	parsed, err := tsvt.ParseDirectiveValue(tsvt.ValueText(strings.TrimSpace(string(text))))
	if err != nil {
		return text
	}
	shifted := edit.apply(parsed)
	if shifted.Render() == parsed.Render() {
		return text
	}
	return fieldText(shifted.Render())
}
