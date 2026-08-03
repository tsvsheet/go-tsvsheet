package index_test

import (
	"io"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/index"
)

// gridOracle is the naive truth: every grid row of src, whole-file split.
func gridOracle(src string) [][]string {
	var rows [][]string
	line := 0
	for _, text := range splitAll(src) {
		line++
		if !isComment(index.LineNumber(line), []byte(text)) {
			rows = append(rows, strings.Split(text, "\t"))
		}
	}
	return rows
}

// splitAll applies the test terminator rule to the whole source at once.
func splitAll(src string) []string {
	if src == "" {
		return nil
	}
	normalized := strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(src)
	normalized = strings.TrimSuffix(normalized, "\n")
	return strings.Split(normalized, "\n")
}

// reader builds a Reader over src with the test rules.
func reader(t *testing.T, src string, stride int, capacity index.CellCount) *index.Reader {
	t.Helper()
	opts := index.Options{
		Stride:       index.StrideLines(stride),
		Split:        splitTerminators,
		IsComment:    isComment,
		MaxLineBytes: 1 << 20,
	}
	ix, err := index.Scan(strings.NewReader(src), index.SourceSize(len(src)), opts)
	require.NoError(t, err)
	return index.NewReader(strings.NewReader(src), index.SourceSize(len(src)), ix, opts, capacity)
}

// TestReadRowsMatchesTheOracleForEveryWindow sweeps every (from, n) window of
// every corpus document at several strides and cache capacities: the rows
// returned must equal the oracle's slice, clipped at the grid's end.
func TestReadRowsMatchesTheOracleForEveryWindow(t *testing.T) {
	t.Parallel()

	for name, src := range corpus {
		for _, stride := range []int{1, 2, 64} {
			for _, capacity := range []index.CellCount{1, 8, 1 << 20} {
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					truth := gridOracle(src)
					r := reader(t, src, stride, capacity)
					for from := 0; from <= len(truth)+1; from++ {
						for n := 0; n <= len(truth)-from+2; n++ {
							got, err := r.ReadRows(index.GridRow(from), index.RowCount(n))
							require.NoError(t, err)
							want := truth[min(from, len(truth)):min(from+n, len(truth))]
							if len(want) == 0 {
								assert.Empty(t, got, "from=%d n=%d", from, n)
								continue
							}
							assert.Equal(t, want, got, "from=%d n=%d", from, n)
						}
					}
				})
			}
		}
	}
}

// dribblingAt serves at most one byte per ReadAt call, so every terminator
// can straddle a scanner refill.
type dribblingAt struct{ data []byte }

func (d dribblingAt) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(d.data)) {
		return 0, io.EOF
	}
	p[0] = d.data[off]
	return 1, nil
}

// TestReadRowsAgreesUnderADribblingReader pins the buffer-boundary behavior:
// a one-byte-at-a-time source returns exactly the same rows as a whole-buffer
// one, for every corpus document.
func TestReadRowsAgreesUnderADribblingReader(t *testing.T) {
	t.Parallel()

	for name, src := range corpus {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			opts := index.Options{Stride: 2, Split: splitTerminators, IsComment: isComment, MaxLineBytes: 1 << 20}
			ix, err := index.Scan(strings.NewReader(src), index.SourceSize(len(src)), opts)
			require.NoError(t, err)
			whole := index.NewReader(strings.NewReader(src), index.SourceSize(len(src)), ix, opts, 1<<20)
			dribbled := index.NewReader(dribblingAt{data: []byte(src)}, index.SourceSize(len(src)), ix, opts, 1<<20)
			wantRows, err := whole.ReadRows(0, 1<<20)
			require.NoError(t, err)
			gotRows, err := dribbled.ReadRows(0, 1<<20)
			require.NoError(t, err)
			assert.Equal(t, wantRows, gotRows)
			assert.Equal(t, gridOracle(src), wantRows)
		})
	}
}

// TestReadRowsClipsHostileWindows pins the both-ends clip: a negative start, a
// negative count, a huge count, and a past-the-grid start are smaller results,
// never a panic or an oversized allocation.
func TestReadRowsClipsHostileWindows(t *testing.T) {
	t.Parallel()

	src := "a\nb\nc\n"
	r := reader(t, src, 2, 1<<20)
	truth := gridOracle(src)

	got, err := r.ReadRows(-1, 2)
	require.NoError(t, err)
	assert.Equal(t, truth[0:2], got, "a negative start clips to row 0")
	got, err = r.ReadRows(0, -1)
	require.NoError(t, err)
	assert.Empty(t, got, "a negative count is an empty window")
	got, err = r.ReadRows(1, index.RowCount(1)<<40)
	require.NoError(t, err)
	assert.Equal(t, truth[1:], got, "a huge count clips to the grid without allocating for it")
	got, err = r.ReadRows(1, math.MaxInt64)
	require.NoError(t, err)
	assert.Equal(t, truth[1:], got, "the extreme count answers identically — no sum may wrap")
	got, err = r.ReadRows(9, 3)
	require.NoError(t, err)
	assert.Empty(t, got, "past the grid is empty")
}

// TestReadRowsReturnsCallerOwnedCopies pins the aliasing contract the review
// exposed: writing into a returned row must never poison a later read, under
// any cache pressure — identical call sequences answer identically.
func TestReadRowsReturnsCallerOwnedCopies(t *testing.T) {
	t.Parallel()

	for _, capacity := range []index.CellCount{1, 1 << 20} {
		r := reader(t, "a\tb\nc\n", 2, capacity)
		got, err := r.ReadRows(0, 1)
		require.NoError(t, err)
		got[0][0] = "POISON"
		again, err := r.ReadRows(0, 1)
		require.NoError(t, err)
		assert.Equal(t, [][]string{{"a", "b"}}, again, "capacity %d", capacity)
	}
}

// TestReadRowsRefusesASourceThatNoLongerMatchesTheIndex pins the skew
// contract: a same-size edit that turns indexed data rows into comments makes
// a block unable to cover its promised rows — ErrScan, never a spin with the
// lock held.
func TestReadRowsRefusesASourceThatNoLongerMatchesTheIndex(t *testing.T) {
	t.Parallel()

	indexed := "a\nb\nc\nd\n"
	edited := "a\nb\n#.\n#.\n"
	opts := index.Options{Stride: 2, Split: splitTerminators, IsComment: isComment, MaxLineBytes: 1 << 20}
	ix, err := index.Scan(strings.NewReader(indexed), index.SourceSize(len(indexed)), opts)
	require.NoError(t, err)
	r := index.NewReader(strings.NewReader(edited), index.SourceSize(len(edited)), ix, opts, 1<<20)
	_, err = r.ReadRows(2, 1)
	assert.ErrorIs(t, err, index.ErrScan)
}

// TestReadRowsRefusalsSurfaceErrScan pins that a source failing mid-read is
// the specific sentinel, not a bare error or a silent short result.
func TestReadRowsRefusalsSurfaceErrScan(t *testing.T) {
	t.Parallel()

	good := "a\nb\nc\nd\ne\nf\ng\nh\n"
	opts := index.Options{Stride: 2, Split: splitTerminators, IsComment: isComment, MaxLineBytes: 1 << 20}
	ix, err := index.Scan(strings.NewReader(good), index.SourceSize(len(good)), opts)
	require.NoError(t, err)

	failing := index.NewReader(failingReader{prefix: []byte("a\nb\n")}, index.SourceSize(len(good)), ix, opts, 1<<20)
	_, err = failing.ReadRows(6, 2)
	assert.ErrorIs(t, err, index.ErrScan)
}

// TestReaderAppliesTheOptionDefaults pins that a Reader built with zero
// Stride and MaxLineBytes reads under the package defaults — the same
// fallbacks Scan applies — rather than a degenerate zero stride or bound.
func TestReaderAppliesTheOptionDefaults(t *testing.T) {
	t.Parallel()

	src := "a\tb\nc\n"
	opts := index.Options{Split: splitTerminators, IsComment: isComment}
	ix, err := index.Scan(strings.NewReader(src), index.SourceSize(len(src)), opts)
	require.NoError(t, err)
	r := index.NewReader(strings.NewReader(src), index.SourceSize(len(src)), ix, opts, 1<<20)
	got, err := r.ReadRows(0, 10)
	require.NoError(t, err)
	assert.Equal(t, gridOracle(src), got)
}

// TestScanBlockCarriesTheExactPhysicalLine pins the checkpoint's carried line
// number with a rule that is line-sensitive PAST line 1 (the shebang rule
// is not, since a shebang line is never a data line, so any off-by-one in
// the carried line was invisible): line 2, and only line 2, is a comment when
// it starts with '%' — sitting INSIDE the first block, where scanBlock's
// arithmetic is what numbers it. A one-off error in scanBlock's line arithmetic
// misclassifies it and the rows diverge from the oracle.
func TestScanBlockCarriesTheExactPhysicalLine(t *testing.T) {
	t.Parallel()

	lineThree := func(line index.LineNumber, text []byte) bool {
		return line == 2 && len(text) > 0 && text[0] == '%'
	}
	src := "a\n%skip\nb\n%data\nc\n" // line 2 is a comment INSIDE block one; line 4's % is data
	opts := index.Options{Stride: 2, Split: splitTerminators, IsComment: lineThree, MaxLineBytes: 1 << 20}
	ix, err := index.Scan(strings.NewReader(src), index.SourceSize(len(src)), opts)
	require.NoError(t, err)
	r := index.NewReader(strings.NewReader(src), index.SourceSize(len(src)), ix, opts, 1<<20)
	got, err := r.ReadRows(0, 10)
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"a"}, {"b"}, {"%data"}, {"c"}}, got)
}
