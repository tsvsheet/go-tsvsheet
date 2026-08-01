// What the other spreadsheet would compute.
//
// A divergence is only worth announcing when the two languages actually answer
// differently, and for an expression built from literals that is decidable:
// evaluate it under this language's reading and under Excel's, and compare. The
// alternative — matching the shape of the expression — announces `=+2^2`,
// `=-2^3` and `=-0^2`, all of which read the same in both, and a checker that
// cries wolf is one whose reader stops reading.
//
// Two details make the comparison honest rather than merely arithmetic:
//
//   - An overflow is an error, not a number. `=200^200` is #NUM! here and in
//     Excel, so two readings that both overflow AGREE even though the floats
//     they overflow to (+Inf and -Inf) differ. Comparing raw floats reported
//     `=-200^200` as divergent when both languages show the same error.
//   - The error propagates from the inside out. `=0^200^200` is #NUM! here
//     because 200^200 overflows before the outer power is reached, while Excel
//     groups it as (0^200)^200 and gets 0 — a real divergence that comparing
//     only the final values missed, because 0^(+Inf) is 0.
package engine

import (
	"math"

	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// reading is one language's answer for an expression: a number, or the error a
// spreadsheet would show in the cell instead.
type reading struct {
	value   floatVal
	isError bool
}

// numberRead is a computed number, or the error an infinity really represents.
// Every non-finite result reaches a cell as #NUM!, so it is folded to an error
// here rather than carried as a float nothing can compare meaningfully.
func numberRead(value floatVal) reading {
	if f := float64(value); math.IsNaN(f) || math.IsInf(f, 0) {
		return reading{isError: true}
	}
	return reading{value: value}
}

// errorRead is the reading of an expression that failed.
var errorRead = reading{isError: true}

// agrees reports whether two readings would look the same in a cell: the same
// number, or an error in both.
func (r reading) agrees(other reading) bool {
	if r.isError || other.isError {
		return r.isError && other.isError
	}
	return r.value == other.value
}

// raise lifts one reading to the power of another, propagating a failure from
// either side — the inside-out propagation a cell performs, so an intermediate
// overflow is still an error once the outer operator has been applied.
func raise(base, exponent reading) reading {
	if base.isError || exponent.isError {
		return errorRead
	}
	return numberRead(floatVal(math.Pow(float64(base.value), float64(exponent.value))))
}

// negate flips a reading's sign, leaving an error an error.
func negate(operand reading) reading {
	if operand.isError {
		return errorRead
	}
	return numberRead(-operand.value)
}

// precedenceAgrees reports whether a whole expression means the same under both
// languages' precedence. It is the only sound way to ask: judging one operator
// at a time answers a question nobody asked, because Excel's two differences
// interact. `2^-2^2` contains a sign over a power AND a chained power, each of
// which looks divergent alone, yet the expression is 0.0625 in both languages;
// `2^-1^3`, the same shape with one digit changed, is 0.5 here and 0.125 there.
// Only the whole expression distinguishes them.
//
// The comparison is decidable only for an expression built from literals. With
// a reference in it, nothing can be evaluated, so nothing is proven and the
// author is told — which is the safe direction for an advisory to err in.
func precedenceAgrees(expr tsvt.Expr) bool {
	ours, decidable := literalReadingOf(expr)
	if !decidable {
		return false
	}
	theirs, decidable := literalReadingOf(excelTree(expr))
	return decidable && ours.agrees(theirs)
}

// excelTree rewrites an expression into the tree Excel's precedence would have
// built from the same text. The two differences (SPECIFICATION §5.2) are that a
// sign binds tighter than `^` and that `^` groups leftward, so each is undone
// as a local rewrite; authored parentheses stop both, because a parenthesized
// expression is an atom to either language.
func excelTree(expr tsvt.Expr) tsvt.Expr {
	switch node := expr.(type) {
	case tsvt.Unary:
		return signBindsTighter(tsvt.Unary{Op: node.Op, X: excelTree(node.X), IsGrouped: node.IsGrouped})
	case tsvt.Percent:
		return tsvt.Percent{X: excelTree(node.X)}
	case tsvt.Binary:
		return groupsLeftward(tsvt.Binary{
			Op: node.Op, Left: excelTree(node.Left), Right: excelTree(node.Right), IsGrouped: node.IsGrouped,
		})
	default:
		return expr
	}
}

// signBindsTighter turns `-(x^y)` into `(-x)^y`, Excel's reading of `-x^y`.
func signBindsTighter(node tsvt.Unary) tsvt.Expr {
	power, isPower := node.X.(tsvt.Binary)
	if !isPower || power.Op != tsvt.OpPow || power.IsGrouped {
		return node
	}
	lifted := tsvt.Unary{Op: node.Op, X: power.Left}
	return groupsLeftward(tsvt.Binary{Op: tsvt.OpPow, Left: lifted, Right: power.Right, IsGrouped: node.IsGrouped})
}

// groupsLeftward turns `x^(y^z)` into `(x^y)^z`, Excel's reading of `x^y^z`.
func groupsLeftward(node tsvt.Binary) tsvt.Expr {
	inner, isPower := node.Right.(tsvt.Binary)
	if node.Op != tsvt.OpPow || !isPower || inner.Op != tsvt.OpPow || inner.IsGrouped {
		return node
	}
	rebased := groupsLeftward(tsvt.Binary{Op: tsvt.OpPow, Left: node.Left, Right: inner.Left})
	return groupsLeftward(tsvt.Binary{Op: tsvt.OpPow, Left: rebased, Right: inner.Right, IsGrouped: node.IsGrouped})
}

// literalReadingOf evaluates an expression built entirely from literals under
// THIS language's reading of the tree it is given, reporting whether it could be
// evaluated at all. Only the operators a precedence difference can reach are
// handled; anything else makes the expression undecidable rather than wrong.
func literalReadingOf(expr tsvt.Expr) (reading, bool) {
	switch node := expr.(type) {
	case tsvt.Number:
		value, ok := literalNumber(node)
		return numberRead(value), ok
	case tsvt.Unary:
		return unaryReading(node)
	case tsvt.Percent:
		operand, ok := literalReadingOf(node.X)
		return scaleDown(operand, percentDivisor), ok
	case tsvt.Binary:
		return binaryReading(node)
	default:
		return errorRead, false
	}
}

// unaryReading evaluates a signed literal expression.
func unaryReading(node tsvt.Unary) (reading, bool) {
	operand, ok := literalReadingOf(node.X)
	if node.Op == tsvt.OpNeg {
		return negate(operand), ok
	}
	return operand, ok
}

// binaryReading evaluates an arithmetic literal expression.
func binaryReading(node tsvt.Binary) (reading, bool) {
	left, leftOK := literalReadingOf(node.Left)
	right, rightOK := literalReadingOf(node.Right)
	if !leftOK || !rightOK {
		return errorRead, false
	}
	return applyArithmetic(node.Op, left, right)
}

// applyArithmetic applies one operator to two readings, reporting an operator
// this oracle does not model as undecidable.
func applyArithmetic(op tsvt.BinaryOp, left, right reading) (reading, bool) {
	switch op {
	case tsvt.OpPow:
		return raise(left, right), true
	case tsvt.OpMul:
		return combine(left, right, func(l, r floatVal) floatVal { return l * r }), true
	case tsvt.OpDiv:
		return quotient(left, right), true
	case tsvt.OpAdd:
		return combine(left, right, func(l, r floatVal) floatVal { return l + r }), true
	case tsvt.OpSub:
		return combine(left, right, func(l, r floatVal) floatVal { return l - r }), true
	default:
		return errorRead, false
	}
}

// combine applies a total arithmetic operation, propagating a failed operand.
func combine(left, right reading, op func(l, r floatVal) floatVal) reading {
	if left.isError || right.isError {
		return errorRead
	}
	return numberRead(op(left.value, right.value))
}

// quotient guards the zero denominator a cell reports as #DIV/0!.
func quotient(left, right reading) reading {
	if right.value == 0 {
		return errorRead
	}
	return combine(left, right, func(l, r floatVal) floatVal { return l / r })
}

// scaleDown divides a reading by a constant, as a postfix percent does.
func scaleDown(operand reading, divisor floatVal) reading {
	return combine(operand, numberRead(divisor), func(l, r floatVal) floatVal { return l / r })
}
