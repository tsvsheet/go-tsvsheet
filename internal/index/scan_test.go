package index_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/index"
)

// markOpts is the engine's scan wiring plus a Mark request: newline-split
// lines, `#`-prefixed comments, and the marker under test.
func markOpts(mark string, maxMarks index.CellCount) index.Options {
	return index.Options{
		Split: bufio.ScanLines,
		IsComment: func(_ index.LineNumber, text []byte) bool {
			return strings.HasPrefix(string(text), "#")
		},
		Mark:     []byte(mark),
		MaxMarks: maxMarks,
	}
}

// scanMarks scans src under a mark request.
func scanMarks(t *testing.T, src, mark string, maxMarks index.CellCount) index.Index {
	t.Helper()
	ix, err := index.Scan(bytes.NewReader([]byte(src)), index.SourceSize(len(src)), markOpts(mark, maxMarks))
	require.NoError(t, err)
	return ix
}

// TestScanMarksDataRowsContainingTheSequence states the marker contract: the
// scan records the GRID rows — comment lines skipped, exactly as A1 rows are
// counted — whose bytes contain the requested sequence, in row order.
func TestScanMarksDataRowsContainingTheSequence(t *testing.T) {
	t.Parallel()

	src := "plain\n# c |@ commented out\n=1 |@ named(x)\tcell\nplain again\n=2 |@ named(y)\n"
	ix := scanMarks(t, src, "|@", 100)
	assert.Equal(t, []index.GridRow{1, 3}, ix.Marked(),
		"the comment line is not a grid row and cannot be marked")
	assert.False(t, ix.MarkedOverflowed())
}

// TestScanWithoutAMarkRecordsNothing states the default: no Mark requested, no
// marks collected, no per-line search paid.
func TestScanWithoutAMarkRecordsNothing(t *testing.T) {
	t.Parallel()

	ix := scanMarks(t, "=1 |@ named(x)\n", "", 0)
	assert.Empty(t, ix.Marked())
	assert.False(t, ix.MarkedOverflowed())
}

// TestScanMarkOverflowDropsTheSet states the cap: past MaxMarks rows the whole
// set is dropped and the overflow flagged — a truncated set would answer
// "which rows are marked" with a silent lie, where the flag makes the refusal
// inspectable.
func TestScanMarkOverflowDropsTheSet(t *testing.T) {
	t.Parallel()

	ix := scanMarks(t, strings.Repeat("=1 |@ named(x)\n", 5), "|@", 3)
	assert.True(t, ix.MarkedOverflowed())
	assert.Empty(t, ix.Marked(), "an incomplete set is not served")
}

// TestScanMarkAtExactlyTheCapDoesNotOverflow states the boundary from both
// sides — the adversary's surviving mutation M7 was exactly the untested
// second half: MaxMarks rows is within the cap, and ONE more row flips the
// whole set to the overflow refusal.
func TestScanMarkAtExactlyTheCapDoesNotOverflow(t *testing.T) {
	t.Parallel()

	at := scanMarks(t, strings.Repeat("=1 |@ named(x)\n", 3), "|@", 3)
	assert.False(t, at.MarkedOverflowed())
	assert.Len(t, at.Marked(), 3)

	over := scanMarks(t, strings.Repeat("=1 |@ named(x)\n", 4), "|@", 3)
	assert.True(t, over.MarkedOverflowed(), "cap + 1 overflows")
	assert.Empty(t, over.Marked())
}

// TestCheckedSplitNeverAdmitsADefectiveRule pins the two injected-Split defects as
// ErrScan: the skip form (advance without a token — bufio drops the remainder
// at EOF, making the census depend on read chunking) and a token without
// progress (an unbounded loop otherwise). A defective engine rule must surface
// as a refusal, never as a wrong index or a hang.
func TestCheckedSplitNeverAdmitsADefectiveRule(t *testing.T) {
	t.Parallel()

	skipForm := func(data []byte, isAtEOF bool) (int, []byte, error) {
		if len(data) > 0 && data[0] == '!' {
			i := bytes.IndexByte(data, '\n')
			if i >= 0 {
				return i + 1, nil, nil // consume the line, emit nothing
			}
		}
		return splitTerminators(data, isAtEOF)
	}
	_, err := index.Scan(strings.NewReader("!skip\na\n"), 8, index.Options{Split: skipForm, IsComment: isComment})
	assert.ErrorIs(t, err, index.ErrScan, "the skip form is refused")

	stuck := func([]byte, bool) (int, []byte, error) { return 0, []byte{}, nil }
	_, err = index.Scan(strings.NewReader("abc"), 3, index.Options{Split: stuck, IsComment: isComment})
	assert.ErrorIs(t, err, index.ErrScan, "a token without progress is refused, not looped on")
}

// failingReader errors on any read past its prefix, standing in for a source
// that dies mid-scan.
type failingReader struct{ prefix []byte }

func (r failingReader) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(r.prefix)) {
		return 0, assert.AnError
	}
	n := copy(p, r.prefix[off:])
	return n, nil
}

// TestScanRefusals pins the two ErrScan shapes: the required injected rules
// missing, and a source that fails mid-read — each the specific sentinel,
