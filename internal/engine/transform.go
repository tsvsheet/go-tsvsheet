// Package engine's structural-edit algebra: how a row or column insert/delete
// maps every existing line index to its new one, expressed as pure transforms
// so the reference rewriting and the grid surgery share one definition.
package engine

import "github.com/tsvsheet/go-tsvsheet/internal/tsvt"

// lineIndex is a 0-based row or column position on one axis.
type (
	lineIndex int
	// axis reads and writes a CellRef's coordinate on one dimension as a 0-based
	// lineIndex, so the row and column structural edits share one implementation.
	axis struct {
		get func(tsvt.CellRef) lineIndex
		set func(tsvt.CellRef, lineIndex) tsvt.CellRef
	}
)

// transform shifts reference coordinates for one structural edit: point maps a
// single-cell reference (ok=false when it was deleted), and lo/hi map a range's
// two endpoints (a range collapses to #REF! when its endpoints cross).
type transform struct {
	point func(lineIndex) (lineIndex, boolResult)
	lo    func(lineIndex) (lineIndex, boolResult)
	hi    func(lineIndex) (lineIndex, boolResult)
}

// shiftUp maps a coordinate for an insertion before `at`: coordinates at or
// after `at` move up one; nothing is ever deleted.
func shiftUp(at lineIndex) func(lineIndex) (lineIndex, boolResult) {
	return func(x lineIndex) (lineIndex, boolResult) {
		if x >= at {
			return x + 1, true
		}
		return x, true
	}
}

// insertTransform shifts every reference on the axis down by one at `at`.
func insertTransform(at lineIndex) transform {
	up := shiftUp(at)
	return transform{point: up, lo: up, hi: up}
}

// deletePoint maps a single-cell coordinate for a deletion at `at`: the deleted
// line yields ok=false (#REF!); lines after it move up one.
func deletePoint(at lineIndex) func(lineIndex) (lineIndex, boolResult) {
	return func(x lineIndex) (lineIndex, boolResult) {
		if x == at {
			return 0, false
		}
		if x > at {
			return x - 1, true
		}
		return x, true
	}
}

// deleteLo maps a range's low endpoint for a deletion: it clamps to `at` (the
// line that slides into the deleted slot), so the range's start survives.
func deleteLo(at lineIndex) func(lineIndex) (lineIndex, boolResult) {
	return func(x lineIndex) (lineIndex, boolResult) {
		if x > at {
			return x - 1, true
		}
		return x, true
	}
}

// deleteHi maps a range's high endpoint for a deletion: it clamps to `at`-1 (the
// last line before the deleted slot); combined with deleteLo, a range that was
// exactly the deleted line collapses (lo > hi → #REF!).
func deleteHi(at lineIndex) func(lineIndex) (lineIndex, boolResult) {
	return func(x lineIndex) (lineIndex, boolResult) {
		if x >= at {
			return x - 1, true
		}
		return x, true
	}
}

// deleteTransform shifts references on the axis up by one past `at`, turning a
// reference to the deleted line into #REF!.
func deleteTransform(at lineIndex) transform {
	return transform{point: deletePoint(at), lo: deleteLo(at), hi: deleteHi(at)}
}
