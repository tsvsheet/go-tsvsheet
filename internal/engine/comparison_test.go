package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// Comparison diverges for three unrelated reasons — type strictness, case
// sensitivity, and how a boolean ranks — and each is announced only where the
// answer actually differs. The silent cases below matter as much as the loud
// ones: they are the shapes a structural check would have flagged wrongly.

func TestComparingTextWithANonTextLiteralIsAnnouncedAndRefused(t *testing.T) {
	// Excel never errors on a cross-type comparison: it orders every number
	// below every text and every text below every boolean, so it answers
	// FALSE to the equalities here and TRUE to `="5">3` (text outranks a
	// number). We refuse all of them instead.
	for _, formula := range []string{`="1"=1`, `=1="1"`, `="5">3`, `="a"=TRUE`, `="a"<>2`} {
		value, messages := checkOne(t, formula)

		assert.Equal(t, "#VALUE!", value, formula)
		requireOneNoting(t, messages, "comparing text with a non-text value", "Excel would compare them")
	}
}

func TestComparingTextWithAReferenceIsNotAnnouncedBecauseItIsNotProvable(t *testing.T) {
	// Whether B1 holds text is a runtime fact. Firing here would put a
	// diagnostic on the most common formula in any spreadsheet, and the
	// message would be wrong whenever the cell does hold text.
	sheet, err := engine.Parse([]byte("=\"yes\"=B1\tyes"))
	require.NoError(t, err)

	assert.Equal(t, "TRUE", sheet.Compute()[0][0], "text against text compares fine")
	assert.Empty(t, engine.Check(sheet))
}

func TestATextComparisonThatTurnsOnCaseIsAnnounced(t *testing.T) {
	// Case-sensitive here, case-blind in Excel: Excel answers TRUE to both.
	for _, formula := range []string{`="a"="A"`, `="B">"a"`} {
		value, messages := checkOne(t, formula)

		assert.Equal(t, "FALSE", value, formula)
		requireOneNoting(t, messages, "case-sensitive here", "Excel ignores case")
	}
}

func TestTextComparisonTurnsOnCaseGatesWhetherACaseFindingIsAnnounced(t *testing.T) {
	// Reporting a comparison whose answer is the same either way would be
	// noise, and noise is what makes a checker stop being read.
	for formula, want := range map[string]string{
		`="a"<"b"`:  "TRUE",
		`="a"="a"`:  "TRUE",
		`="a"<>"b"`: "TRUE",
		`="b">="a"`: "TRUE",
		`="a"<="b"`: "TRUE",
		`="b"<"a"`:  "FALSE",
	} {
		value, messages := checkOne(t, formula)

		assert.Equal(t, want, value, formula)
		assert.Empty(t, messages, "%s reads the same in both", formula)
	}
}

func TestABooleanComparedWithANumberIsAnnounced(t *testing.T) {
	// Booleans fold to 1/0 here; Excel orders every boolean above every number,
	// so it answers the opposite.
	for formula, want := range map[string]string{`=TRUE=1`: "TRUE", `=1<TRUE`: "FALSE"} {
		value, messages := checkOne(t, formula)

		assert.Equal(t, want, value, formula)
		requireOneNoting(t, messages, "compares as 1 or 0 here", "Excel orders every boolean")
	}
}

func TestABooleanComparedWithAPercentLiteralIsStillJudged(t *testing.T) {
	// A percent is a numeric literal wearing a suffix, so the comparison is
	// just as decidable as `=TRUE=1`. Reading only bare numbers left sixteen
	// provably divergent comparisons silent.
	sheet, err := engine.Parse([]byte(`=TRUE=100%`))
	require.NoError(t, err)

	assert.Equal(t, "TRUE", sheet.Compute()[0][0], "100% folds to 1, and TRUE folds to 1")
	assert.Len(t, engine.Check(sheet), 1, "Excel ranks TRUE above 1 and answers FALSE")
}

func TestABooleanComparedWithAPercentThatAgreesIsSilent(t *testing.T) {
	// The discrimination has to survive the new literal form: 50% folds to 0.5,
	// so this is FALSE under 1/0 folding and FALSE under Excel's ranking.
	sheet, err := engine.Parse([]byte(`=TRUE=50%`))
	require.NoError(t, err)

	assert.Equal(t, "FALSE", sheet.Compute()[0][0])
	assert.Empty(t, engine.Check(sheet))
}

func TestABooleanBelowTheNumberItIsComparedWithIsStillJudged(t *testing.T) {
	// TRUE folds to 1, which is less than 2, so this language answers TRUE.
	// Excel ranks every boolean above every number and answers FALSE. The
	// ordering has to be computed in both directions for that to be seen.
	sheet, err := engine.Parse([]byte(`=TRUE<2`))
	require.NoError(t, err)

	assert.Equal(t, "TRUE", sheet.Compute()[0][0])
	assert.Len(t, engine.Check(sheet), 1, "Excel ranks TRUE above 2 and answers FALSE")
}

func TestEveryComparisonOperatorIsJudgedOnItsOwnBoundary(t *testing.T) {
	// The ordering is applied twice and the answers compared, so an operator
	// whose boundary case is wrong goes silent rather than loud. `<=` and `>=`
	// are the ones that hide it: they differ from `<` and `>` only when the two
	// operands rank alike, which is exactly where a case-blind comparison and a
	// case-sensitive one part company.
	for _, formula := range []string{`="a"<="A"`, `="A">="a"`, `="a"="A"`, `="a"<>"A"`} {
		sheet, err := engine.Parse([]byte(formula))
		require.NoError(t, err)

		assert.Len(t, engine.Check(sheet), 1, "%s turns on case and Excel ignores case", formula)
	}

	// The strict forms answer FALSE either way — "a" outranks "A" here and
	// ranks equal to it in Excel, and neither is strictly less — so they are
	// correctly silent. Asserting that keeps the boundary honest in both
	// directions rather than only where a finding is expected.
	for _, formula := range []string{`="a"<"A"`, `="A">"a"`} {
		sheet, err := engine.Parse([]byte(formula))
		require.NoError(t, err)

		assert.Equal(t, "FALSE", sheet.Compute()[0][0], formula)
		assert.Empty(t, engine.Check(sheet), "%s answers FALSE under both rules", formula)
	}
}

func TestAComparisonWithANumberInAnySpellingIsJudged(t *testing.T) {
	// The non-text side is recognised however the author wrote it. Matching
	// only a bare number announced `="a"<1` and stayed silent on `="a"<-1`,
	// which is the same divergence with a sign typed in front of it.
	for _, formula := range []string{`="a"<1`, `="a"<-1`, `="a"=-1`, `="a"<100%`, `="a"<+1`, `="5">-0.5`, `=-1="a"`} {
		sheet, err := engine.Parse([]byte(formula))
		require.NoError(t, err)

		assert.Equal(t, "#VALUE!", sheet.Compute()[0][0], formula)
		assert.Len(t, engine.Check(sheet), 1, "%s is a plain comparison in Excel", formula)
	}
}
