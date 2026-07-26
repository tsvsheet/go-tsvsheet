package engine_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

func TestReadTSV(t *testing.T) {
	t.Parallel()

	g, err := engine.ReadTSV(strings.NewReader("a\tb\n1\t2\n"))
	require.NoError(t, err)
	assert.Equal(t, engine.Grid{{"a", "b"}, {"1", "2"}}, g)
}

func TestReadTSV_SkipsComments(t *testing.T) {
	t.Parallel()

	// A first-line shebang and any `# ` line are skipped and do not occupy a
	// row; a `#N/A` cell (hash then a non-space) stays data.
	g, err := engine.ReadTSV(strings.NewReader(
		"#!/usr/bin/env tsvsheet\n# a note\na\tb\n# mid\n#N/A\t=A2\n",
	))
	require.NoError(t, err)
	assert.Equal(t, engine.Grid{{"a", "b"}, {"#N/A", "=A2"}}, g)
}

func TestReadTSV_CommentOrDataOnFirstLine(t *testing.T) {
	t.Parallel()

	// A `# ` comment on the first line is skipped; a data first line is kept.
	comment, err := engine.ReadTSV(strings.NewReader("# header\nx\ty\n"))
	require.NoError(t, err)
	assert.Equal(t, engine.Grid{{"x", "y"}}, comment)

	data, err := engine.ReadTSV(strings.NewReader("x\ty\n"))
	require.NoError(t, err)
	assert.Equal(t, engine.Grid{{"x", "y"}}, data)
}

func TestReadTSV_Ragged(t *testing.T) {
	t.Parallel()

	g, err := engine.ReadTSV(strings.NewReader("a\tb\tc\n1\n"))
	require.NoError(t, err)
	assert.Equal(t, engine.Grid{{"a", "b", "c"}, {"1"}}, g)
}

func TestReadTSV_Empty(t *testing.T) {
	t.Parallel()

	g, err := engine.ReadTSV(strings.NewReader(""))
	require.NoError(t, err)
	assert.Empty(t, g)
}

// failingReader always errors, exercising the ReadTSV scan-error path.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errReadTest }

var errReadTest = errors.New("read failed")

func TestReadTSV_Error(t *testing.T) {
	t.Parallel()

	_, err := engine.ReadTSV(failingReader{})
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrReadInput)
}

func TestWriteTSV(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	require.NoError(t, engine.WriteTSV(&b, engine.Grid{{"a", "b"}, {"1", "2"}}))
	assert.Equal(t, "a\tb\n1\t2\n", b.String())
}

// failingWriter errors after n successful bytes, exercising the WriteTSV error
// path.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errWriteTest }

var errWriteTest = errors.New("write failed")

func TestWriteTSV_Error(t *testing.T) {
	t.Parallel()

	err := engine.WriteTSV(failingWriter{}, engine.Grid{{"a"}})
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrWriteFile)
}

func TestReadWriteRoundTrip(t *testing.T) {
	t.Parallel()

	const in = "1\t2\t3\n4\t5\t6\n"
	g, err := engine.ReadTSV(strings.NewReader(in))
	require.NoError(t, err)

	var b strings.Builder
	require.NoError(t, engine.WriteTSV(&b, g))
	assert.Equal(t, in, b.String())
}

// TestIsCommentLine pins the marker family of SPECIFICATION §3: `#!` shebangs
// the first line only, `#.` marks a directive or comment anywhere, `# `
// (hash-space) is the legacy comment form still accepted, and anything else —
// including a hash followed by a TAB or by a letter — is data. The hash-TAB
// case is the one the `#.` marker exists to disambiguate: it is data, and a
// phantom data row would shift every A1 address below it.
func TestIsCommentLine(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		text engine.SourceLine
		at   engine.LineNumber
		want bool
	}{
		{"directive", "#.hide-cols\tB-M", 3, true},
		{"directive on line 1", "#.header-rows\t1", 1, true},
		{"prose with the directive marker", "#.some comment", 2, true},
		{"bare marker", "#.", 2, true},
		{"legacy hash-space", "# a note", 2, true},
		{"legacy hash-space, empty", "# ", 2, true},
		{"shebang on line 1", "#!/usr/bin/env tsvsheet", 1, true},
		{"shebang below line 1 is data", "#!/usr/bin/env tsvsheet", 2, false},
		{"hash-tab is data", "#\tnot a comment", 2, false},
		{"error value is data", "#N/A\t=A2", 2, false},
		{"bare hash is data", "#", 2, false},
		{"ordinary row", "a\tb", 2, false},
		{"empty line", "", 2, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.want, engine.IsCommentLine(c.at, c.text))
		})
	}
}

// TestReadTSV_SkipsDirectiveMarker proves the reader honors `#.` exactly as it
// honors the legacy marker — skipped, occupying no row — while a hash-TAB line
// stays a data row, so mistaking one for the other is visible in the grid
// rather than silent.
func TestReadTSV_SkipsDirectiveMarker(t *testing.T) {
	t.Parallel()

	g, err := engine.ReadTSV(strings.NewReader(
		"#!/usr/bin/env tsvsheet\n#.hide-cols\tB-M\na\tb\n#.a note\n#N/A\t=A2\n",
	))
	require.NoError(t, err)
	assert.Equal(t, engine.Grid{{"a", "b"}, {"#N/A", "=A2"}}, g)

	data, err := engine.ReadTSV(strings.NewReader("#\tnot a comment\na\tb\n"))
	require.NoError(t, err)
	assert.Equal(t, engine.Grid{{"#", "not a comment"}, {"a", "b"}}, data)
}

// TestParseDocumentRoundTripsBothMarkers proves a sheet mixing the canonical
// and legacy markers survives ParseDocument -> Text byte-for-byte, so adopting
// `#.` never rewrites a file that still uses `# `.
func TestParseDocumentRoundTripsBothMarkers(t *testing.T) {
	t.Parallel()

	src := "#!/usr/bin/env tsvsheet\n#.hide-cols\tB-M\n# legacy note\na\tb\n#.trailing\n"
	doc, err := engine.ParseDocument([]byte(src))
	require.NoError(t, err)
	assert.Equal(t, src, string(doc.Text()))
}
