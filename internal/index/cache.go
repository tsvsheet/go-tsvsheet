// The reader's block cache: an LRU bounded in cells, block-granular. The
// block being read always survives eviction — it is the unit the answer is
// served from — so a capacity at or below one block degrades to exactly one
// resident block, never to thrash or refusal.
package index

// block is one cached run of rows starting at a checkpoint.
type block[T any] struct {
	rows  []T
	start GridRow
	cells CellCount
}

// CachedCells reports the cells currently resident — the number the capacity
// bounds.
func (r *Reader[T]) CachedCells() CellCount {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.cached
}

// block returns the cached block holding row, scanning it in on a miss and
// evicting least-recent blocks past the capacity. The caller's window is
// already clipped to the grid, so Locate cannot miss.
func (r *Reader[T]) block(row GridRow) (*block[T], error) {
	cp, _ := r.ix.Locate(row)
	if el, ok := r.blocks[cp.Row]; ok {
		r.order.MoveToFront(el)
		return el.Value.(*block[T]), nil
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
func (r *Reader[T]) evict() {
	for r.cached > r.capacity && r.order.Len() > 1 {
		el := r.order.Back()
		dropped := el.Value.(*block[T])
		r.order.Remove(el)
		delete(r.blocks, dropped.start)
		r.cached -= dropped.cells
	}
}
