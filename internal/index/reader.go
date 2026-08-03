// The row reader: grid rows on demand through the index, held in a cache
// bounded in cells. A block — the rows between two checkpoints — is the unit
// of scanning, caching, and eviction, so any row costs at most one stride of
// forward scanning from its checkpoint, and a small document ends up fully
// resident as a cache consequence, never as a mode (spec 016, ruling 1).
package index

import (
	"bufio"
	"container/list"
	"io"
	"strings"
	"sync"
)

// RowCount counts grid rows in a read request.
type RowCount int

// Reader reads grid rows through an index with a bounded block cache. It is a
// pointer type deliberately: the cache is stateful machinery guarded by a
// mutex (the TUI reads on a tick while scrolling), which a value copy would
// tear.
type Reader struct {
	src    io.ReaderAt
	blocks map[GridRow]*list.Element // checkpoint row → recency element
	order  *list.List                // front = most recent; values are *block
	opts   Options
	ix     Index

	mu       sync.Mutex
	size     SourceSize
	capacity CellCount
	cached   CellCount
}

// NewReader builds a Reader; ix is expected to have been scanned from the
// same bytes with the same opts (a source that has since changed surfaces as
// ErrScan when a block fails to cover its indexed rows). A capacity at or
// below zero degrades to a single resident block — the one being read; the
// eviction contract lives in cache.go.
func NewReader(src io.ReaderAt, size SourceSize, ix Index, opts Options, capacity CellCount) *Reader {
	return &Reader{
		src: src, size: size, ix: ix, opts: opts, capacity: capacity,
		blocks: make(map[GridRow]*list.Element), order: list.New(),
	}
}

// ReadRows returns the intersection of the window [from, from+n) with the
// grid — clipped on BOTH ends, so a negative start, a negative count, or a
// window past the grid is simply a smaller (possibly empty) result, never a
// panic or an oversized allocation. The returned rows are the caller's own
// copies: no later read, eviction, or caller write can alter them. A source
// failure — including a source whose bytes no longer match the index, which
// shows up as a block unable to cover its promised rows — surfaces as ErrScan.
func (r *Reader) ReadRows(from GridRow, n RowCount) ([][]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	row, end := clipWindow(from, n, RowCount(r.ix.Census().Rows))
	if row >= end {
		return nil, nil
	}
	out := make([][]string, 0, int(end-row))
	for row < end {
		b, err := r.block(row)
		if err != nil {
			return nil, err
		}
		copied := copyRows(*b, row, end)
		if len(copied) == 0 {
			return nil, ErrScan.With(nil,
				"reason", "block cannot cover its indexed rows; the source no longer matches the index",
				"row", int(row))
		}
		out = append(out, copied...)
		row += GridRow(len(copied))
	}
	return out, nil
}

// clipWindow intersects [from, from+n) with [0, rows). The end is computed as
// start plus the smaller of n and the rows remaining — subtraction of two
// small non-negatives — so a count near MaxInt64 clips instead of wrapping a
// sum negative (the adversary's boundary find: start+n overflowed and answered
// an empty window one count past where it answered the whole grid).
func clipWindow(from GridRow, n, rows RowCount) (GridRow, GridRow) {
	start := max(from, 0)
	if n <= 0 || RowCount(start) >= rows {
		return start, start
	}
	return start, start + GridRow(min(n, rows-RowCount(start)))
}

// copyRows copies the block's rows covering [row, end) — caller-owned copies,
// so no eviction, rescan, or caller write can reach cache memory.
func copyRows(b block, row, end GridRow) [][]string {
	var out [][]string
	for i := int(row - b.start); i >= 0 && i < len(b.rows) && row < end; i++ {
		out = append(out, append([]string(nil), b.rows[i]...))
		row++
	}
	return out
}

// scanBlock reads one block's rows from the checkpoint's byte offset: at most
// one stride of data lines, classified with the same injected rules the index
// was built with (the checkpoint carries the physical line number that keeps
// the first-line shebang rule honest).
func (r *Reader) scanBlock(cp Checkpoint) (*block, error) {
	out := &block{start: cp.Row}
	line := int(cp.Line) - 1
	scanner := bufio.NewScanner(io.NewSectionReader(r.src, int64(cp.Offset), int64(r.size)-int64(cp.Offset)))
	scanner.Buffer(nil, r.maxLine())
	scanner.Split(r.opts.Split)
	for len(out.rows) < int(r.stride()) && scanner.Scan() {
		line++
		text := scanner.Text()
		if r.opts.IsComment(LineNumber(line), []byte(text)) {
			continue
		}
		fields := strings.Split(text, "\t")
		out.rows = append(out.rows, fields)
		out.cells += CellCount(len(fields))
	}
	if err := scanner.Err(); err != nil {
		return nil, ErrScan.With(err)
	}
	return out, nil
}

// stride is the configured checkpoint interval with its default applied.
func (r *Reader) stride() StrideLines {
	if r.opts.Stride <= 0 {
		return DefaultStride
	}
	return r.opts.Stride
}

// maxLine is the configured line bound with its default applied.
func (r *Reader) maxLine() int {
	if r.opts.MaxLineBytes <= 0 {
		return DefaultMaxLineBytes
	}
	return r.opts.MaxLineBytes
}
