// Package index builds a strided line index over raw .tsvt bytes, the seam the
// indexed document stands on (spec 016): one sequential scan records every Kth
// physical line's byte offset together with the running grid-row and cell
// counts, so any grid row is reachable by a binary search plus at most one
// stride of forward scanning, and the census — rows, cells, formulas, width —
// is known without materializing anything.
//
// The package is deliberately language-blind: the two .tsvt rules it needs —
// how a line terminates and which lines are comments — are injected by the
// caller (the engine hands in its own single definitions), so the rules keep
// exactly one home and this package needs none of the engine's types.
package index

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"sort"

	errs "github.com/gomatic/go-error"
)

// ErrScan wraps a failure while scanning the source: a read error, or a line
// past the configured byte bound.
const ErrScan errs.Const = "failed to scan the source"

type (
	// StrideLines is the checkpoint interval in physical lines: smaller finds
	// rows with less forward scanning, larger keeps the index smaller.
	StrideLines int

	// GridRow is a 0-based data-row position — the A1 row coordinate space,
	// which skips comment lines.
	GridRow int

	// LineNumber is a 1-based physical line position, the coordinate space
	// comment classification speaks.
	LineNumber int

	// ByteOffset is a position in the source bytes.
	ByteOffset int64

	// SourceSize is the source's total byte length.
	SourceSize int64

	// CellCount counts cells (fields), cumulatively or as a width.
	CellCount int64

	// FormulaCount counts `=`-cells, cumulatively.
	FormulaCount int64

	// CommentRule reports whether the 1-based physical line with the given
	// text is a comment line (skipped by the grid). Injected — the engine owns
	// the one definition.
	CommentRule func(line LineNumber, text []byte) bool
)

// Options configures a Scan. Split and IsComment are required — they are the
// injected language rules; Stride and MaxLineBytes fall back to defaults.
type Options struct {
	Split        bufio.SplitFunc
	IsComment    CommentRule
	Stride       StrideLines
	MaxLineBytes int
}

// Default bounds: a 256-line stride keeps a 50M-line index in single-digit
// megabytes; the 1 MiB line bound mirrors the engine's own scanner cap.
const (
	DefaultStride       StrideLines = 256
	DefaultMaxLineBytes             = 1 << 20
)

// Checkpoint is one strided record: the byte offset of a physical line start,
// which line and grid row it is, and the cumulative cell and formula counts
// before it. Cumulative counters make per-block questions (does this stride
// hold formulas? how many cells to here?) subtractions between neighbors.
//
// A checkpoint is only recorded AT a data line, so Offset is always the line
// start of grid row Row — the reader seeks here and the first line it scans is
// exactly that row.
type Checkpoint struct {
	Offset   ByteOffset
	Line     LineNumber
	Row      GridRow
	Cells    CellCount
	Formulas FormulaCount
}

// Census is what one scan learns about the whole document — the numbers every
// policy decision (editable? computable? how lazily?) reads.
type Census struct {
	Lines    int
	Rows     int
	MaxWidth int
	Cells    CellCount
	Formulas FormulaCount
}

// Index is the immutable scan result: the checkpoint table and the census.
type Index struct {
	stride []Checkpoint
	census Census
}

// Census returns the document's totals.
func (ix Index) Census() Census { return ix.census }

// Checkpoints reports the stride table's length — the index's memory, which
// striding exists to bound: one checkpoint per started stride of data rows
// (pinned by TestCheckpointsHoldExactlyOnePerStartedStride).
func (ix Index) Checkpoints() int { return len(ix.stride) }

// Locate returns the last checkpoint at or before row — the seek target from
// which forward scanning reaches the row within one stride. ok is false for a
// row past the grid.
func (ix Index) Locate(row GridRow) (Checkpoint, bool) {
	if int(row) < 0 || int(row) >= ix.census.Rows {
		return Checkpoint{}, false
	}
	at := sort.Search(len(ix.stride), func(i int) bool { return ix.stride[i].Row > row })
	return ix.stride[at-1], true
}

// Scan reads the source once, sequentially, and builds the index. It never
// retains source bytes: memory is the checkpoint table alone. Both language
// rules are required, and the injected Split must emit a token for every byte
// it consumes and consume at least one byte with every token: bufio's skip
// form (advance without a token) is REFUSED, because at EOF bufio treats any
// nil-token return as end-of-scan and silently drops the remainder — the
// census would then depend on how the reader chunks — and a token without
// progress would loop forever. Both defects surface as ErrScan, never as a
// wrong index or a hang.
func Scan(r io.ReaderAt, size SourceSize, opts Options) (Index, error) {
	if reason, refused := scanRefusal(size, opts); refused {
		return Index{}, ErrScan.With(nil, "reason", reason)
	}
	state := newScanState(opts)
	advance := 0
	scanner := bufio.NewScanner(io.NewSectionReader(r, 0, int64(size)))
	// A nil initial buffer, deliberately: bufio takes the LARGER of the buffer
	// capacity and max, so any preallocated capacity above maxLine would defeat
	// the bound.
	scanner.Buffer(nil, state.maxLine)
	scanner.Split(func(data []byte, isAtEOF bool) (int, []byte, error) {
		advanced, token, err := checkedSplit(opts.Split, data, isEOF(isAtEOF))
		if err == nil && token != nil {
			advance = advanced // the token plus its terminator: this line's true byte length
		}
		return advanced, token, err
	})
	for scanner.Scan() {
		state = state.line(scanner.Bytes(), advance)
	}
	if err := scanner.Err(); err != nil {
		return Index{}, scanFailure(err)
	}
	return Index{stride: state.stride, census: state.census}, nil
}

// scanFailure wraps a scanner error as ErrScan once: a checkedSplit refusal
// already carries the sentinel and passes through unstuttered.
func scanFailure(err error) error {
	if errors.Is(err, ErrScan) {
		return err
	}
	return ErrScan.With(err)
}

// scanRefusal names the up-front refusal, if any: a negative size, or a
// missing injected rule.
func scanRefusal(size SourceSize, opts Options) (string, bool) {
	switch {
	case size < 0:
		return "negative source size", true
	case opts.Split == nil || opts.IsComment == nil:
		return "Split and IsComment are required", true
	default:
		return "", false
	}
}

// isEOF reports whether the scanner has reached the end of its input.
type isEOF bool

// checkedSplit runs the injected Split and refuses its two defect shapes: the
// chunk-dependent skip form (advance without a token — bufio silently drops
// the remainder at EOF), and a token without progress, at EOF included (bufio
// panics after a hundred of those). Each surfaces as ErrScan — pinned by
// TestCheckedSplitNeverAdmitsADefectiveRule. bufio's ErrFinalToken form
// passes through on the error branch.
func checkedSplit(split bufio.SplitFunc, data []byte, isAtEnd isEOF) (int, []byte, error) {
	advanced, token, err := split(data, bool(isAtEnd))
	switch {
	case err != nil:
	case token == nil && advanced > 0:
		return 0, nil, ErrScan.With(nil, "reason", "skip-form Split is unsupported")
	case token != nil && advanced == 0:
		return 0, nil, ErrScan.With(nil, "reason", "Split returned a token without consuming input")
	}
	return advanced, token, err
}

// scanState is one Scan pass's fold state — the growing checkpoint table, the
// census counters, and the byte cursor — carried by value: each line folds to
// the next state, so nothing is mutated through a shared pointer.
type scanState struct {
	stride    []Checkpoint
	opts      Options
	census    Census
	offset    ByteOffset
	maxLine   int
	sinceMark int
}

// newScanState applies the option defaults.
func newScanState(opts Options) scanState {
	if opts.Stride <= 0 {
		opts.Stride = DefaultStride
	}
	maxLine := opts.MaxLineBytes
	if maxLine <= 0 {
		maxLine = DefaultMaxLineBytes
	}
	return scanState{opts: opts, maxLine: maxLine}
}

// line folds one physical line into the next state: classification happens
// while offset names the line's first byte (the checkpoint anchor), then the
// cursor advances by the line's true consumed length — token plus terminator,
// as the injected Split reported it.
func (s scanState) line(text []byte, advance int) scanState {
	s.census.Lines++
	if !s.opts.IsComment(LineNumber(s.census.Lines), text) {
		s = s.dataLine(text)
	}
	s.offset += ByteOffset(advance)
	return s
}

// dataLine folds one grid row: checkpointing at the stride, then counting its
// cells and formulas.
func (s scanState) dataLine(text []byte) scanState {
	if s.sinceMark == 0 || s.sinceMark >= int(s.opts.Stride) {
		s.stride = append(s.stride, Checkpoint{
			Offset:   s.offset,
			Line:     LineNumber(s.census.Lines),
			Row:      GridRow(s.census.Rows),
			Cells:    s.census.Cells,
			Formulas: s.census.Formulas,
		})
		s.sinceMark = 0
	}
	s.sinceMark++
	s.census.Rows++
	width := bytes.Count(text, []byte{'\t'}) + 1
	if width > s.census.MaxWidth {
		s.census.MaxWidth = width
	}
	s.census.Cells += CellCount(width)
	s.census.Formulas += countFormulas(text)
	return s
}

// countFormulas counts the fields of one line that begin with `=`.
func countFormulas(text []byte) FormulaCount {
	var n FormulaCount
	rest := text
	for {
		if len(rest) > 0 && rest[0] == '=' {
			n++
		}
		tab := bytes.IndexByte(rest, '\t')
		if tab < 0 {
			return n
		}
		rest = rest[tab+1:]
	}
}
