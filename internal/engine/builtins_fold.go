// The text folds: CONCAT and TEXTJOIN, whose operands flatten from ranges, so
// a single argument may be a whole column. Split from builtins_text.go for
// size; they share the byte budget that keeps a fold from building what it
// would then refuse.
package engine

import "strings"

// separator is TEXTJOIN's delimiter — the text placed between operands, and
// only between them (TestTextJoin_ExcelSemantics pins both edges: a lone
// operand renders bare, and the fold ends without a trailing delimiter).
type separator string

// skipEmpties is TEXTJOIN's ignore_empty flag: when set, an empty operand
// contributes neither text nor a separator.
type skipEmpties bool

// joinText folds operands into one string separated by sep, refusing with
// #LIMIT! the moment the result would pass the byte budget — the check is made
// BEFORE each append, so an over-budget fold does not allocate what it
// refuses, which is the whole point of bounding a fold whose operands may be a
// whole column. TestTextJoin_RefusalsAndArity and
// TestConcat_FoldsRangesUnderTheBudget state that by folding a column of cells
// each of which fits the budget while their result does not.
//
// Excel semantics: an empty operand contributes an empty string and still earns
// its separator unless ignore_empty is set; a boolean renders as TRUE/FALSE and
// a number in the sheet's canonical form, both through Value.String, so the text
// a formula joins is the text the grid shows.
func joinText(args []Value, sep separator, isSkippingEmpties skipEmpties, limit byteBudget) Value {
	var b strings.Builder
	for _, arg := range args {
		piece, isSkipped := foldPiece(arg, sep, isSkippingEmpties, writtenBytes(b.Len()))
		if isSkipped {
			continue
		}
		if int64(b.Len())+int64(len(piece)) > int64(limit) {
			return errorValue(ErrLimit) // result exceeds the byte budget
		}
		_, _ = b.WriteString(piece)
	}
	return stringValue(textVal(b.String()))
}

// writtenBytes is how much of the fold is already built — zero marks the first
// operand, which takes no leading separator.
type writtenBytes int

// foldPiece renders one operand together with the separator that precedes it,
// reporting a skip for an empty operand the caller asked to drop.
func foldPiece(arg Value, sep separator, isSkippingEmpties skipEmpties, written writtenBytes) (string, boolResult) {
	switch arg.kind {
	case kindEmpty:
		if bool(isSkippingEmpties) {
			return "", true
		}
	default:
	}
	if written > 0 {
		return string(sep) + arg.String(), false
	}
	return arg.String(), false
}
