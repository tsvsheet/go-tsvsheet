package engine_test

import (
	"crypto/sha256"
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

func TestApplyFoldsInOrder(t *testing.T) {
	assert.Equal(t, "b\n", applied(t, "x\n", "setCell\tA1\ta\nsetCell\tA1\tb\n"))
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

func TestApplyPreservesComments(t *testing.T) {
	assert.Equal(t, "#. header\n1\t9\n", applied(t, "#. header\n1\t2\n", "setCell\tB1\t9\n"))
}
