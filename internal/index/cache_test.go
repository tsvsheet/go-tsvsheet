package index_test

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/index"
)

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

// poisonableAt serves its bytes until poisoned, then fails every read — the
// probe that tells a cache hit (no I/O) from a rescan.
type poisonableAt struct {
	data     []byte
	poisoned bool
}

func (pa *poisonableAt) ReadAt(p []byte, off int64) (int, error) {
	if pa.poisoned {
		return 0, assert.AnError
	}
	if off >= int64(len(pa.data)) {
		return 0, io.EOF
	}
	n := copy(p, pa.data[off:])
	return n, nil
}

// TestRecencyPromotionDecidesWhoSurvivesEviction pins the LRU's recency: with
// room for two blocks, touching A again before C arrives must evict B, not A —
// proven by poisoning the source and showing A still answers from cache while
// B needs a rescan and fails.
func TestRecencyPromotionDecidesWhoSurvivesEviction(t *testing.T) {
	t.Parallel()

	src := &poisonableAt{data: []byte("a\nb\nc\nd\ne\nf\n")} // three 2-row blocks of 2 cells
	opts := index.Options{Stride: 2, Split: splitTerminators, IsComment: isComment, MaxLineBytes: 1 << 20}
	ix, err := index.Scan(strings.NewReader(string(src.data)), index.SourceSize(len(src.data)), opts)
	require.NoError(t, err)
	r := index.NewRowReader(src, index.SourceSize(len(src.data)), ix, opts, 4) // two blocks

	mustRead := func(row index.GridRow) {
		_, readErr := r.ReadRows(row, 2)
		require.NoError(t, readErr)
	}
	mustRead(0) // A
	mustRead(2) // B
	mustRead(0) // touch A: promote
	mustRead(4) // C arrives: B must be the eviction
	src.poisoned = true

	_, err = r.ReadRows(0, 2)
	assert.NoError(t, err, "A was promoted, so it survived and answers from cache")
	_, err = r.ReadRows(2, 2)
	assert.ErrorIs(t, err, index.ErrScan,
		"B was evicted for real — the map holds no stale entry — so it needs the poisoned source")
}
