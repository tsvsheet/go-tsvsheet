package engine

import (
	"strconv"
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// evalCriteria dispatches the conditional-aggregate builtins, which pair a range
// with a criterion. ok is false for any other name.
func (r resolver) evalCriteria(name funcName, args []tsvt.Expr) (Value, boolResult) {
	switch name {
	case fnCountif:
		return r.criteriaCount(args), true
	case "sumif":
		return r.criteriaSum(args, false), true
	case "averageif":
		return r.criteriaSum(args, true), true
	default:
		return Value{}, false
	}
}

// isCriteria reports whether name is one of the conditional-aggregate builtins.
func isCriteria(name funcName) boolResult { return isAmong(name, criteriaFns) }

// criteriaFns are the lazily-dispatched criteria builtins.
// fnCountif names the counting criteria builtin.
const fnCountif = "countif"

var criteriaFns = []funcName{fnCountif, "sumif", "averageif"}

// criteriaCount implements COUNTIF(range, criterion).
func (r resolver) criteriaCount(args []tsvt.Expr) Value {
	if len(args) != 2 {
		return errorValue(ErrValue)
	}
	m, refused := r.argMatrix(args[0])
	if refused.isError() {
		return refused
	}
	cells := flatten1D(m)
	crit := r.eval(args[1])
	count := 0
	for _, cell := range cells {
		if matchesCriterion(cell, crit) {
			count++
		}
	}
	return numberValue(floatVal(count))
}

// criteriaRanges resolves a criteria call's cell range (args[0]) and optional
// sum range (args[2], defaulting to the cell range), flattened row-major; a
// wholesale refusal of either range propagates.
func (r resolver) criteriaRanges(args []tsvt.Expr) (cells, sumCells []Value, refused Value) {
	m, refused := r.argMatrix(args[0])
	if refused.isError() {
		return nil, nil, refused
	}
	cells = flatten1D(m)
	sumCells = cells
	if len(args) == 3 {
		sm, sumRefused := r.argMatrix(args[2])
		if sumRefused.isError() {
			return nil, nil, sumRefused
		}
		sumCells = flatten1D(sm)
	}
	return cells, sumCells, Value{}
}

// criteriaSum implements SUMIF/AVERAGEIF(range, criterion, [sumRange]); when a
// sum range is given the matching positions are summed there.
func (r resolver) criteriaSum(args []tsvt.Expr, isAverage boolResult) Value {
	if len(args) < 2 || len(args) > 3 {
		return errorValue(ErrValue)
	}
	cells, sumCells, refused := r.criteriaRanges(args)
	if refused.isError() {
		return refused
	}
	total, matched, bad := foldMatches(cells, sumCells, r.eval(args[1]))
	if bad.isError() {
		return bad
	}
	if isAverage {
		if matched == 0 {
			return errorValue(ErrDiv)
		}
		return numberValue(floatVal(total / float64(matched)))
	}
	return numberValue(floatVal(total))
}

// foldMatches sums the sumCells at positions whose cells match the criterion,
// reporting the total, the match count, and any error operand in the sum.
func foldMatches(cells, sumCells []Value, criterion Value) (float64, int, Value) {
	total := 0.0
	matched := 0
	for i, cell := range cells {
		if !matchesCriterion(cell, criterion) || i >= len(sumCells) {
			continue
		}
		n, bad := sumCells[i].asNumber()
		if bad.isError() {
			return 0, 0, bad
		}
		total += n
		matched++
	}
	return total, matched, Value{}
}

// matchesCriterion tests a cell against a criterion value: a bare value matches
// by equality; a value prefixed with a comparison operator (>, <, >=, <=, <>, =)
// matches numerically.
func matchesCriterion(cell, criterion Value) boolResult {
	op, operand := parseCriterion(textVal(criterion.String()))
	if op == "" {
		return equalValues(cell, value(operand))
	}
	cellNum, cellBad := cell.asNumber()
	operandNum, err := strconv.ParseFloat(string(operand), 64)
	if cellBad.isError() || err != nil {
		return false
	}
	return boolResult(numberOrder(op, floatVal(cellNum), floatVal(operandNum)))
}

// parseCriterion splits a leading comparison operator from a criterion string;
// the operator is a tsvt.BinaryOp (empty for a bare equality match).
func parseCriterion(crit textVal) (tsvt.BinaryOp, textVal) {
	for _, op := range []tsvt.BinaryOp{tsvt.OpGe, tsvt.OpLe, tsvt.OpNe, tsvt.OpGt, tsvt.OpLt, tsvt.OpEq} {
		if rest, ok := strings.CutPrefix(string(crit), string(op)); ok {
			return op, textVal(rest)
		}
	}
	return "", crit
}
