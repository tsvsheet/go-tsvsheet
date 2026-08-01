// The Excel divergences, announced.
//
// Five of this language's behaviours differ from Excel deliberately
// (SPECIFICATION §5.2): `^` binds tighter than a unary sign and associates to
// the right, where Excel does the opposite of both; arithmetic and comparison
// refuse text where Excel coerces it; and text comparison is case-sensitive
// where Excel's ignores case. Each is a defensible choice, and each is a place
// a spreadsheet user's muscle memory produces a value they did not mean.
//
// So the language says so. Check reports every occurrence at authoring time,
// naming what Excel would do and what to write instead. A divergence a tool
// announces is a lesson; one it hides is a bug.
//
// Two constraints keep this honest, because an advisory nobody trusts is worse
// than none:
//
//   - Nothing is reported unless it is provable from the text alone. Whether a
//     *reference* holds text is a runtime fact, so `A1="yes"` is not flagged —
//     it diverges only for some contents, and a checker that fired on the most
//     common formula in any spreadsheet would train its reader to ignore it.
//     Where the operands are literals, the divergence is proven, not guessed.
//   - Authored parentheses silence the precedence findings. `-(A1^2)` is the
//     remedy this file recommends; flagging it would make the advice impossible
//     to take. See tsvt.Binary.IsGrouped.
//
// The findings are advisory: the sheet computes exactly as it did before they
// existed. The point is that the author is told once, while writing, instead of
// by a wrong total months later.
package engine

import (
	"strconv"
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// divergenceDiagnostics reports every construct in one formula whose meaning
// differs from the same text typed into Excel.
func divergenceDiagnostics(expr tsvt.Expr, at Address) []Diagnostic {
	label := at.String()
	provenAlike := precedenceAgrees(expr)
	var diags []Diagnostic
	walkExprs(expr, func(node tsvt.Expr) {
		message, found := divergence(node)
		if found && (!provenAlike || !message.isAboutPrecedence()) {
			diags = append(diags, Diagnostic{Cell: label, Message: string(message)})
		}
	})
	return diags
}

// isAboutPrecedence reports whether a finding is one the whole-expression proof
// can retire. The precedence findings are located per node but decided for the
// expression as a whole, because Excel's two precedence differences interact;
// the coercion and comparison findings are local and stand on their own.
func (m divergenceMessage) isAboutPrecedence() bool {
	return m == msgSignBindsLooser || m == msgPowerAssociates
}

// walkExprs visits every node of an expression, parents before children.
func walkExprs(expr tsvt.Expr, visit func(tsvt.Expr)) {
	visit(expr)
	for _, child := range children(expr) {
		walkExprs(child, visit)
	}
}

// divergenceMessage is one explanation an author reads.
type divergenceMessage string

// The explanations. Each names the reading this language takes, the reading
// Excel takes, and what to write to mean either — so an author who wants one or
// the other can say which without knowing a precedence table.
const (
	msgSignBindsLooser divergenceMessage = "unary sign binds looser than ^ here, so this reads -(x^y); " +
		"Excel reads (-x)^y — write -(x^y) or (-x)^y to say which"
	msgPowerAssociates divergenceMessage = "^ associates to the right here, so this reads x^(y^z); " +
		"Excel reads (x^y)^z — parenthesize to say which"
	msgTextInArithmetic divergenceMessage = "text in arithmetic is #VALUE! here; Excel would coerce it to a number — " +
		"write the number itself, or value(text)"
	msgTextInComparison divergenceMessage = "comparing text with a non-text value is #VALUE! here; " +
		"Excel would compare them — write the number itself, or value(text)"
	msgTextComparedByCase divergenceMessage = "text comparison is case-sensitive here and this comparison turns on " +
		"case; Excel ignores case — wrap both sides in lower() to match Excel, or exact() to mean case-sensitive"
	msgBooleanComparedAsNumber divergenceMessage = "a boolean compares as 1 or 0 here; Excel orders every boolean " +
		"above every number, so it reads this the other way — compare against TRUE or FALSE to say which you mean"
)

// divergence classifies one node, reporting the explanation when Excel reads
// that node differently.
func divergence(node tsvt.Expr) (divergenceMessage, bool) {
	switch typed := node.(type) {
	case tsvt.Unary:
		return unaryDivergence(typed)
	case tsvt.Percent:
		return msgTextInArithmetic, isCoercibleText(typed.X)
	case tsvt.Binary:
		return binaryDivergence(typed)
	default:
		return "", false
	}
}

// unaryDivergence classifies a signed operand. Only a minus is reported over a
// power: `+x` is `x` under either reading, so `=+2^2` is 4 in both languages and
// announcing it would be noise.
func unaryDivergence(node tsvt.Unary) (divergenceMessage, bool) {
	if node.Op == tsvt.OpNeg && isPower(node.X) {
		return msgSignBindsLooser, true
	}
	return msgTextInArithmetic, isCoercibleText(node.X)
}

// binaryDivergence classifies a binary node: a right-nested power, text where
// arithmetic refuses it, or a comparison Excel would read differently.
func binaryDivergence(node tsvt.Binary) (divergenceMessage, bool) {
	switch {
	case node.Op == tsvt.OpPow && isPower(node.Right):
		return msgPowerAssociates, true
	case isArithmetic(node.Op) && hasCoercibleTextOperand(node):
		return msgTextInArithmetic, true
	case isComparison(node.Op):
		return comparisonDivergence(node)
	default:
		return "", false
	}
}

// literalNumber reports a node's value when it is a numeric literal, seeing
// through the signs, percents and parentheses an author may have written around
// it, so `-3`, `+3`, `(-3)` and `100%` are all recognised as numbers.
func literalNumber(node tsvt.Expr) (floatVal, bool) {
	switch typed := node.(type) {
	case tsvt.Number:
		value, err := strconv.ParseFloat(typed.Text, 64)
		return floatVal(value), err == nil
	case tsvt.Unary:
		value, ok := literalNumber(typed.X)
		if typed.Op == tsvt.OpNeg {
			return -value, ok
		}
		return value, ok
	case tsvt.Percent:
		value, ok := literalNumber(typed.X)
		return value / percentDivisor, ok
	default:
		return 0, false
	}
}

// isPower reports whether a node is a `^` operation the author left ungrouped.
func isPower(node tsvt.Expr) bool {
	binary, ok := node.(tsvt.Binary)
	return ok && binary.Op == tsvt.OpPow && !binary.IsGrouped
}

// hasCoercibleTextOperand reports whether either operand is text Excel would
// read as a number. `="abc"+1` is #VALUE! in Excel too, so the two languages
// agree and there is nothing to announce.
func hasCoercibleTextOperand(node tsvt.Binary) bool {
	return isCoercibleText(node.Left) || isCoercibleText(node.Right)
}

// isCoercibleText reports whether a node is a text literal naming a number.
func isCoercibleText(node tsvt.Expr) bool {
	text, ok := stringLiteral(node)
	return ok && text.namesANumber()
}

// namesANumber reports whether Excel would read this text as a number. Excel
// accepts surrounding space, a currency prefix, and a percent suffix; anything
// else is text to both languages.
//
// The spellings are Excel's, not Go's. ParseFloat alone would accept "Inf",
// "NaN", "0x1p4" and "1_0" — none of which Excel coerces — so the text is first
// required to be made of decimal number characters. Otherwise the checker would
// announce a divergence on `="Inf"+1`, which is #VALUE! in both languages.
func (t literalText) namesANumber() bool {
	trimmed := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(string(t)), "$"), "%"))
	if strings.TrimLeft(trimmed, "0123456789+-.eE") != "" {
		return false
	}
	_, err := strconv.ParseFloat(trimmed, 64)
	return err == nil
}

// stringLiteral reports a node's text when it is a quoted literal.
func stringLiteral(node tsvt.Expr) (literalText, bool) {
	lit, ok := node.(tsvt.StringLit)
	return literalText(lit.Value), ok
}

// isValueLiteral reports whether a node is a non-text literal — the operands
// against which a text comparison provably fails here and provably does not in
// Excel. A reference is deliberately excluded: its type is a runtime fact.
//
// A number is recognised in every spelling literalNumber accepts, signs and
// percents included. Matching only a bare tsvt.Number left `="a"<-1` and
// `="a"<100%` silent while `="a"<1` was announced — the same divergence, missed
// because of how the author happened to write the other operand.
func isValueLiteral(node tsvt.Expr) bool {
	if _, isBool := node.(tsvt.BoolLit); isBool {
		return true
	}
	_, isNumber := literalNumber(node)
	return isNumber
}

// isArithmetic reports whether the operator does arithmetic, where Excel
// coerces numeric text. `&` is excluded: concatenation takes text on purpose.
func isArithmetic(op tsvt.BinaryOp) bool {
	switch op {
	case tsvt.OpAdd, tsvt.OpSub, tsvt.OpMul, tsvt.OpDiv, tsvt.OpPow:
		return true
	default:
		return false
	}
}

// percentDivisor is what a postfix percent divides by: `100%` is 1.
const percentDivisor = 100
