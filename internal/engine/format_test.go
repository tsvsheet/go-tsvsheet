package engine_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// How a computed number is written. The rule is "at most 15 significant
// digits", and every test here exists because some plausible reading of that
// sentence is wrong: it does not mean rounding the value (that fabricates
// digits), it does not mean trimming an integer's zeros (that divides by
// powers of ten), and it does not apply to a literal the file already holds.

// TestFractionDigitsWriteComputedNumbersAtFifteenSignificantDigits pins the display
// rule through the path a reader actually meets: a computed number is written
// at the precision a float64 carries honestly, so what appears in the grid is
// the number meant rather than an artefact of binary representation.
func TestFractionDigitsWriteComputedNumbersAtFifteenSignificantDigits(t *testing.T) {
	t.Parallel()
	for formula, want := range map[string]string{
		"=0.1+0.2":   "0.3",
		"=1/3":       "0.333333333333333",
		"=2/3":       "0.666666666666667",
		"=0.1+0.7":   "0.8",
		"=-0.1-0.2":  "-0.3",
		"=1.5":       "1.5",
		"=0":         "0",
		"=1.1*1.1":   "1.21",
		"=100000*10": "1000000",
	} {
		t.Run(formula, func(t *testing.T) {
			sheet, err := engine.Parse([]byte(formula + "\n"))
			require.NoError(t, err)
			assert.Equal(t, want, sheet.Compute()[0][0])
		})
	}
}

// TestDisplayRoundingDoesNotChangeTheStoredValue pins that the rounding is
// presentation only: a later computation sees the full float64, so a chain of
// arithmetic never accumulates the error the display hides.
func TestDisplayRoundingDoesNotChangeTheStoredValue(t *testing.T) {
	t.Parallel()
	// If rounding touched the value, the second cell would read 0.3 exactly
	// and this comparison would be TRUE.
	sheet, err := engine.Parse([]byte("=0.1+0.2\n=A1=0.3\n"))
	require.NoError(t, err)
	computed := sheet.Compute()
	assert.Equal(t, "0.3", computed[0][0], "the reader sees the number meant")
	assert.Equal(t, "FALSE", computed[1][0], "and the engine still computes on the honest float")
}

// TestAStoredLiteralIsNeverReformatted pins SPECIFICATION §3 against the
// display rule: a literal is stored verbatim, so a sheet holding 4.50 or a
// long decimal keeps exactly those characters — display rounding applies to
// what a formula computes, never to what the file says.
func TestFormatNumberNeverReformatsAStoredLiteral(t *testing.T) {
	t.Parallel()
	const source = "4.50\t0.30000000000000004\t1e20\n"
	sheet, err := engine.Parse([]byte(source))
	require.NoError(t, err)
	assert.Equal(t, []string{"4.50", "0.30000000000000004", "1e20"}, sheet.Compute()[0])
}

// TestFormatNumberWritesLargeMagnitudesExactly is the property that makes the display
// rule safe to apply to a document engine. Computed text is written back into
// documents — Compute feeds Paste, render feeds a file — so a display that
// rounds by re-expanding a rounded value would not merely look different, it
// would silently replace the number with a nearby one. Every case here is a
// magnitude whose integer part alone exceeds the digit budget, where there is
// no fraction to trim and so nothing to lose: the text must parse back to the
// identical float64. (Trimming a fraction IS lossy, deliberately — that is what
// makes 0.1+0.2 read 0.3 — so this exactness is claimed only in this regime.) Rounding to 15 significant digits and
// re-expanding would instead have written 1152921504606850000 for 2^60 — a
// different number, four of whose digits were never in the value. What is
// written is the exact integer the float is.
func TestFormatNumberWritesLargeMagnitudesExactly(t *testing.T) {
	t.Parallel()
	for formula, want := range map[string]string{
		"=2^60":   "1152921504606846976",
		"=-2^60":  "-1152921504606846976",
		"=2^53+1": "9007199254740992",
		"=2^70":   "1180591620717411303424",
	} {
		t.Run(formula, func(t *testing.T) {
			sheet, err := engine.Parse([]byte(formula + "\n"))
			require.NoError(t, err)
			shown := sheet.Compute()[0][0]

			assert.Equal(t, want, shown)

			// The text must be the exact decimal expansion of the float it
			// denotes — not merely close enough to re-parse to it. Asserting
			// against the *shortest* re-parsing form would pass for
			// 1152921504606847000 too, which is a different number wearing
			// three zeros the value never had.
			back, err := strconv.ParseFloat(shown, 64)
			require.NoError(t, err)
			assert.Equal(t, shown, strconv.FormatFloat(back, 'f', 0, 64),
				"the text must name the float exactly")
		})
	}
}

// TestComputedTextPastedBackIsTheSameNumber closes the loop the previous test
// describes: the two public calls a caller naturally chains must not change
// the data in passing.
func TestComputedTextPastedBackIsTheSameNumber(t *testing.T) {
	t.Parallel()
	sheet, err := engine.Parse([]byte("=2^60\n"))
	require.NoError(t, err)

	pasted, err := sheet.Paste(engine.Address{Row: 1}, engine.Address{}, sheet.Compute(), engine.DefaultLimits())
	require.NoError(t, err)

	assert.Equal(t, "1152921504606846976", pasted.Source()[1][0])
	assert.Equal(t, pasted.Compute()[0][0], pasted.Compute()[1][0],
		"the pasted value and the value it came from must read alike")
}

// TestAnOverflowIsAnErrorValueSoTheDisplayRuleNeverMeetsAnInfinity pins the
// precondition formatNumber relies on. The earlier version of this test
// asserted that an overflow was "not written in exponent form", which passed
// vacuously: the cell holds #NUM!, so it was checking the spelling of an error
// code and would have passed no matter what the display rule did.
func TestAnOverflowIsAnErrorValueSoTheDisplayRuleNeverMeetsAnInfinity(t *testing.T) {
	t.Parallel()
	sheet, err := engine.Parse([]byte("=power(10,308)*10\t=0/0\t=-power(10,308)*10\n"))
	require.NoError(t, err)
	computed := sheet.Compute()[0]

	assert.Equal(t, "#NUM!", computed[0], "an overflow never reaches the display rule as an infinity")
	assert.Equal(t, "#DIV/0!", computed[1])
	assert.Equal(t, "#NUM!", computed[2])
}

// TestFormatNumberTrimsOnlyAFractionNeverAnIntegersZeros pins the guard that
// separates the two writing modes. Without it, a magnitude with no fraction has
// its trailing zeros trimmed as though they were fraction padding, and
// 1000000000000000 is written as "1" — a value a thousand trillion times wrong,
// with the whole suite otherwise green.
func TestFormatNumberTrimsOnlyAFractionNeverAnIntegersZeros(t *testing.T) {
	t.Parallel()
	for formula, want := range map[string]string{
		"=100000*10000000000": "1000000000000000",
		"=123456789012345*10": "1234567890123450",
		"=2^60":               "1152921504606846976",
		"=1000000*1000":       "1000000000",
	} {
		sheet, err := engine.Parse([]byte(formula + "\n"))
		require.NoError(t, err)

		assert.Equal(t, want, sheet.Compute()[0][0], formula)
	}
}

// TestAComputedZeroIsWrittenWithoutItsSign pins the normalization of negative
// zero. -0 is a float64 distinction with no meaning in a grid, and a cell
// reading "-0" reads as a bug to everyone who sees it.
func TestAComputedZeroIsWrittenWithoutItsSign(t *testing.T) {
	t.Parallel()
	sheet, err := engine.Parse([]byte("=round(-0.4)\t=-1*0\t=0*-1\t=-0\n"))
	require.NoError(t, err)

	assert.Equal(t, []string{"0", "0", "0", "0"}, sheet.Compute()[0])
}

// TestNumberValueMapsAnOverflowToAnErrorNeverAnInfinity pins the invariant at the
// constructor. It held for the binary operators and not for the aggregates, so
// `=sum(1e308,1e308)` wrote the literal text "+Inf" into a cell — a value the
// formula grammar cannot express, which then travels into documents through
// Compute and Paste.
func TestNumberValueMapsAnOverflowToAnErrorNeverAnInfinity(t *testing.T) {
	t.Parallel()
	sheet, err := engine.Parse([]byte("1e308\t1e308\t=sum(A1:B1)\t=avg(A1:B1)\t=product(A1:B1)\t=len(sum(A1:B1))\n"))
	require.NoError(t, err)
	computed := sheet.Compute()[0]

	assert.Equal(t, []string{"1e308", "1e308", "#NUM!", "#NUM!", "#NUM!", "#NUM!"}, computed)
}

// TestSignificantExponentPlacesTheLeadingDigitExactly pins the digit count at
// the boundary that a logarithm gets wrong. For a magnitude a hair below a
// power of ten, math.Log10 returns exactly the integer above — Log10 of
// 9.999999999999994e-09 is -8, not -8.0000000000000003 — so the count came out
// one too high and the value was written at fourteen significant digits.
func TestSignificantExponentPlacesTheLeadingDigitExactly(t *testing.T) {
	t.Parallel()
	for formula, want := range map[string]string{
		"=9999999999999994/1000000000000000000000000": "0.00000000999999999999999",
		"=1/3":       "0.333333333333333",
		"=0.1+0.2":   "0.3",
		"=1000000/8": "125000",
	} {
		sheet, err := engine.Parse([]byte(formula + "\n"))
		require.NoError(t, err)

		assert.Equal(t, want, sheet.Compute()[0][0], formula)
	}
}
