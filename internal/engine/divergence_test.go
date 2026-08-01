package engine_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// Every test here pairs a diagnostic with the behaviour its text claims: a
// message is a promise about what the engine does, and an assertion matching
// only the words would keep passing while the promise went false — which is how
// the first draft told an author `"a"="b"` was #VALUE! when it computes FALSE.

// checkOne parses a one-cell sheet and returns its computed value and messages.
func checkOne(t *testing.T, formula string) (string, []string) {
	t.Helper()
	sheet, err := engine.Parse([]byte(formula))
	require.NoError(t, err)
	diags := engine.Check(sheet)
	messages := make([]string, len(diags))
	for i, diag := range diags {
		assert.Equal(t, "A1", diag.Cell, "a divergence names the cell that contains it")
		assert.False(t, diag.IsFatal, "a divergence is advisory, never fatal")
		messages[i] = diag.Message
	}
	return sheet.Compute()[0][0], messages
}

// requireOneNoting asserts exactly one diagnostic, mentioning each fragment.
func requireOneNoting(t *testing.T, messages []string, fragments ...string) {
	t.Helper()
	require.Len(t, messages, 1)
	for _, fragment := range fragments {
		assert.Contains(t, messages[0], fragment)
	}
}

func TestASignAppliedToAPowerIsAnnouncedAndBindsLooserThanInExcel(t *testing.T) {
	value, messages := checkOne(t, `=-2^2`)

	// Excel yields 4 for this text; the diagnostic exists because we yield -4.
	assert.Equal(t, "-4", value)
	requireOneNoting(t, messages, "unary sign binds looser", "Excel reads (-x)^y")
}

func TestAChainedPowerIsAnnouncedAndAssociatesRightUnlikeExcel(t *testing.T) {
	value, messages := checkOne(t, `=2^3^2`)

	// 2^(3^2) = 512. Excel reads (2^3)^2 = 64.
	assert.Equal(t, "512", value)
	requireOneNoting(t, messages, "associates to the right", "Excel reads (x^y)^z")
}

func TestAUnaryPlusOverAPowerIsNotAnnouncedBecauseBothReadingsAgree(t *testing.T) {
	// +x is x, so +(2^2) and (+2)^2 are 4 either way. There is nothing to say,
	// and the sign message would name a minus that is not in the formula.
	value, messages := checkOne(t, `=+2^2`)

	assert.Equal(t, "4", value)
	assert.Empty(t, messages)
}

func TestNegationDistributesOverAnOddExponentSoNothingIsAnnounced(t *testing.T) {
	// (-x)^n == -(x^n) for every odd n, so the grouping cannot change a value.
	for formula, want := range map[string]string{`=-2^3`: "-8", `=-2^1`: "-2"} {
		value, messages := checkOne(t, formula)

		assert.Equal(t, want, value, formula)
		assert.Empty(t, messages, formula)
	}
}

func TestASignOverANonLiteralPowerIsAnnouncedBecauseItIsNotProvable(t *testing.T) {
	// The exponent is not known here, so the readings may differ and the author
	// is told. This is the canonical Excel gotcha and must not be lost.
	sheet, err := engine.Parse([]byte("2\t3\t=-A1^B1"))
	require.NoError(t, err)

	assert.NotEmpty(t, engine.Check(sheet))
}

func TestTextExcelWouldNotCoerceIsNotAnnounced(t *testing.T) {
	// Excel yields #VALUE! for these too, so the languages agree.
	// "Inf" and "NaN" are here because Go's ParseFloat accepts them as
	// spellings of a float and Excel does not; treating them as numeric would
	// put a finding on an expression the two languages answer identically.
	for _, formula := range []string{`="abc"+1`, `=""+1`, `="1a"*2`, `="Inf"+1`, `="NaN"+1`, `="0x10"+1`} {
		value, messages := checkOne(t, formula)

		assert.Equal(t, "#VALUE!", value, formula)
		assert.Empty(t, messages, "%s is #VALUE! in Excel as well", formula)
	}
}

func TestTextExcelWouldCoerceIsStillAnnounced(t *testing.T) {
	for _, formula := range []string{`="3"+1`, `=" 3 "+1`, `="3%"+1`, `="$3"+1`, `="-3"+1`} {
		value, messages := checkOne(t, formula)

		assert.Equal(t, "#VALUE!", value, formula)
		requireOneNoting(t, messages, "text in arithmetic is #VALUE!")
	}
}

func TestTextInArithmeticIsAnnouncedAndRefusedUnlikeExcel(t *testing.T) {
	// Excel coerces "3" and yields 4; every one of these is #VALUE! here.
	for _, formula := range []string{`="3"+1`, `="3"-1`, `="3"*2`, `="3"/1`, `="3"^2`, `=1+"3"`, `=-"3"`, `="3"%`} {
		value, messages := checkOne(t, formula)

		assert.Equal(t, "#VALUE!", value, formula)
		requireOneNoting(t, messages, "text in arithmetic is #VALUE!", "Excel would coerce")
	}
}

func TestConcatenationOfTextIsNotADivergence(t *testing.T) {
	// & takes text on purpose, and Excel agrees, so there is nothing to say.
	value, messages := checkOne(t, `="a"&"b"`)

	assert.Equal(t, "ab", value)
	assert.Empty(t, messages)
}

func TestADivergenceIsAdvisorySoTheSheetStillComputes(t *testing.T) {
	// check reports; it does not refuse. A divergent sheet is a valid sheet.
	sheet, err := engine.Parse([]byte("=-2^2\t=A1+1"))
	require.NoError(t, err)

	require.NotEmpty(t, engine.Check(sheet))
	assert.Equal(t, []string{"-4", "-3"}, sheet.Compute()[0])
}

func TestADivergenceNestedInsideAFormulaIsStillFound(t *testing.T) {
	// The walk is over the whole tree, not just its root.
	value, messages := checkOne(t, `=sum(1, -2^2)`)

	assert.Equal(t, "-3", value)
	requireOneNoting(t, messages, "unary sign binds looser")
}

func TestEveryDivergenceInOneFormulaIsReported(t *testing.T) {
	_, messages := checkOne(t, `=-2^2 + "3"`)

	assert.Len(t, messages, 2, "one finding per divergent construct: %v", messages)
}

func TestACleanFormulaAnnouncesNothing(t *testing.T) {
	value, messages := checkOne(t, `=(1 + 2) * 3`)

	assert.Equal(t, "9", value)
	assert.Empty(t, messages)
}

func TestDivergenceNotesRepeatExactlyWhatCheckReports(t *testing.T) {
	// The two must not drift: an author who reads a sheet without running the
	// checker still learns why the cell reads the way it does.
	sheet, err := engine.Parse([]byte(`=-2^2`))
	require.NoError(t, err)

	trace, err := engine.Explain(sheet, engine.Address{})
	require.NoError(t, err)

	require.Len(t, trace.Notes, 1)
	assert.Equal(t, engine.Check(sheet)[0].Message, trace.Notes[0])
	assert.Contains(t, trace.Notes[0], "Excel reads (-x)^y")
}

func TestExplainOfACleanCellCarriesNoNotes(t *testing.T) {
	sheet, err := engine.Parse([]byte("=1+1\t2"))
	require.NoError(t, err)

	formula, err := engine.Explain(sheet, engine.Address{})
	require.NoError(t, err)
	literal, err := engine.Explain(sheet, engine.Address{Col: 1})
	require.NoError(t, err)

	assert.Empty(t, formula.Notes)
	assert.Empty(t, literal.Notes)
}

func TestEveryDivergenceMessageNamesExcelAndTheRemedy(t *testing.T) {
	// A finding that does not say what Excel does is not an explanation, and
	// one that does not say what to write instead is not actionable.
	for _, formula := range []string{`=-2^2`, `=2^3^2`, `="3"+1`, `="1"=1`, `="a"="A"`} {
		_, messages := checkOne(t, formula)

		require.Len(t, messages, 1, formula)
		assert.Contains(t, messages[0], "Excel", formula)
		hasRemedy := strings.Contains(messages[0], "write") ||
			strings.Contains(messages[0], "parenthesize") ||
			strings.Contains(messages[0], "wrap") ||
			strings.Contains(messages[0], "compare against")
		assert.True(t, hasRemedy, "%s: no remedy in %q", formula, messages[0])
	}
}

// TestTheAnnouncedSetIsExactlyTheDivergentSet is the whole contract in one
// table: every construct that reads differently in Excel is announced, and
// nothing else is. A missed divergence teaches nothing and a false alarm
// teaches the reader to skip the output, so both matter. Every `false` row here
// was a real false positive at some point.
func TestTheAnnouncedSetIsExactlyTheDivergentSet(t *testing.T) {
	announced := map[string]bool{
		`=-2^2`: true, `=2^3^2`: true, `="3"+1`: true, `="1"=1`: true,
		`="a"="A"`: true, `=TRUE=1`: true, `=-A1^2`: true, `="3%"*2`: true,

		`=+2^2`:     false, // +x is x: both read 4
		`=-2^3`:     false, // (-x)^n == -(x^n) for odd n
		`=-2^-3`:    false, // a signed exponent is still a literal odd one
		`=2^2^1`:    false, // 2^(2^1) == (2^2)^1
		`="abc"+1`:  false, // #VALUE! in Excel too
		`="Inf"+1`:  false, // Go spells a float this way; Excel does not
		`="1_0"+1`:  false,
		`=TRUE<0`:   false, // FALSE under 1/0 folding AND under Excel's ranking
		`=0>FALSE`:  false,
		`="a"<"b"`:  false, // same answer with or without case
		`="a"&"b"`:  false, // concatenation takes text on purpose
		`=-(2^2)`:   false, // the author already said which reading
		`=A1="yes"`: false, // a runtime fact, not a provable divergence
		`=(1+2)*3`:  false,
	}
	for formula, wantAnnounced := range announced {
		sheet, err := engine.Parse([]byte(formula))
		require.NoError(t, err, formula)

		assert.Equal(t, wantAnnounced, len(engine.Check(sheet)) > 0,
			"%s computes %s", formula, sheet.Compute()[0][0])
	}
}

func TestAChainedPowerOverReferencesIsAnnouncedBecauseItIsNotProvable(t *testing.T) {
	// With literal operands the two groupings can be computed and compared, and
	// `=2^2^1` is 4 either way. Reference operands cannot be, so the author is
	// told — the readings really may differ once the cells hold values.
	sheet, err := engine.Parse([]byte("2\t3\t4\t=A1^B1^C1"))
	require.NoError(t, err)

	diags := engine.Check(sheet)

	require.Len(t, diags, 1)
	assert.Contains(t, diags[0].Message, "associates to the right")
}
