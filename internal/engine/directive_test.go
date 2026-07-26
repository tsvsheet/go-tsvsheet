package engine_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// linesOf splits a source sheet into the physical lines DirectivesOf reads.
func linesOf(src string) []engine.SourceLine {
	split := strings.Split(strings.TrimSuffix(src, "\n"), "\n")
	lines := make([]engine.SourceLine, 0, len(split))
	for _, l := range split {
		lines = append(lines, engine.SourceLine(l))
	}
	return lines
}

// keysOf lists the parsed keys in order, for asserting what a source yielded.
func keysOf(ds []engine.Directive) []engine.Key {
	keys := make([]engine.Key, 0, len(ds))
	for _, d := range ds {
		keys = append(keys, d.Key)
	}
	return keys
}

// TestDirectivesOfProseIsSilent covers the rule that keeps `#.` usable as an
// ordinary comment marker: a line is a directive only when its first field
// names a known key, so prose costs nothing and produces no diagnostic.
func TestDirectivesOfProseIsSilent(t *testing.T) {
	t.Parallel()

	src := "#!/usr/bin/env tsvsheet\n" +
		"#.a note about this sheet\n" +
		"#.see also\tthe sheet next door\n" +
		"#. hidden is not a key either\n" +
		"a\tb\n"

	found, diags := engine.DirectivesOf(linesOf(src))
	assert.Empty(t, found)
	assert.Empty(t, diags, "prose must never diagnose")
}

// TestDirectivesOfReadsKeys covers the three keys, both axes, packed pairs, and
// a directive interspersed with data — position carries no meaning, so where a
// line sits never changes what it declares.
func TestDirectivesOfReadsKeys(t *testing.T) {
	t.Parallel()

	src := "#.hide\tcols(range(B:M))\n" +
		"a\tb\n" +
		"#.header\trows(count(1))\n" +
		"#.freeze\trows(count(2), count(-1))\tfreeze\tcols(count(1))\n" +
		"c\td\n"

	found, diags := engine.DirectivesOf(linesOf(src))
	require.Empty(t, diags)
	assert.Equal(t,
		[]engine.Key{engine.KeyHide, engine.KeyHeader, engine.KeyFreeze, engine.KeyFreeze},
		keysOf(found))
	assert.Equal(t, engine.LineNumber(1), found[0].At)
	assert.Equal(t, engine.LineNumber(3), found[1].At, "an interspersed directive keeps its own line")
}

// TestDirectivesOfDiagnoses covers every way a line that *is* a directive can
// fail: an odd field count, an unknown key beside a known one, a value the
// language refuses, and a freeze pane that anchors to nothing. Each carries the
// line it came from, since a directive occupies a line and no row.
func TestDirectivesOfDiagnoses(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name  string
		src   string
		says  string
		keeps int
	}{
		{name: "key with no value", src: "#.hide\n", says: "pairs each key with a value"},
		{name: "odd field count", src: "#.hide\tcols(range(B:B))\theader\n", says: "pairs each key with a value"},
		{
			name: "unknown key beside a known one",
			src:  "#.hide\tcols(range(B:B))\thidden\trows(count(1))\n",
			says: "hide, header and freeze",
			// The readable pair survives: one broken pair must not discard the
			// declaration beside it.
			keeps: 1,
		},
		{name: "bare item", src: "#.hide\trows(3)\n", says: "range(3:3)"},
		{name: "wrong axis", src: "#.hide\trows(range(B:M))\n", says: "rows are numbered"},
		{name: "freeze anchored to nothing", src: "#.freeze\trows(range(5:7))\n", says: "anchors to an edge"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			found, diags := engine.DirectivesOf(linesOf(c.src))
			require.Len(t, diags, 1)
			assert.Contains(t, diags[0].Message, c.says)
			assert.Equal(t, 1, diags[0].Line)
			assert.Len(t, found, c.keeps, "a pair that cannot be read declares nothing; its neighbours still do")
		})
	}
}

// TestResolveView covers the whole resolution contract: counts anchor to an
// edge, ranges stand where written, negatives count back from the last, items
// union, and a span naming rows the grid does not have simply selects fewer.
func TestResolveView(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		get  func(engine.View) engine.Selection
		name string
		src  string
		want []int
		ext  engine.Extent
	}{
		{
			name: "count from the start", src: "#.header\trows(count(2))\n",
			ext: engine.Extent{Rows: 10, Cols: 5}, want: []int{1, 2},
			get: func(v engine.View) engine.Selection { return v.HeaderRows },
		},
		{
			name: "count from the end", src: "#.freeze\trows(count(-2))\n",
			ext: engine.Extent{Rows: 10, Cols: 5}, want: []int{9, 10},
			get: func(v engine.View) engine.Selection { return v.FreezeRows },
		},
		{
			name: "range stands where written", src: "#.hide\trows(range(3:5))\n",
			ext: engine.Extent{Rows: 10, Cols: 5}, want: []int{3, 4, 5},
			get: func(v engine.View) engine.Selection { return v.HiddenRows },
		},
		{
			name: "range to the last row", src: "#.hide\trows(range(8:-1))\n",
			ext: engine.Extent{Rows: 10, Cols: 5}, want: []int{8, 9, 10},
			get: func(v engine.View) engine.Selection { return v.HiddenRows },
		},
		{
			name: "columns resolve on their own axis", src: "#.hide\tcols(range(B:D))\n",
			ext: engine.Extent{Rows: 10, Cols: 5}, want: []int{2, 3, 4},
			get: func(v engine.View) engine.Selection { return v.HiddenCols },
		},
		{
			name: "items union, overlap and all", src: "#.hide\trows(count(3), range(2:5))\n",
			ext: engine.Extent{Rows: 10, Cols: 5}, want: []int{1, 2, 3, 4, 5},
			get: func(v engine.View) engine.Selection { return v.HiddenRows },
		},
		{
			name: "repetition unions across lines", src: "#.hide\trows(range(1:1))\n#.hide\trows(range(3:3))\n",
			ext: engine.Extent{Rows: 10, Cols: 5}, want: []int{1, 3},
			get: func(v engine.View) engine.Selection { return v.HiddenRows },
		},
		{
			name: "beyond the grid selects nothing", src: "#.hide\trows(range(40:40))\n",
			ext: engine.Extent{Rows: 10, Cols: 5}, want: nil,
			get: func(v engine.View) engine.Selection { return v.HiddenRows },
		},
		{
			name: "a count larger than the grid clamps", src: "#.header\trows(count(99))\n",
			ext: engine.Extent{Rows: 3, Cols: 5}, want: []int{1, 2, 3},
			get: func(v engine.View) engine.Selection { return v.HeaderRows },
		},
		{
			name: "freeze accepts a range anchored at the first row", src: "#.freeze\trows(range(1:2))\n",
			ext: engine.Extent{Rows: 10, Cols: 5}, want: []int{1, 2},
			get: func(v engine.View) engine.Selection { return v.FreezeRows },
		},
		{
			name: "freeze accepts a range anchored at the last row", src: "#.freeze\trows(range(-2:-1))\n",
			ext: engine.Extent{Rows: 10, Cols: 5}, want: []int{9, 10},
			get: func(v engine.View) engine.Selection { return v.FreezeRows },
		},
		{
			name: "freeze accepts a column range anchored at A", src: "#.freeze\tcols(range(A:B))\n",
			ext: engine.Extent{Rows: 10, Cols: 5}, want: []int{1, 2},
			get: func(v engine.View) engine.Selection { return v.FreezeCols },
		},
		{
			name: "a count from the end larger than the grid clamps", src: "#.freeze\trows(count(-99))\n",
			ext: engine.Extent{Rows: 3, Cols: 5}, want: []int{1, 2, 3},
			get: func(v engine.View) engine.Selection { return v.FreezeRows },
		},
		{
			name: "freeze pins both edges", src: "#.freeze\trows(count(2), count(-1))\n",
			ext: engine.Extent{Rows: 10, Cols: 5}, want: []int{1, 2, 10},
			get: func(v engine.View) engine.Selection { return v.FreezeRows },
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			found, diags := engine.DirectivesOf(linesOf(c.src))
			require.Empty(t, diags)
			view, resolveDiags := engine.ResolveView(found, c.ext)
			require.Empty(t, resolveDiags)
			assert.Equal(t, selectionOf(c.want), c.get(view))
		})
	}
}

// selectionOf builds the expected selection from a list of positions.
func selectionOf(positions []int) engine.Selection {
	sel := engine.Selection{}
	for _, p := range positions {
		sel[p] = true
	}
	return sel
}

// TestResolveViewBackwardsAfterSubstitution covers the one resolution failure:
// a span written forwards that runs backwards once its edge-anchored endpoint
// is substituted against a grid too small for it. It selects nothing and says
// so, rather than silently selecting a reversed span.
func TestResolveViewBackwardsAfterSubstitution(t *testing.T) {
	t.Parallel()

	found, diags := engine.DirectivesOf(linesOf("#.hide\trows(range(4:-3))\n"))
	require.Empty(t, diags)

	view, resolveDiags := engine.ResolveView(found, engine.Extent{Rows: 5, Cols: 2})
	require.Len(t, resolveDiags, 1)
	assert.Contains(t, resolveDiags[0].Message, "backwards")
	assert.Equal(t, 1, resolveDiags[0].Line)
	assert.Empty(t, view.HiddenRows)
}

// TestResolveViewIgnoresLegacyComments proves the legacy hash-space marker
// carries no directives: it is a comment form, never a directive form, so a
// sheet written before the marker changed cannot acquire a view by accident.
func TestResolveViewIgnoresLegacyComments(t *testing.T) {
	t.Parallel()

	found, diags := engine.DirectivesOf(linesOf("# hide\trows(count(1))\na\tb\n"))
	assert.Empty(t, found)
	assert.Empty(t, diags)
}

// TestDocumentViewReadsItsOwnExtent proves a document resolves its directives
// against its own grid, so a from-the-end item lands where the sheet actually
// ends rather than where a caller guesses it does.
func TestDocumentViewReadsItsOwnExtent(t *testing.T) {
	t.Parallel()

	doc, err := engine.ParseDocument([]byte(
		"#.header\trows(count(1))\n#.hide\trows(count(-1))\n#.freeze\tcols(count(1))\n" +
			"name\tqty\nwidget\t3\ngadget\t2\ntotal\t5\n",
	))
	require.NoError(t, err)

	view, diags := doc.View()
	require.Empty(t, diags)
	assert.Equal(t, engine.Extent{Rows: 4, Cols: 2}, doc.Extent())
	assert.Equal(t, engine.Selection{1: true}, view.HeaderRows)
	assert.Equal(t, engine.Selection{4: true}, view.HiddenRows, "the last row is the totals line")
	assert.Equal(t, engine.Selection{1: true}, view.FreezeCols)
}

// TestDocumentEditsShiftDirectives is the R5 contract for directives: an edit
// rewrites what moved and nothing else, so an untouched line — prose, a
// comment, a directive the edit missed — comes back byte-identical.
func TestDocumentEditsShiftDirectives(t *testing.T) {
	t.Parallel()

	const src = "#.a note that must survive verbatim\n" +
		"#.header\trows(count(1))\n" +
		"#.hide\trows(range(3:4))\n" +
		"#.hide\tcols(range(B:B))\n" +
		"name\tqty\nwidget\t3\ngadget\t2\ntotal\t5\n"

	doc, err := engine.ParseDocument([]byte(src))
	require.NoError(t, err)

	// A row inserted at the top: the header widens to take it in, the row span
	// moves down, the column directive and the prose are untouched.
	inserted := doc.InsertRow(engine.Address{Row: 0, Col: 0})
	assert.Equal(t,
		"#.a note that must survive verbatim\n"+
			"#.header\trows(count(2))\n"+
			"#.hide\trows(range(4:5))\n"+
			"#.hide\tcols(range(B:B))\n"+
			"\t\nname\tqty\nwidget\t3\ngadget\t2\ntotal\t5\n",
		string(inserted.Text()))

	// A column inserted at the front moves the column directive only.
	widened := doc.InsertCol(engine.Address{Row: 0, Col: 0})
	assert.Contains(t, string(widened.Text()), "#.hide\tcols(range(C:C))")
	assert.Contains(t, string(widened.Text()), "#.hide\trows(range(3:4))", "the row directive stands still")
}

// TestDocumentEditLeavesUnparseableDirectivesAlone proves an edit is no moment
// to discard someone's text: a directive that cannot be read is carried through
// verbatim, to be fixed by its author rather than silently rewritten.
func TestDocumentEditLeavesUnparseableDirectivesAlone(t *testing.T) {
	t.Parallel()

	const src = "#.hide\trows(3)\n#.hide\trows(range(3:4))\na\tb\nc\td\n"
	doc, err := engine.ParseDocument([]byte(src))
	require.NoError(t, err)

	got := string(doc.InsertRow(engine.Address{Row: 0, Col: 0}).Text())
	assert.Contains(t, got, "#.hide\trows(3)\n", "the unreadable line is untouched")
	assert.Contains(t, got, "#.hide\trows(range(4:5))\n", "its readable neighbour still moves")
}

// TestDocumentRoundTripsUneditedDirectives proves the byte-exactness rule:
// parsing and re-serializing a sheet full of directives changes nothing, so a
// file only ever diffs when an edit actually moved something.
func TestDocumentRoundTripsUneditedDirectives(t *testing.T) {
	t.Parallel()

	const src = "#!/usr/bin/env tsvsheet\n" +
		"#.hide\tcols(range(B:M))\theader\trows(count(1))\n" +
		"#.freeze\trows( count(2) , count(-1) )\n" +
		"#.prose that is not a directive\n" +
		"a\tb\n"

	doc, err := engine.ParseDocument([]byte(src))
	require.NoError(t, err)
	assert.Equal(t, src, string(doc.Text()), "an unedited document keeps every byte, padding included")
}

// TestDocumentDeletesShiftDirectives is the removal half of the R5 contract: a
// deleted row narrows the blocks that contained it and moves the ones below,
// and a deleted column does the same on its own axis.
func TestDocumentDeletesShiftDirectives(t *testing.T) {
	t.Parallel()

	const src = "#.header\trows(count(2))\n" +
		"#.hide\trows(range(4:5))\n" +
		"#.hide\tcols(range(C:D))\n" +
		"a\tb\tc\td\ne\tf\tg\th\ni\tj\tk\tl\nm\tn\to\tp\nq\tr\ts\tt\n"

	doc, err := engine.ParseDocument([]byte(src))
	require.NoError(t, err)

	// Deleting row 1 is inside the two-row header, so the header narrows; the
	// hidden span sits below and moves up.
	shorter := string(doc.DeleteRow(engine.Address{Row: 0, Col: 0}).Text())
	assert.Contains(t, shorter, "#.header\trows(count(1))")
	assert.Contains(t, shorter, "#.hide\trows(range(3:4))")
	assert.Contains(t, shorter, "#.hide\tcols(range(C:D))", "the column directive is untouched")

	// Deleting column A moves the hidden columns left; the row directives stand.
	narrower := string(doc.DeleteCol(engine.Address{Row: 0, Col: 0}).Text())
	assert.Contains(t, narrower, "#.hide\tcols(range(B:C))")
	assert.Contains(t, narrower, "#.header\trows(count(2))", "the row directive is untouched")
}
