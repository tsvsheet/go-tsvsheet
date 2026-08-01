package engine_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// parseEdits parses src as an edits document, failing the test on error.
func parseEdits(t *testing.T, src string) engine.Edits {
	t.Helper()
	edits, err := engine.ParseEdits([]byte(src))
	require.NoError(t, err)
	return edits
}

// applied parses doc and edits, applies, and returns the resulting text.
func applied(t *testing.T, doc, edits string) string {
	t.Helper()
	out, err := engine.Apply(parseDoc(t, doc), parseEdits(t, edits), engine.DefaultLimits())
	require.NoError(t, err)
	return string(out.Text())
}

func TestRevisionIsSHA256OfCanonicalText(t *testing.T) {
	doc := parseDoc(t, "a\tb\n1\t2\n")
	sum := sha256.Sum256(doc.Text())
	assert.Equal(t, engine.RevisionHex(hex.EncodeToString(sum[:])), engine.Revision(doc))
}

func TestRevisionDistinguishesDocuments(t *testing.T) {
	assert.NotEqual(t, engine.Revision(parseDoc(t, "a\n")), engine.Revision(parseDoc(t, "b\n")))
}

func TestRevisionEqualForByteEqualDocuments(t *testing.T) {
	assert.Equal(t, engine.Revision(parseDoc(t, "a\tb\n")), engine.Revision(parseDoc(t, "a\tb\n")))
}

func TestParseEditsRoundTripsVerbatim(t *testing.T) {
	src := "#.base\t9f2c\n#. a prose note\n# legacy comment\n\nsetCell\tB2\t=A1+1\ninsertRow\t2\n"
	assert.Equal(t, src, string(parseEdits(t, src).Text()))
}

func TestParseEditsBase(t *testing.T) {
	assert.Equal(t, engine.RevisionHex("9f2c"), parseEdits(t, "#.base\t9f2c\nsetCell\tA1\tx\n").Base())
}

func TestParseEditsNoBaseIsEmpty(t *testing.T) {
	assert.Equal(t, engine.RevisionHex(""), parseEdits(t, "setCell\tA1\tx\n").Base())
}

func TestParseEditsLenCountsOpsNotComments(t *testing.T) {
	assert.Equal(t, 2, parseEdits(t, "#.base\tff\nsetCell\tA1\tx\n#. note\ndeleteRow\t1\n").Len())
}

func TestParseEditsEmptyDocumentIsValid(t *testing.T) {
	assert.Equal(t, 0, parseEdits(t, "").Len())
}

func TestParseEditsRejections(t *testing.T) {
	for name, tc := range map[string]struct {
		want error
		src  string
	}{
		"unknown op":            {src: "nope\tA1\n", want: constants.ErrEditsOp},
		"setCell missing text":  {src: "setCell\tB2\n", want: constants.ErrEditsArity},
		"setCell no args":       {src: "setCell\n", want: constants.ErrEditsArity},
		"insertRow no args":     {src: "insertRow\n", want: constants.ErrEditsArity},
		"insertRow extra args":  {src: "insertRow\t2\t3\n", want: constants.ErrEditsArity},
		"setCell bad address":   {src: "setCell\tZZ\t5\n", want: constants.ErrEditsAddress},
		"insertRow non-number":  {src: "insertRow\tzero\n", want: constants.ErrEditsAddress},
		"insertRow zero":        {src: "insertRow\t0\n", want: constants.ErrEditsAddress},
		"insertRow negative":    {src: "insertRow\t-1\n", want: constants.ErrEditsAddress},
		"insertCol digits":      {src: "insertCol\t5\n", want: constants.ErrEditsAddress},
		"insertCol lowercase":   {src: "insertCol\tc\n", want: constants.ErrEditsAddress},
		"insertCol no args":     {src: "insertCol\n", want: constants.ErrEditsArity},
		"fill bad span":         {src: "fill\tB2\tnope\n", want: constants.ErrEditsAddress},
		"fill reversed colon":   {src: "fill\tB2\tB3:\n", want: constants.ErrEditsAddress},
		"fill bad source":       {src: "fill\tx\tB2:B3\n", want: constants.ErrEditsAddress},
		"fill no args":          {src: "fill\n", want: constants.ErrEditsArity},
		"paste bad base64":      {src: "paste\tC1\tA1\t!!!\n", want: constants.ErrEditsBlock},
		"paste bad target":      {src: "paste\tx\tA1\tOQ==\n", want: constants.ErrEditsAddress},
		"paste bad origin":      {src: "paste\tC1\tx\tOQ==\n", want: constants.ErrEditsAddress},
		"paste no args":         {src: "paste\n", want: constants.ErrEditsArity},
		"duplicateRow bad row":  {src: "duplicateRow\tB\n", want: constants.ErrEditsAddress},
		"duplicateCol bad col":  {src: "duplicateCol\t2\n", want: constants.ErrEditsAddress},
		"deleteCol empty":       {src: "deleteCol\t\n", want: constants.ErrEditsAddress},
		"tab-only line no name": {src: "\tA1\n", want: constants.ErrEditsOp},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := engine.ParseEdits([]byte(tc.src))
			require.Error(t, err)
			assert.ErrorIs(t, err, tc.want)
		})
	}
}

func TestParseEditsErrorNamesTheLine(t *testing.T) {
	_, err := engine.ParseEdits([]byte("setCell\tA1\tx\n#. fine\nnope\tA1\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "3")
}

func TestApplySetCellEditsTheCell(t *testing.T) {
	assert.Equal(t, "1\t2\n3\t9\n", applied(t, "1\t2\n3\t4\n", "setCell\tB2\t9\n"))
}

func TestApplySetCellStoresFormulaVerbatim(t *testing.T) {
	assert.Equal(t, "1\t2\n3\t=A1+1\n", applied(t, "1\t2\n3\t4\n", "setCell\tB2\t=A1+1\n"))
}

func TestApplySetCellEmptyTextClears(t *testing.T) {
	assert.Equal(t, "1\t2\n3\t\n", applied(t, "1\t2\n3\t4\n", "setCell\tB2\t\n"))
}

func TestApplyInsertRowRewritesReferences(t *testing.T) {
	// R5: inserting before row 2 shifts the reference A3 to A4.
	got := applied(t, "=A3\nx\n7\n", "insertRow\t2\n")
	assert.Contains(t, got, "=A4")
}

func TestApplyFoldsInOrder(t *testing.T) {
	assert.Equal(t, "b\n", applied(t, "x\n", "setCell\tA1\ta\nsetCell\tA1\tb\n"))
}

// TestApplyOpsMatchDocumentMethods pins the delegation contract: each wire op
// is exactly its Document method — same argument mapping, same bytes out.
func TestApplyOpsMatchDocumentMethods(t *testing.T) {
	const src = "1\t2\t3\n4\t5\t6\n=A1+B2\t8\t9\n"
	limits := engine.DefaultLimits()
	block := engine.ParseBlock("9\t8")
	for name, tc := range map[string]struct {
		want func(engine.Document) engine.Document
		edit string
	}{
		"insertRow\t2\n":    {edit: "insertRow\t2\n", want: func(d engine.Document) engine.Document { return d.InsertRow(engine.Address{Row: 1}) }},
		"deleteRow\t1\n":    {edit: "deleteRow\t1\n", want: func(d engine.Document) engine.Document { return d.DeleteRow(engine.Address{Row: 0}) }},
		"insertCol\tB\n":    {edit: "insertCol\tB\n", want: func(d engine.Document) engine.Document { return d.InsertCol(engine.Address{Col: 1}) }},
		"deleteCol\tA\n":    {edit: "deleteCol\tA\n", want: func(d engine.Document) engine.Document { return d.DeleteCol(engine.Address{Col: 0}) }},
		"duplicateRow\t2\n": {edit: "duplicateRow\t2\n", want: func(d engine.Document) engine.Document { return d.DuplicateRow(engine.Address{Row: 1}) }},
		"duplicateCol\tA\n": {edit: "duplicateCol\tA\n", want: func(d engine.Document) engine.Document { return d.DuplicateCol(engine.Address{Col: 0}) }},
		"fill\tB1\tB2:B3\n": {edit: "fill\tB1\tB2:B3\n", want: func(d engine.Document) engine.Document {
			return d.Fill(engine.Address{Row: 0, Col: 1}, engine.Span{From: engine.Address{Row: 1, Col: 1}, To: engine.Address{Row: 2, Col: 1}})
		}},
		"fill single-cell span": {edit: "fill\tB1\tB2\n", want: func(d engine.Document) engine.Document {
			return d.Fill(engine.Address{Row: 0, Col: 1}, engine.Span{From: engine.Address{Row: 1, Col: 1}, To: engine.Address{Row: 1, Col: 1}})
		}},
		"paste\tC1\tA1\t<block>\n": {edit: "paste\tC1\tA1\t" + base64.StdEncoding.EncodeToString([]byte("9\t8")) + "\n", want: func(d engine.Document) engine.Document {
			out, err := d.Paste(engine.Address{Row: 0, Col: 2}, engine.Address{Row: 0, Col: 0}, block, limits)
			require.NoError(t, err)
			return out
		}},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := engine.Apply(parseDoc(t, src), parseEdits(t, tc.edit), limits)
			require.NoError(t, err)
			assert.Equal(t, string(tc.want(parseDoc(t, src)).Text()), string(got.Text()))
		})
	}
}

func TestApplyPasteLandsBlock(t *testing.T) {
	edit := "paste\tC1\tA1\t" + base64.StdEncoding.EncodeToString([]byte("9")) + "\n"
	assert.Equal(t, "1\t2\t9\n", applied(t, "1\t2\n", edit))
}

func TestApplyBaseMatchProceeds(t *testing.T) {
	doc := parseDoc(t, "1\n")
	edits := parseEdits(t, "#.base\t"+string(engine.Revision(doc))+"\nsetCell\tA1\t2\n")
	out, err := engine.Apply(doc, edits, engine.DefaultLimits())
	require.NoError(t, err)
	assert.Equal(t, "2\n", string(out.Text()))
}

func TestApplyBaseMismatchRefuses(t *testing.T) {
	doc := parseDoc(t, "1\n")
	edits := parseEdits(t, "#.base\t"+strings.Repeat("0", 64)+"\nsetCell\tA1\t2\n")
	_, err := engine.Apply(doc, edits, engine.DefaultLimits())
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrEditsBase)
}

func TestApplyRefusedOpWrapsCauseAndNamesLine(t *testing.T) {
	// GridDim 2 makes the second op's target address refusable at apply time.
	limits := engine.Limits{ResultCells: 10, GridDim: 2, ResultBytes: 100}
	doc := parseDoc(t, "1\t2\n")
	edits := parseEdits(t, "setCell\tA1\tok\nsetCell\tE9\tfar\n")
	_, err := engine.Apply(doc, edits, limits)
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrEditsApply)
	assert.ErrorIs(t, err, constants.ErrInvalidValue)
	assert.Contains(t, err.Error(), "2")
}

func TestApplyIsAtomicOnMidBatchFailure(t *testing.T) {
	limits := engine.Limits{ResultCells: 10, GridDim: 2, ResultBytes: 100}
	doc := parseDoc(t, "1\t2\n")
	edits := parseEdits(t, "setCell\tA1\tok\nsetCell\tE9\tfar\n")
	_, err := engine.Apply(doc, edits, limits)
	require.Error(t, err)
	// The input document is untouched — nothing was partially applied.
	assert.Equal(t, "1\t2\n", string(doc.Text()))
}

func TestApplyDeterminism(t *testing.T) {
	const src, edit = "1\t2\n3\t4\n", "setCell\tB2\t=A1*2\ninsertRow\t1\n"
	first, err := engine.Apply(parseDoc(t, src), parseEdits(t, edit), engine.DefaultLimits())
	require.NoError(t, err)
	second, err := engine.Apply(parseDoc(t, src), parseEdits(t, edit), engine.DefaultLimits())
	require.NoError(t, err)
	assert.Equal(t, engine.Revision(first), engine.Revision(second))
}

func TestApplyEmptyEditsIsIdentity(t *testing.T) {
	doc := parseDoc(t, "1\t2\n")
	out, err := engine.Apply(doc, parseEdits(t, "#. just a note\n"), engine.DefaultLimits())
	require.NoError(t, err)
	assert.Equal(t, engine.Revision(doc), engine.Revision(out))
}

// TestApplyCRLFBatchStoresNoCarriageReturn pins the fixed point: a CRLF batch
// must store exactly what an LF batch stores, so the document re-reads
// identically and its revision names the bytes on disk.
func TestApplyCRLFBatchStoresNoCarriageReturn(t *testing.T) {
	got := applied(t, "1\t2\n", "#.base\t\r\nsetCell\tB1\t9\r\n")
	assert.Equal(t, "1\t9\n", got)
	reparsed, err := engine.ParseDocument([]byte(got))
	require.NoError(t, err)
	assert.Equal(t, got, string(reparsed.Text()), "the written document is a fixed point")
}

func TestApplyCRLFRevisionNamesTheStoredBytes(t *testing.T) {
	doc := parseDoc(t, "1\t2\n")
	out, err := engine.Apply(doc, parseEdits(t, "setCell\tB1\t9\r\n"), engine.DefaultLimits())
	require.NoError(t, err)
	reparsed, err := engine.ParseDocument(out.Text())
	require.NoError(t, err)
	assert.Equal(t, engine.Revision(out), engine.Revision(reparsed))
}

func TestApplyCRLFBaseIsComparable(t *testing.T) {
	doc := parseDoc(t, "1\n")
	edits := parseEdits(t, "#.base\t"+string(engine.Revision(doc))+"\r\nsetCell\tA1\t2\r\n")
	_, err := engine.Apply(doc, edits, engine.DefaultLimits())
	require.NoError(t, err, "a CRLF base line must match the revision it names")
}

// TestApplyRefusesCommentMarkerCells pins the injection refusal: a first-column
// cell that would make its row a comment line deletes that row on the next
// read, shifting every address below it — so the edit is refused, not written.
func TestApplyRefusesCommentMarkerCells(t *testing.T) {
	for name, text := range map[string]string{
		"directive": "#.note",
		"legacy":    "# note",
		"shebang":   "#!/usr/bin/env tsvsheet",
	} {
		t.Run(name, func(t *testing.T) {
			doc := parseDoc(t, "alpha\t10\nbeta\t20\n")
			_, err := engine.Apply(doc, parseEdits(t, "setCell\tA1\t"+text+"\n"), engine.DefaultLimits())
			require.Error(t, err)
			assert.ErrorIs(t, err, constants.ErrCommentCell)
			assert.Equal(t, "alpha\t10\nbeta\t20\n", string(doc.Text()))
		})
	}
}

func TestApplyAllowsCommentMarkerTextAwayFromTheFirstColumn(t *testing.T) {
	// Only the first cell can start a line, so the marker is ordinary text
	// anywhere else — and `#<TAB>` and `#N/A` are data even in column A.
	assert.Equal(t, "1\t#.note\n", applied(t, "1\t2\n", "setCell\tB1\t#.note\n"))
	assert.Equal(t, "#N/A\t2\n", applied(t, "1\t2\n", "setCell\tA1\t#N/A\n"))
}

func TestApplyRefusesCommentMarkerInPastedBlock(t *testing.T) {
	block := base64.StdEncoding.EncodeToString([]byte("#.note\tx"))
	doc := parseDoc(t, "alpha\t10\n")
	_, err := engine.Apply(doc, parseEdits(t, "paste\tA1\tA1\t"+block+"\n"), engine.DefaultLimits())
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrCommentCell)
	assert.Equal(t, "alpha\t10\n", string(doc.Text()))
}

// TestApplyFillHonoursTheGridLimit pins that fill is bounded: Document.Fill
// takes no limits, so without the edits-layer check one short line could grow
// the grid without ceiling.
func TestApplyFillHonoursTheGridLimit(t *testing.T) {
	limits := engine.Limits{ResultCells: 100, GridDim: 10, ResultBytes: 100}
	for name, edit := range map[string]string{
		"rows": "fill\tA1\tA1:A99\n",
		"cols": "fill\tA1\tA1:CZ1\n",
	} {
		t.Run(name, func(t *testing.T) {
			doc := parseDoc(t, "1\n")
			_, err := engine.Apply(doc, parseEdits(t, edit), limits)
			require.Error(t, err)
			assert.ErrorIs(t, err, constants.ErrEditsApply)
			assert.ErrorIs(t, err, constants.ErrInvalidValue)
		})
	}
}

func TestApplyFillWithinTheGridLimitProceeds(t *testing.T) {
	limits := engine.Limits{ResultCells: 100, GridDim: 10, ResultBytes: 100}
	out, err := engine.Apply(parseDoc(t, "7\n"), parseEdits(t, "fill\tA1\tA1:A3\n"), limits)
	require.NoError(t, err)
	assert.Equal(t, "7\n7\n7\n", string(out.Text()))
}

// TestApplyReturnsNoDocumentOnRefusal pins the atomicity contract on the
// *returned* value: a refused batch yields the zero Document, so a caller that
// ignores the error cannot persist a half-applied grid.
func TestApplyReturnsNoDocumentOnRefusal(t *testing.T) {
	limits := engine.Limits{ResultCells: 10, GridDim: 2, ResultBytes: 100}
	out, err := engine.Apply(parseDoc(t, "1\t2\n"), parseEdits(t, "setCell\tA1\tok\nsetCell\tE9\tfar\n"), limits)
	require.Error(t, err)
	assert.Empty(t, string(out.Text()), "a refused batch returns no document, not a partial one")
}

// TestEditsTextIsACopy pins immutability: a caller mutating the returned bytes
// cannot reach into the parsed Edits.
func TestEditsTextIsACopy(t *testing.T) {
	edits := parseEdits(t, "setCell\tA1\tx\n")
	first := edits.Text()
	first[0] = 'X'
	assert.Equal(t, "setCell\tA1\tx\n", string(edits.Text()))
}

// TestParseEditsLastBaseWins pins the metadata rule for a repeated key, and
// that a base line below the ops still governs the batch.
func TestParseEditsLastBaseWins(t *testing.T) {
	edits := parseEdits(t, "#.base\taaa\nsetCell\tA1\tx\n#.base\tbbb\n")
	assert.Equal(t, engine.RevisionHex("bbb"), edits.Base())
}

func TestApplyPreservesComments(t *testing.T) {
	assert.Equal(t, "#. header\n1\t9\n", applied(t, "#. header\n1\t2\n", "setCell\tB1\t9\n"))
}
