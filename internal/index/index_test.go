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

// The tests wire the two injected language rules with self-contained stand-ins
// (the engine injects its own single definitions in production): lines split on
// LF, CRLF, or lone CR; a line is a comment when it starts "#." or "# ", plus
// "#!" on line 1 only — the .tsvt rule, restated here as the tests' oracle.

// splitTerminators is the test SplitFunc: LF, CRLF, and lone CR all end a line.
func splitTerminators(data []byte, isAtEOF bool) (int, []byte, error) {
	i := bytes.IndexAny(data, "\r\n")
	if i < 0 {
		if isAtEOF && len(data) > 0 {
			return len(data), data, nil
		}
		return 0, nil, nil
	}
	if data[i] == '\n' {
		return i + 1, data[:i], nil
	}
	if i+1 < len(data) && data[i+1] == '\n' {
		return i + 2, data[:i], nil
	}
	if i+1 == len(data) && !isAtEOF {
		return 0, nil, nil
	}
	return i + 1, data[:i], nil
}

// isComment is the test classifier, mirroring the .tsvt comment-line rule.
func isComment(line index.LineNumber, text []byte) bool {
	if line == 1 && bytes.HasPrefix(text, []byte("#!")) {
		return true
	}
	return bytes.HasPrefix(text, []byte("#.")) || bytes.HasPrefix(text, []byte("# "))
}

// scan runs index.Scan over src with the test rules and a given stride.
func scan(t *testing.T, src string, stride int) index.Index {
	t.Helper()
	ix, err := index.Scan(strings.NewReader(src), index.SourceSize(len(src)), index.Options{
		Stride:       index.StrideLines(stride),
		Split:        splitTerminators,
		IsComment:    isComment,
		MaxLineBytes: 1 << 20,
	})
	require.NoError(t, err)
	return ix
}

// oracle computes the census, per-row offsets, physical lines, and cumulative
// counters naively from the whole text — the independent truth every index
// answer is compared against. Its scanner carries the same 1 MiB bound the
// fuzz target scans under and its error is fatal, so the oracle can never
// silently truncate and prove nothing (the adversary's finding: an unchecked
// default 64 KiB cap made long-line fuzz cases compare against a wrong truth).
type oracleDoc struct {
	rowOffsets  []int64              // byte offset of each GRID row's physical line
	rowLines    []index.LineNumber   // physical line of each grid row
	rowCells    []index.CellCount    // cumulative cells BEFORE each grid row
	rowFormulas []index.FormulaCount // cumulative formulas BEFORE each grid row
	census      index.Census
}

func oracle(tb testing.TB, src string) oracleDoc {
	tb.Helper()
	var doc oracleDoc
	var offset int64
	line := 0
	scanner := bufio.NewScanner(strings.NewReader(src))
	scanner.Buffer(nil, 1<<20)
	scanner.Split(splitTerminators)
	rest := []byte(src)
	for scanner.Scan() {
		line++
		text := scanner.Text()
		consumed := advanceOf(rest, text)
		doc.census.Lines++
		if !isComment(index.LineNumber(line), []byte(text)) {
			doc.rowOffsets = append(doc.rowOffsets, offset)
			doc.rowLines = append(doc.rowLines, index.LineNumber(line))
			doc.rowCells = append(doc.rowCells, doc.census.Cells)
			doc.rowFormulas = append(doc.rowFormulas, doc.census.Formulas)
			doc.census.Rows++
			fields := strings.Split(text, "\t")
			doc.census.Cells += index.CellCount(len(fields))
			if len(fields) > doc.census.MaxWidth {
				doc.census.MaxWidth = len(fields)
			}
			for _, f := range fields {
				if strings.HasPrefix(f, "=") {
					doc.census.Formulas++
				}
			}
		}
		offset += int64(consumed)
		rest = rest[consumed:]
	}
	require.NoError(tb, scanner.Err(), "the oracle itself must never truncate")
	return doc
}

// advanceOf reports how many bytes of rest the scanned token plus its
// terminator consumed (token then LF, CRLF, CR, or nothing at EOF).
func advanceOf(rest []byte, token string) int {
	n := len(token)
	switch {
	case n < len(rest) && rest[n] == '\n':
		return n + 1
	case n+1 < len(rest) && rest[n] == '\r' && rest[n+1] == '\n':
		return n + 2
	case n < len(rest) && rest[n] == '\r':
		return n + 1
	default:
		return n
	}
}

// corpus is the shared set of documents the census and locate tests sweep:
// plain grids, comments in every position, ragged widths, formulas, CR/CRLF
// terminators, unterminated finals, and empties.
var corpus = map[string]string{
	"plain":             "a\tb\nc\td\n",
	"empty":             "",
	"one unterminated":  "a\tb",
	"comments woven":    "#! shebang\n1\t=A1\n#. directive\n2\n# legacy\n3\t=sum(A1:A2)\t9\n",
	"shebang not first": "1\n#!x\n2\n",
	"crlf and lone cr":  "a\r\nb\rc\n",
	"ragged":            "a\nb\tc\td\ne\tf\n",
	"formulas only":     "=A1\t=B1\n=sum(A1:B1)\n",
	"trailing comment":  "1\n#. end\n",
	"empty lines":       "\n\n\n",
	"comment only":      "#. nothing else\n",
	"wide then narrow":  "a\tb\tc\td\te\nf\n",
	"tall":              "r0\nr1\n#. mid\nr2\nr3\nr4\nr5\nr6\nr7\nr8\nr9\n",
}

func TestScanCensusMatchesTheOracle(t *testing.T) {
	t.Parallel()

	for name, src := range corpus {
		for _, stride := range []int{1, 2, 3, 64} {
			t.Run(name+"/stride", func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, oracle(t, src).census, scan(t, src, stride).Census(),
					"src=%q stride=%d", src, stride)
			})
		}
	}
}

// TestCheckpointsHoldExactlyOnePerStartedStride pins the index's density — the
// memory bound striding exists for: one checkpoint per started stride of data
// rows, no more (a missed sinceMark reset checkpoints every row past the
// first stride) and no fewer (a doubled stride halves reachability).
func TestCheckpointsHoldExactlyOnePerStartedStride(t *testing.T) {
	t.Parallel()

	for name, src := range corpus {
		for _, stride := range []int{1, 2, 3, 64} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				truth := oracle(t, src)
				ix := scan(t, src, stride)
				assert.Equal(t, (truth.census.Rows+stride-1)/stride, ix.Checkpoints(),
					"src=%q stride=%d", src, stride)
			})
		}
	}
}

// TestLocateFindsACheckpointAtOrBeforeEveryRow pins Locate's contract: for
// every grid row of every corpus document, the returned checkpoint's grid-row
// count is at or before the requested row, its byte offset lands exactly on a
// physical line start, and scanning forward from it reaches the row within one
// stride's worth of lines.
func TestLocateFindsACheckpointAtOrBeforeEveryRow(t *testing.T) {
	t.Parallel()

	for name, src := range corpus {
		for _, stride := range []int{1, 2, 3, 64} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				truth := oracle(t, src)
				ix := scan(t, src, stride)
				for row := 0; row < truth.census.Rows; row++ {
					cp, ok := ix.Locate(index.GridRow(row))
					require.True(t, ok, "row %d must be locatable", row)
					assert.LessOrEqual(t, int(cp.Row), row, "checkpoint may not overshoot")
					assert.Less(t, row-int(cp.Row), stride,
						"forward scanning from the checkpoint is bounded by one stride of data lines")
					assert.Equal(t, truth.rowOffsets[cp.Row], int64(cp.Offset),
						"checkpoint offset must be the line start of its own grid row")
					assert.Equal(t, truth.rowLines[cp.Row], cp.Line, "checkpoint physical line")
					assert.Equal(t, truth.rowCells[cp.Row], cp.Cells, "cumulative cells before the checkpoint")
					assert.Equal(t, truth.rowFormulas[cp.Row], cp.Formulas, "cumulative formulas before the checkpoint")
				}
				_, ok := ix.Locate(index.GridRow(truth.census.Rows))
				assert.False(t, ok, "one past the last row is unlocatable")
			})
		}
	}
}

// TestScanRefusesAnOverlongLine pins the memory bound: a line past
// MaxLineBytes is a refusal, not an allocation.
func TestScanRefusesAnOverlongLine(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", 64) + "\n"
	_, err := index.Scan(strings.NewReader(long), index.SourceSize(len(long)), index.Options{
		Split:        splitTerminators,
		IsComment:    isComment,
		MaxLineBytes: 16,
	})
	assert.ErrorIs(t, err, index.ErrScan)
	assert.ErrorIs(t, err, bufio.ErrTooLong, "the cause survives the sentinel wrap")
}

// TestScanRefusesEachMissingRuleAlone pins the requirement one rule at a time:
// either injected rule missing — not only both — is ErrScan before any read.
func TestScanRefusesEachMissingRuleAlone(t *testing.T) {
	t.Parallel()

	_, err := index.Scan(strings.NewReader("x\n"), 2, index.Options{Split: splitTerminators})
	assert.ErrorIs(t, err, index.ErrScan, "missing IsComment alone is refused")
	_, err = index.Scan(strings.NewReader("x\n"), 2, index.Options{IsComment: isComment})
	assert.ErrorIs(t, err, index.ErrScan, "missing Split alone is refused")
	_, err = index.Scan(strings.NewReader("x\n"), -1, index.Options{Split: splitTerminators, IsComment: isComment})
	assert.ErrorIs(t, err, index.ErrScan, "a negative size is refused, never a silent empty index")
}

// never a bare error.
func TestScanRefusals(t *testing.T) {
	t.Parallel()

	_, err := index.Scan(strings.NewReader("x\n"), 2, index.Options{})
	assert.ErrorIs(t, err, index.ErrScan, "missing rules are refused up front")

	src := failingReader{prefix: []byte("a\nb\nc\nd\ne\nf\ng\nh\n")}
	_, err = index.Scan(src, 1<<20, index.Options{Split: splitTerminators, IsComment: isComment})
	assert.ErrorIs(t, err, index.ErrScan, "a mid-scan read failure surfaces as ErrScan")
}
