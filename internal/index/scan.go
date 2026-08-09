package index

import (
	"bufio"
	"bytes"
	"errors"
)

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
	stride          []Checkpoint
	marks           []GridRow
	opts            Options
	census          Census
	offset          ByteOffset
	maxLine         int
	sinceMark       int
	hasMarkOverflow bool
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
	s = s.markLine(text)
	s.census.Rows++
	width := bytes.Count(text, []byte{'\t'}) + 1
	if width > s.census.MaxWidth {
		s.census.MaxWidth = width
	}
	s.census.Cells += CellCount(width)
	s.census.Formulas += countFormulas(text)
	return s
}

// markLine records a data line containing the requested Mark, or the overflow
// once the cap is passed. The overflow drops the whole set, not just the tail:
// a truncated set would answer "which rows are marked" with a silent lie,
// where the flag makes the refusal inspectable.
func (s scanState) markLine(text []byte) scanState {
	if len(s.opts.Mark) == 0 || s.hasMarkOverflow || !bytes.Contains(text, s.opts.Mark) {
		return s
	}
	if s.opts.MaxMarks > 0 && CellCount(len(s.marks)) >= s.opts.MaxMarks {
		s.marks, s.hasMarkOverflow = nil, true
		return s
	}
	s.marks = append(s.marks, GridRow(s.census.Rows))
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
