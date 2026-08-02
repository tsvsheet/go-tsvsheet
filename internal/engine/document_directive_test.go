package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

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

// TestShiftDirectiveLineLeavesAnUnreadableLineExactlyAsWritten pins the rule
// that an edit is no moment to discard someone's text. A line the parser cannot
// read still belongs to its author, and rewriting or dropping it would destroy
// work in the middle of an unrelated operation.
func TestShiftDirectiveLineLeavesAnUnreadableLineExactlyAsWritten(t *testing.T) {
	t.Parallel()
	const unreadable = "#.hide\tnonsense(\n"
	doc, err := engine.ParseDocument([]byte(unreadable + "a\nb\n"))
	require.NoError(t, err)

	edited := doc.InsertRow(engine.Address{})

	assert.Contains(t, string(edited.Text()), unreadable, "the line survives the edit verbatim")
}

// TestShiftDirectiveFieldKeepsAnUntouchedLineByteIdentical pins that an edit
// rewrites only what it moved. Re-rendering an untouched directive into
// canonical form would put an unrelated diff in front of a reviewer and lose
// whichever equivalent spelling its author chose.
func TestShiftDirectiveFieldKeepsAnUntouchedLineByteIdentical(t *testing.T) {
	t.Parallel()
	const untouched = "#.header\tcols(count(1))\n"
	doc, err := engine.ParseDocument([]byte(untouched + "a\tb\nc\td\n"))
	require.NoError(t, err)

	edited := doc.InsertRow(engine.Address{Row: 1})

	assert.Contains(t, string(edited.Text()), untouched,
		"a column directive is untouched by a row insert, byte for byte")
}
