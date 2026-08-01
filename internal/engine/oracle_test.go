package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// These are the cases that showed structural detection is not good enough. Each
// one has a shape that looks divergent and a value that is not — or the reverse
// — and each was silently wrong before the two readings were actually evaluated.
// They are written through Check, because the property that matters is what an
// author is told, not what an internal helper returns.

// announcedFor reports whether Check says anything about a one-cell formula.
func announcedFor(t *testing.T, formula string) (string, bool) {
	t.Helper()
	sheet, err := engine.Parse([]byte(formula))
	require.NoError(t, err)
	return sheet.Compute()[0][0], len(engine.Check(sheet)) > 0
}

func TestASignOverAPowerIsSilentWhereverTheSignCannotMatter(t *testing.T) {
	// `-(x^y)` and `(-x)^y` are the same for an odd exponent and for a zero
	// base, and both overflow together for a large one. Every case here was
	// announced when the rule was "an odd integer exponent" rather than "the
	// two readings agree".
	for formula, want := range map[string]string{
		`=-2^3`:     "-8",     // odd exponent
		`=-2^-3`:    "-0.125", // signed odd exponent
		`=-2^+3`:    "-8",     // and a plus-signed one, which is the same literal
		`=-0^2`:     "0",      // a zero base: the sign cannot survive either way
		`=-0^200`:   "0",
		`=-200^200`: "#NUM!", // both readings overflow, so both cells show #NUM!
	} {
		value, announced := announcedFor(t, formula)

		assert.Equal(t, want, value, formula)
		assert.False(t, announced, "%s reads %s under both groupings", formula, want)
	}
}

func TestASignOverAPowerIsAnnouncedWhereverItCanMatter(t *testing.T) {
	for formula, want := range map[string]string{
		`=-2^2`:   "-4",               // Excel: (-2)^2 = 4
		`=-2^0.5`: "-1.4142135623731", // Excel: (-2)^0.5 = #NUM!
		`=-3^4`:   "-81",
	} {
		value, announced := announcedFor(t, formula)

		assert.Equal(t, want, value, formula)
		assert.True(t, announced, "%s reads differently in Excel", formula)
	}
}

func TestASignInsideAPowerIsAnnouncedEvenWhenItAgreesInIsolation(t *testing.T) {
	// `-1^3` is -1 under either grouping, so in isolation there is nothing to
	// say. Inside a power there is: Excel binds the sign tighter AND groups the
	// outer `^` leftward, so `2^-1^3` is 0.5 here and 0.125 there. Judging the
	// inner node alone reported nothing at all.
	value, announced := announcedFor(t, `=2^-1^3`)

	assert.Equal(t, "0.5", value)
	assert.True(t, announced, "Excel reads (2^-1)^3 = 0.125")
}

func TestAChainedPowerIsSilentWhereBothGroupingsShowTheSameThing(t *testing.T) {
	for formula, want := range map[string]string{
		`=2^2^1`:        "4", // 2^(2^1) and (2^2)^1 are both 4
		`=2^3^1`:        "8",
		`=1^2^3`:        "1",
		`=(-2)^1.5^200`: "#NUM!", // this way an overflow, that way a NaN: #NUM! either way
	} {
		value, announced := announcedFor(t, formula)

		assert.Equal(t, want, value, formula)
		assert.False(t, announced, "%s shows %s under both groupings", formula, want)
	}
}

func TestAChainedPowerIsAnnouncedWhenAnIntermediateOverflows(t *testing.T) {
	// The whole reason readings are evaluated inside-out. Comparing only the
	// final numbers says these agree — 0^(200^200) is 0 and (0^200)^200 is 0 —
	// but 200^200 overflows before the outer power is reached, so the cell holds
	// #NUM! here while Excel, grouping leftward, shows 0.
	for _, formula := range []string{`=0^200^200`, `=1^200^200`, `=0.5^200^200`} {
		value, announced := announcedFor(t, formula)

		assert.Equal(t, "#NUM!", value, formula)
		assert.True(t, announced, "%s is #NUM! here and a number in Excel", formula)
	}
}

func TestAnOverflowInBothReadingsIsNotADivergence(t *testing.T) {
	// Two readings that both fail agree: the cell shows #NUM! either way, and
	// the +Inf and -Inf behind them are not values a reader ever sees.
	value, announced := announcedFor(t, `=-200^200`)

	assert.Equal(t, "#NUM!", value)
	assert.False(t, announced)
}

func TestASignedOperandIsReadWithItsSign(t *testing.T) {
	// `2^(-2)^2` is 2^((-2)^2) = 16 here and (2^-2)^2 = 0.0625 in Excel. Read
	// without its sign the exponent chain would appear to agree (2^(2^2) and
	// (2^2)^2 are both 16), and the divergence would go unannounced.
	value, announced := announcedFor(t, `=2^(-2)^2`)

	assert.Equal(t, "16", value)
	assert.True(t, announced, "Excel reads (2^-2)^2 = 0.0625")
}

func TestPrecedenceIsJudgedForTheWholeExpressionNotOneOperatorAtATime(t *testing.T) {
	// Excel's two precedence differences interact, so a per-node verdict answers
	// a question nobody asked. Each formula here contains a construct that looks
	// divergent alone and is not, once the expression is read whole. They were
	// 11% of all literal-only power expressions before the whole-expression
	// proof replaced the per-node guesses.
	for formula, want := range map[string]string{
		`=2^-2^2`:  "0.0625", // sign inside a chain, but m^e == m*e
		`=2^-1^1`:  "0.5",
		`=2^-3^1`:  "0.125",
		`=3^-2^1`:  "0.111111111111111",
		`=-2^3^1`:  "-8", // sign over a chain, odd exponent
		`=-10^3^1`: "-1000",
		`=--2^2`:   "4", // doubled sign cancels under both readings
		`=--3^4`:   "81",
	} {
		value, announced := announcedFor(t, formula)

		assert.Equal(t, want, value, formula)
		assert.False(t, announced, "%s is %s under Excel's precedence too", formula, want)
	}
}

func TestTheSameShapeIsAnnouncedWhenTheWholeExpressionReallyDiffers(t *testing.T) {
	// The counterpart: `2^-2^2` and `2^-1^3` are the same shape with one digit
	// changed, and only the second diverges. Anything that silenced the first by
	// pattern would silence this too.
	for formula, want := range map[string]string{
		`=2^-1^3`:    "0.5",   // Excel: (2^-1)^3 = 0.125
		`=0^200^200`: "#NUM!", // Excel: (0^200)^200 = 0
		`=2^(-2)^2`:  "16",    // Excel: (2^-2)^2 = 0.0625
	} {
		value, announced := announcedFor(t, formula)

		assert.Equal(t, want, value, formula)
		assert.True(t, announced, "%s reads differently in Excel", formula)
	}
}

func TestArithmeticAroundAPowerIsCarriedIntoTheProof(t *testing.T) {
	// The proof is about the whole expression, so everything around the power
	// has to be evaluated too — otherwise a divergent power inside an ordinary
	// sum could not be judged either way.
	for formula, want := range map[string]string{
		`=-2^3*2`:   "-16", // agrees: -(2^3)*2 and (-2)^3*2 are both -16
		`=-2^3/2`:   "-4",
		`=-2^3-1`:   "-9",
		`=-2^3+1`:   "-7",
		`=-2^3*50%`: "-4",
		`=-2^3/0`:   "#DIV/0!", // a division a cell reports, on both sides
	} {
		value, announced := announcedFor(t, formula)

		assert.Equal(t, want, value, formula)
		assert.False(t, announced, "%s reads %s under Excel's precedence too", formula, want)
	}
}

func TestAnUnmodelledOperatorLeavesTheDivergenceAnnounced(t *testing.T) {
	// The proof models arithmetic, not every operator. Concatenation is not
	// modelled, so the expression is undecidable and the finding stands —
	// "cannot prove they agree" must never quietly become "they agree".
	value, announced := announcedFor(t, `=-2^2&1`)

	assert.Equal(t, "-41", value)
	assert.True(t, announced)
}

func TestADivergentPowerInsideArithmeticIsStillAnnounced(t *testing.T) {
	// The counterpart: surrounding arithmetic must not launder a real
	// divergence into agreement.
	for formula, want := range map[string]string{
		`=-2^2*2`: "-8",                // Excel: (-2)^2*2 = 8
		`=-2^2+1`: "-3",                // Excel: 5
		`=-2^2/2`: "-2",                // Excel: 2
		`=-2^3%`:  "-1.02101212570719", // % binds to the exponent; Excel's (-2)^0.03 is #NUM!
	} {
		value, announced := announcedFor(t, formula)

		assert.Equal(t, want, value, formula)
		assert.True(t, announced, "%s reads differently in Excel", formula)
	}
}

func TestTheProofRespectsParenthesesInsideTheExpression(t *testing.T) {
	// The Excel tree is built by undoing two precedence differences, and both
	// stop at a parenthesis, because a parenthesized expression is an atom to
	// either language. If either rewrite ignored that, the proof would compare
	// against a tree Excel would never build and reach the wrong verdict — in
	// one direction announcing a divergence that is not there, in the other
	// hiding one that is. Both formulas were found by searching 4,778 generated
	// expressions for a case where the guards change the answer.
	inner, announced := announcedFor(t, `=-(-2^(2^2))*(2^2)`)

	assert.Equal(t, "64", inner)
	assert.True(t, announced, "the outer sign still meets an unparenthesized power")

	chained, announced := announcedFor(t, `=(2^2)^((2^2)^(2^3^2))`)

	assert.Equal(t, "#NUM!", chained)
	assert.False(t, announced, "every chain here is parenthesized, so Excel groups it the same way")
}
