package engine_test

import (
	"errors"
	"io"
	"strconv"
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

// TestMaxLineBytesBoundsASingleScannedRow pins the memory ceiling on input. A
// scanner with no bound turns one pathological line into an allocation the
// caller never asked for and cannot catch — in the wasm build a fatal abort
// that kills the engine for the page's life.
func TestMaxLineBytesBoundsASingleScannedRow(t *testing.T) {
	t.Parallel()
	_, err := engine.Parse([]byte(strings.Repeat("a", 2<<20) + "\n"))

	require.Error(t, err, "a line past the ceiling is refused, not allocated")
}

// TestScanLineNormalizesEveryLineTerminator pins the CR fix FuzzParseDocument
// forced: LF, CRLF, and a lone CR each terminate a line, so no carriage return
// can survive into cell text and break the canonical round-trip.
func TestScanLineNormalizesEveryLineTerminator(t *testing.T) {
	t.Parallel()

	cases := map[string][][]string{
		"a\nb\n":   {{"a"}, {"b"}},
		"a\r\nb\n": {{"a"}, {"b"}},
		"a\rb\n":   {{"a"}, {"b"}},
		"\r\r":     {{""}, {""}},
		"a\r":      {{"a"}},
	}
	for src, want := range cases {
		t.Run(strconv.Quote(src), func(t *testing.T) {
			t.Parallel()
			g, err := engine.ReadTSV(strings.NewReader(src))
			require.NoError(t, err)
			assert.Equal(t, engine.Grid(want), g)
		})
	}
}

// TestWriteTSVRefusesARowThatWouldReadBackAsAComment pins the round-trip
// closure FuzzReadTSV forced: a document may legally hold `#!x` as data below
// its comment lines, but the bare grid format has no escape for a leading
// marker, so writing that row would emit a line the next read drops — data
// loss dressed as output. WriteTSV refuses instead.
func TestWriteTSVRefusesARowThatWouldReadBackAsAComment(t *testing.T) {
	t.Parallel()

	g, err := engine.ReadTSV(strings.NewReader("# \n#!x\n"))
	require.NoError(t, err)
	require.Equal(t, engine.Grid{{"#!x"}}, g, "below its comments, #!x is data")

	var b strings.Builder
	err = engine.WriteTSV(&b, g)
	assert.ErrorIs(t, err, constants.ErrInvalidValue)
	assert.Empty(t, b.String(), "nothing is written before the refusal")

	for _, row := range []string{"#. d", "# legacy"} {
		err := engine.WriteTSV(&strings.Builder{}, engine.Grid{{"x"}, {row}})
		assert.ErrorIs(t, err, constants.ErrInvalidValue, row)
	}
	require.NoError(
		t,
		engine.WriteTSV(&strings.Builder{}, engine.Grid{{"#N/A", "#!not-first-cell"}}),
		"markers away from column A are data",
	)
}

// oneByteReader dribbles its content a single byte per Read, so a terminator
// split across scanner refills is exercised the way a socket or pipe delivers
// it.
type oneByteReader struct{ rest []byte }

func (r *oneByteReader) Read(p []byte) (int, error) {
	if len(r.rest) == 0 {
		return 0, io.EOF
	}
	p[0] = r.rest[0]
	r.rest = r.rest[1:]
	return 1, nil
}

// TestCrLineHoldsASplitCRLFAcrossReads pins crLine's wait-for-one-more-byte
// branch: with a CR as the last buffered byte mid-stream, the scanner must ask
// for the next byte to tell a lone CR from a split CRLF, or every CRLF whose
// halves land in different reads would mint a phantom empty row. Every test
// reader elsewhere delivers its whole content in one read, so only a dribbling
// reader can reach this branch.
func TestCrLineHoldsASplitCRLFAcrossReads(t *testing.T) {
	t.Parallel()

	for _, src := range []string{"a\r\nb\n", "a\r\n", "x\r\ny\r\nz\n", "\r\n\r\n", "a\rb\n", "a\r"} {
		t.Run(strconv.Quote(src), func(t *testing.T) {
			t.Parallel()
			whole, err := engine.ReadTSV(strings.NewReader(src))
			require.NoError(t, err)
			dribbled, err := engine.ReadTSV(&oneByteReader{rest: []byte(src)})
			require.NoError(t, err)
			assert.Equal(t, whole, dribbled, "streamed and whole-buffer reads must agree")
		})
	}
}
