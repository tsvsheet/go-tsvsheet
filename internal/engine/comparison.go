// Comparison, where the two languages part company three ways.
//
// This language compares by value and by type: text against text by code point,
// a boolean as 1 or 0, and a mixed comparison not at all. Excel compares
// everything, in a total order that puts every number below every text and every
// text below every boolean, and its text comparison ignores case.
//
// So the same three-operator expression can read differently for three
// unrelated reasons, and each is announced only where the two orderings actually
// produce different answers — which is far less often than the shapes suggest.
// `=TRUE<0` is FALSE under 1/0 folding and FALSE under Excel's ranking; the
// reasons differ and the answer does not, so there is nothing to tell anybody.
package engine

import (
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// comparisonDivergence classifies a comparison whose operands are literals —
// the only case a divergence is provable without computing the sheet. Text
// against a number or a boolean is #VALUE! here and a comparison in Excel; text
// against text is compared by case here and without regard to it there.
func comparisonDivergence(node tsvt.Binary) (divergenceMessage, bool) {
	left, isLeftText := stringLiteral(node.Left)
	right, isRightText := stringLiteral(node.Right)
	switch {
	case isLeftText && isRightText:
		return msgTextComparedByCase, textComparisonTurnsOnCase(node.Op, left, right)
	case isLeftText:
		return msgTextInComparison, isValueLiteral(node.Right)
	case isRightText:
		return msgTextInComparison, isValueLiteral(node.Left)
	default:
		return msgBooleanComparedAsNumber, booleanComparisonTurnsOnRank(node)
	}
}

// booleanComparisonTurnsOnRank reports whether a boolean-against-number
// comparison is answered differently by the two languages. It is not enough
// that the shape occurs: this language folds a boolean to 1 or 0 and Excel ranks
// every boolean above every number, and those two rules happen to agree far
// more often than they differ — `=TRUE<0` is FALSE either way. So both readings
// are computed and compared.
func booleanComparisonTurnsOnRank(node tsvt.Binary) bool {
	left, isLeftBool := booleanLiteral(node.Left)
	right, isRightBool := booleanLiteral(node.Right)
	number, hasNumber := literalNumber(node.Right)
	if isRightBool {
		number, hasNumber = literalNumber(node.Left)
	}
	if isLeftBool == isRightBool || !hasNumber {
		return false
	}
	folded, ranked := foldedOrder(booleanAsNumber(left || right), number, boolResult(isLeftBool))
	return applyOrder(node.Op, folded) != applyOrder(node.Op, ranked)
}

// foldedOrder is the ordering of a boolean against a number under each
// language's rule: this language compares the boolean's 1-or-0 value, while
// Excel puts every boolean above every number regardless of magnitude.
func foldedOrder(value, number floatVal, isBoolOnLeft boolResult) (folded, ranked rankOrder) {
	folded, ranked = orderOf(value, number), 1
	if !isBoolOnLeft {
		folded, ranked = orderOf(number, value), -1
	}
	return folded, ranked
}

// booleanAsNumber is what this language compares a boolean as: 1 or 0.
func booleanAsNumber(isTrue boolResult) floatVal {
	if isTrue {
		return 1
	}
	return 0
}

// orderOf ranks two numbers as -1, 0 or +1.
func orderOf(left, right floatVal) rankOrder {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

// booleanLiteral reports a node's truth when it is a boolean literal.
func booleanLiteral(node tsvt.Expr) (boolResult, bool) {
	lit, ok := node.(tsvt.BoolLit)
	return boolResult(lit.IsTrue), ok
}

// textComparisonTurnsOnCase reports whether two literals compare one way under
// this language's case-sensitive ordering and the other way under Excel's
// case-blind one — the exact condition under which the two disagree, so a
// comparison they both answer the same (`"a" < "b"`) is never reported.
func textComparisonTurnsOnCase(op tsvt.BinaryOp, left, right literalText) bool {
	sensitive := compareText(op, left, right)
	blind := compareText(op, left.folded(), right.folded())
	return sensitive != blind
}

// compareText applies a comparison operator to two strings by byte order, the
// ordering the evaluator uses.
func compareText(op tsvt.BinaryOp, left, right literalText) bool {
	return applyOrder(op, rankOrder(strings.Compare(string(left), string(right))))
}

// rankOrder is how two operands rank: -1 when the left ranks below the right,
// 0 when they rank alike, +1 above.
type rankOrder int

// applyOrder answers a comparison operator from the ordering of its operands:
// Every divergence here is decided by asking this twice — once with this
// language's ordering and once with Excel's — so a finding is raised only where
// the two answers genuinely differ, rather than wherever a construct happens to
// look suspicious.
func applyOrder(op tsvt.BinaryOp, order rankOrder) bool {
	switch op {
	case tsvt.OpEq:
		return order == 0
	case tsvt.OpNe:
		return order != 0
	case tsvt.OpLt:
		return order < 0
	case tsvt.OpLe:
		return order <= 0
	case tsvt.OpGt:
		return order > 0
	default: // tsvt.OpGe
		return order >= 0
	}
}

// literalText is the text of a quoted literal.
type literalText string

// folded is the operand as Excel sees it for comparison: without case.
func (t literalText) folded() literalText { return literalText(strings.ToLower(string(t))) }
