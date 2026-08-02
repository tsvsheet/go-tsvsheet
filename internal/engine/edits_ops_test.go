package engine_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// The operand grammar of each wire op: which fields it takes, how they are
// spelled, and what a malformed one reports. The envelope — framing, revision
// checking, folding — is exercised in edits_test.go beside the code that owns it.

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

// TestParseEditsLastBaseWins pins the metadata rule for a repeated key, and
// that a base line below the ops still governs the batch.
func TestParseEditsLastBaseWins(t *testing.T) {
	edits := parseEdits(t, "#.base\taaa\nsetCell\tA1\tx\n#.base\tbbb\n")
	assert.Equal(t, engine.RevisionHex("bbb"), edits.Base())
}

// TestEditParsersCoverTheDocumentEditSurfaceAndNothingElse pins the op set as
// a set, not as a list someone remembered to update. Every Document edit has a
// wire op, so a client can express anything the engine can do; move ops are
// deliberately absent because cut/move semantics await their own ruling, and a
// half-specified move op would be worse than none.
func TestEditParsersCoverTheDocumentEditSurfaceAndNothingElse(t *testing.T) {
	t.Parallel()
	for _, op := range []string{
		"setCell\tA1\t1", "insertRow\t2", "deleteRow\t1", "insertCol\tB",
		"deleteCol\tB", "duplicateRow\t1", "duplicateCol\tB",
		"fill\tA1\tA2:B3", "paste\tA1\tA1\teA==",
	} {
		_, err := engine.ParseEdits([]byte(op + "\n"))

		assert.NoError(t, err, "%s is part of the edit surface", op)
	}

	for _, absent := range []string{"moveRow\t1\t2", "cut\tA1", "swap\tA1\tA2"} {
		_, err := engine.ParseEdits([]byte(absent + "\n"))

		assert.Error(t, err, "%s is not an op, and is not quietly ignored either", absent)
	}
}

// TestEditArityRefusesAnOpWithTheWrongFieldCount pins the arity check. A
// silently ignored extra field would let a client believe it had said something
// the engine never heard.
func TestEditArityRefusesAnOpWithTheWrongFieldCount(t *testing.T) {
	t.Parallel()
	_, exact := engine.ParseEdits([]byte("insertRow\t2\n"))
	require.NoError(t, exact)

	_, tooMany := engine.ParseEdits([]byte("insertRow\t2\textra\n"))
	_, tooFew := engine.ParseEdits([]byte("insertRow\n"))

	assert.Error(t, tooMany, "an extra field is a mistake, not a courtesy")
	assert.Error(t, tooFew)
}
