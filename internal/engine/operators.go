// Package engine's operator layer: the arithmetic and comparison semantics
// behind the §5.2 binary operators, kept apart from the function registry that
// dispatches named calls.
package engine

import (
	"math"
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// mod is Excel's MOD remainder: it takes the sign of the divisor via floored
// division (MOD(n,d) = n - d*FLOOR(n/d)), unlike Go's math.Mod, which takes the
// sign of the dividend. So MOD(-3,2) is 1 (not -1) and MOD(3,-2) is -1 (not 1).
func mod(l, r floatVal) floatVal {
	return floatVal(float64(l) - float64(r)*math.Floor(float64(l)/float64(r)))
}

// power raises l to the r-th power.
func power(l, r floatVal) floatVal { return floatVal(math.Pow(float64(l), float64(r))) }

// compare applies a comparison, yielding a boolean TRUE/FALSE (ADR 0004 §1):
// numeric when both operands are numeric (a bool compares as its 1/0), and
// lexicographic when both are strings; a mixed pair is #VALUE!.
func compare(op tsvt.BinaryOp, left, right Value) Value {
	if numericish(left) && numericish(right) {
		return boolValue(boolResult(numberOrder(op, floatVal(left.num), floatVal(right.num))))
	}
	if bothText(left, right) {
		return boolValue(boolResult(stringOrder(op, textVal(text(left)), textVal(text(right)))))
	}
	return errorValue(ErrValue)
}

// numericish reports whether a value participates in numeric comparison — a
// number or a boolean (whose 1/0 lives in the number field).
func numericish(v Value) bool {
	return v.kind == kindNumber || v.kind == kindBool || v.kind == kindDate
}

// bothText reports whether both operands compare as text (string or empty).
func bothText(left, right Value) bool {
	return textual(left) && textual(right)
}

// textual reports whether a value participates in string comparison.
func textual(v Value) bool {
	switch v.kind {
	case kindString, kindEmpty:
		return true
	default:
		return false
	}
}

// text is a value's comparable string form (empty for the empty value).
func text(v Value) string {
	switch v.kind {
	case kindString:
		return v.str
	default:
		return ""
	}
}

// numberOrder evaluates a comparison over two numbers.
func numberOrder(op tsvt.BinaryOp, l, r floatVal) bool {
	switch op {
	case tsvt.OpEq:
		return l == r
	case tsvt.OpNe:
		return l != r
	case tsvt.OpLt:
		return l < r
	case tsvt.OpLe:
		return l <= r
	case tsvt.OpGt:
		return l > r
	default: // OpGe
		return l >= r
	}
}

// stringOrder evaluates a comparison over two strings lexicographically.
func stringOrder(op tsvt.BinaryOp, l, r textVal) bool {
	return numberOrder(op, floatVal(strings.Compare(string(l), string(r))), 0)
}
