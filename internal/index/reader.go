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

// block is one cached run of rows starting at a checkpoint.
type block struct {
	rows  [][]string
	start GridRow
	cells CellCount
}

// NewReader builds a Reader over src using ix (which must have been scanned
// from the same bytes with the same opts) and a cache bounded to capacity
// cells.
func NewReader(src io.ReaderAt, size SourceSize, ix Index, opts Options, capacity CellCount) *Reader {
	return &Reader{
		src: src, size: size, ix: ix, opts: opts, capacity: capacity,
		blocks: make(map[GridRow]*list.Element), order: list.New(),
	}
}

// CachedCells reports the cells currently resident — the number the capacity
// bounds.
func (r *Reader) CachedCells() CellCount {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.cached
}

// ReadRows returns rows [from, from+n), clipped at the grid's end; a window
// past the grid is empty. A source failure surfaces as ErrScan.
func (r *Reader) ReadRows(from GridRow, n RowCount) ([][]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([][]string, 0, n)
	row := from
	for RowCount(len(out)) < n && int(row) < r.ix.Census().Rows {
		b, err := r.block(row)
		if err != nil {
			return nil, err
		}
		for i := int(row - b.start); i < len(b.rows) && RowCount(len(out)) < n; i++ {
			out = append(out, b.rows[i])
			row++
		}
	}
	return out, nil
}

// block returns the cached block holding row, scanning it in on a miss and
// evicting least-recent blocks past the capacity.
func (r *Reader) block(row GridRow) (*block, error) {
	cp, _ := r.ix.Locate(row) // row < Rows is the caller's loop guard
	if el, ok := r.blocks[cp.Row]; ok {
		r.order.MoveToFront(el)
		return el.Value.(*block), nil
	}
	loaded, err := r.scanBlock(cp)
	if err != nil {
		return nil, err
	}
	r.blocks[cp.Row] = r.order.PushFront(loaded)
	r.cached += loaded.cells
	r.evict()
	return loaded, nil
}

// evict drops least-recent blocks until the cache is within capacity, always
// keeping the most recent block resident — it is the one being read.
func (r *Reader) evict() {
	for r.cached > r.capacity && r.order.Len() > 1 {
		el := r.order.Back()
		dropped := el.Value.(*block)
		r.order.Remove(el)
		delete(r.blocks, dropped.start)
		r.cached -= dropped.cells
	}
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
