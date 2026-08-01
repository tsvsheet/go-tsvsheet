// Writing a computed number.
//
// Every number a formula produces passes through here on its way into a cell,
// and the cell is a document: what is written gets committed, diffed, pasted
// back in, and read by the next tool. So the rule is not "make it look nice" —
// it is "write text that names the value it came from".
//
// Binary floating point makes those goals pull apart. 0.1+0.2 is genuinely
// 0.30000000000000004, and printing all seventeen digits reports an artefact of
// the representation as though it were information. Trimming to fifteen
// significant digits fixes that, and is what every spreadsheet does. But the
// trimming has to be applied as a decimal PLACE count: rounding the value and
// re-expanding it writes 1152921504606850000 for 2^60, which is a different
// number wearing four digits the value never had. Losing a fraction is a
// bounded, explicable loss; inventing an integer digit is not.
package engine

import (
	"strconv"
	"strings"
)

// displayDigits is how many significant digits a computed number is written
// with: the precision a float64 carries honestly. Beyond it the digits are an
// artefact of binary representation rather than information — which is why
// 0.1+0.2 prints as 0.30000000000000004 without this, and 0.3 with it.
const displayDigits = 15

// formatNumber writes a computed number for a reader.
//
// Binary floating point cannot represent most decimal fractions, so a sum of
// two exact-looking numbers carries a tail of digits that is an artefact of the
// representation rather than information: 0.1+0.2 is 0.30000000000000004. The
// tail is dropped by writing at most displayDigits significant digits.
//
// The rounding is applied as a *decimal place count*, never by re-expanding a
// rounded value. That distinction is the whole correctness of this function:
// rounding 2^60 to 15 significant digits and re-expanding it would write
// 1152921504606850000, four of whose digits were invented. Computed text is
// written back into documents here, so an invented digit is not a display
// artefact — it is data loss. Trimming a fraction is a deliberate, bounded loss;
// fabricating a digit is not a loss the reader can reason about at all.
//
// A magnitude with no room left for a fraction is therefore written as the
// exact integer it is: 2^60 is 1152921504606846976, every digit its own.
//
// The value is always finite: an evaluation that overflows produces #NUM! (an
// error value, not an infinity), so no formula can deliver an infinity here.
func formatNumber(num floatVal) string {
	value := float64(num)
	if value == 0 {
		// Also normalizes negative zero: -0 is a float64 distinction with no
		// meaning in a grid, and `=round(-0.4)` showing "-0" reads as a bug.
		return zeroText
	}
	fraction := fractionDigits(num)
	written := strconv.FormatFloat(value, 'f', fraction, 64)
	if fraction == 0 {
		return written
	}
	return strings.TrimSuffix(strings.TrimRight(written, "0"), ".")
}

// zeroText is how a zero is written, whatever its sign bit.
const zeroText = "0"

// fractionDigits is how many decimal places leave a number at displayDigits
// significant digits: zero places once the integer part alone fills the budget,
// so the artefact tail is trimmed at every magnitude rather than only below
// 1e15.
//
// Zero is handled here rather than assumed away, because log10 of zero is -Inf
// and the integer conversion of that is not defined — leaving it to a caller
// would make the count depend on discipline instead of on arithmetic.
func fractionDigits(num floatVal) int {
	integerDigits := significantExponent(num) + 1
	if integerDigits >= displayDigits {
		return 0
	}
	return displayDigits - integerDigits
}

// significantExponent is the power of ten the number's leading digit sits on,
// taken from a displayDigits-significant rendering rather than from a logarithm.
//
// math.Log10 is not reliable for this. For a magnitude a hair below a power of
// ten it returns exactly the integer above — Log10(9.999999999999994e-09) is
// -8, not -8.0000000000000003 — so the digit count came out one too high and
// the value was written at fourteen significant digits. Reading the exponent
// back off the formatter that will do the writing cannot disagree with it, and
// it accounts for a value that rounds up into the next power of ten.
func significantExponent(num floatVal) int {
	// The text is one this function just produced, so it always carries an 'e'
	// followed by a signed integer; the conversion cannot fail on it, and a
	// branch for that would be a branch no input can reach.
	text := strconv.FormatFloat(float64(num), 'e', displayDigits-1, 64)
	exponent, _ := strconv.Atoi(text[strings.IndexByte(text, 'e')+1:])
	return exponent
}
