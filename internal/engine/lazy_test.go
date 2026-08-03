package engine_test

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// regionFailSource fails reads below a byte offset once armed — the probe that
// lets a window's own block read succeed while a dependency's block, earlier
// in the file, fails.
type regionFailSource struct {
	data      []byte
	failBelow int64
}

func (r *regionFailSource) ReadAt(p []byte, off int64) (int, error) {
	if r.failBelow > 0 && off < r.failBelow {
		return 0, assert.AnError
	}
	if off >= int64(len(r.data)) {
		return 0, io.EOF
	}
	return copy(p, r.data[off:]), nil
}

// TestLazyCellsNeverServeWrongDataAfterAFailure pins the lazy latch: the window
// itself reads clean, a dependency living in an earlier block fails, and the
// whole call answers ErrReadInput — never a partial grid over unread data.
func TestLazyCellsNeverServeWrongDataAfterAFailure(t *testing.T) {
	t.Parallel()

	// A1, the dependency, in block one; 300 filler rows push the formula past
	// the first stride's block.
	doc := "7\n" + strings.Repeat("x\n", 300) + "=A1+1\n"
	src := &regionFailSource{data: []byte(doc)}

	tiny := engine.Limits{ResultCells: 1 << 20, GridDim: 1 << 20, ResultBytes: 100, ResidentCells: 1}
	_, windowed, err := engine.OpenSheet(engine.ByteSource{ReadAt: src, Size: int64(len(src.data))}, tiny)
	require.NoError(t, err)
	require.NotNil(t, windowed)

	src.failBelow = 4 // the dependency's block starts at offset 0; the window's does not
	_, err = windowed.ComputeRows(301, 1, engine.ComputeOptions{Limits: tiny})
	assert.ErrorIs(t, err, constants.ErrReadInput)
}

// countingRegionSource counts reads below a byte offset once armed, failing
// them — the probe that proves the latch short-circuits the SECOND dependency
// read rather than retrying a dead source.
type countingRegionSource struct {
	data      []byte
	failBelow int64
	lowReads  int
}

func (r *countingRegionSource) ReadAt(p []byte, off int64) (int, error) {
	if r.failBelow > 0 && off < r.failBelow {
		r.lowReads++
		return 0, assert.AnError
	}
	if off >= int64(len(r.data)) {
		return 0, io.EOF
	}
	return copy(p, r.data[off:]), nil
}

// TestLazyCellsLatchShortCircuitsRetries pins the latch's other half (the
// adversary's M5 survivor): after the first dependency read fails, later reads
// answer through the latch without touching the dead source again — one failed
// low-offset read, not two — and the call still fails whole.
func TestLazyCellsLatchShortCircuitsRetries(t *testing.T) {
	t.Parallel()

	doc := "7\n" + strings.Repeat("x\n", 300) + "=sum(A1, A1)\n"
	src := &countingRegionSource{data: []byte(doc)}
	tiny := engine.Limits{ResultCells: 1 << 20, GridDim: 1 << 20, ResultBytes: 100, ResidentCells: 1}
	_, windowed, err := engine.OpenSheet(engine.ByteSource{ReadAt: src, Size: int64(len(src.data))}, tiny)
	require.NoError(t, err)
	require.NotNil(t, windowed)

	src.failBelow = 4
	_, err = windowed.ComputeRows(301, 1, engine.ComputeOptions{Limits: tiny})
	assert.ErrorIs(t, err, constants.ErrReadInput)
	assert.Equal(t, 1, src.lowReads, "the second read of A1 answers through the latch, never the dead source")

	// A second window over DIFFERENT dead-region addresses: the memo shields
	// repeats of A1, so =A2 is what proves the latch itself short-circuits a
	// fresh address without touching the source again.
	pair := "7\n8\n" + strings.Repeat("x\n", 300) + "=sum(A1, A1)\t=A2\n"
	pairSrc := &countingRegionSource{data: []byte(pair)}
	_, windowedPair, err := engine.OpenSheet(
		engine.ByteSource{ReadAt: pairSrc, Size: int64(len(pairSrc.data))}, tiny,
	)
	require.NoError(t, err)
	require.NotNil(t, windowedPair)
	pairSrc.failBelow = 4
	_, err = windowedPair.ComputeRows(302, 1, engine.ComputeOptions{Limits: tiny})
	assert.ErrorIs(t, err, constants.ErrReadInput)
	assert.Equal(t, 1, pairSrc.lowReads, "the fresh address A2 answers through the latch too")
}
