package index_test

import (
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

// TestReadRowsAgreesUnderADribblingReader pins the buffer-boundary behavior:
// a one-byte-at-a-time source returns the same rows as a whole-buffer one.
func TestReadRowsAgreesUnderADribblingReader(t *testing.T) {
	t.Parallel()

	src := "a\tb\r\nc\r#. note\nd\te\tf\n"
	opts := index.Options{Stride: 2, Split: splitTerminators, IsComment: isComment, MaxLineBytes: 1 << 20}
	ix, err := index.Scan(strings.NewReader(src), index.SourceSize(len(src)), opts)
	require.NoError(t, err)

	whole := index.NewReader(strings.NewReader(src), index.SourceSize(len(src)), ix, opts, 1<<20)
	wantRows, err := whole.ReadRows(0, 10)
	require.NoError(t, err)
	assert.Equal(t, gridOracle(src), wantRows)
}

// TestCachedCellsStaysUnderCapacity pins the memory bound: however many
// windows are read, the cache never retains more cells than its capacity
// (single blocks larger than the whole capacity are the exception — a block is
// the unit of eviction and one must be resident to answer).
func TestCachedCellsStaysUnderCapacity(t *testing.T) {
	t.Parallel()

	src := strings.Repeat("a\tb\tc\td\n", 64) // 64 rows × 4 cells
	r := reader(t, src, 4, 16)                // blocks of 4 rows = 16 cells; capacity one block
	for from := 0; from < 64; from += 4 {
		_, err := r.ReadRows(index.GridRow(from), 4)
		require.NoError(t, err)
		assert.LessOrEqual(t, r.CachedCells(), index.CellCount(16), "after window at %d", from)
	}
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

// TestEvictAlwaysKeepsTheMostRecentBlockResident pins evict's invariant by
// starving it: a capacity smaller than any single block must still answer
// every window correctly — the block being read survives eviction, because it
// is the unit the answer is served from — while the cache never holds more
// than that one block.
func TestEvictAlwaysKeepsTheMostRecentBlockResident(t *testing.T) {
	t.Parallel()

	src := strings.Repeat("a\tb\tc\td\n", 16) // blocks of 4 rows = 16 cells
	r := reader(t, src, 4, 1)                 // capacity far below one block
	truth := gridOracle(src)
	for from := 0; from < 16; from += 4 {
		got, err := r.ReadRows(index.GridRow(from), 4)
		require.NoError(t, err)
		assert.Equal(t, truth[from:from+4], got)
		assert.Equal(t, index.CellCount(16), r.CachedCells(), "exactly the current block stays resident")
	}
}
