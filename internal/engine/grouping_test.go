package engine_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// Authored parentheses have to survive every path that re-renders a formula.
// The property is not cosmetic: `(-2)^2` is 4 and `-2^2` is -4, so dropping a
// pair changes a number, and dropping a redundant pair re-arms a divergence
// diagnostic the author had already answered.

func TestGroupingSilencesEveryPrecedenceDivergence(t *testing.T) {
	// The remedy the diagnostics recommend has to actually silence them,
	// otherwise the advice cannot be taken.
	for formula, want := range map[string]string{
		`=-(2^2)`:  "-4",
		`=(-2)^2`:  "4",
		`=2^(3^2)`: "512",
		`=(2^3)^2`: "64",
	} {
		value, messages := checkOne(t, formula)

		assert.Equal(t, want, value, formula)
		assert.Empty(t, messages, "%s is explicit and has nothing to announce", formula)
	}
}

func TestGroupingSurvivesRewritingSoAFixIsNotUndone(t *testing.T) {
	// A rewrite re-renders every formula from its parse tree. If grouping were
	// erased there, an author's `-(A1^2)` would silently become `-A1^2` and the
	// diagnostic they had just silenced would come back — the tool undoing the
	// fix it asked for.
	sheet, err := engine.Parse([]byte("2\t=-(A1^2)"))
	require.NoError(t, err)
	require.Empty(t, engine.Check(sheet), "the author already took the advice")

	edited := sheet.InsertRow(engine.Address{})

	assert.Equal(t, "=-(A2 ^ 2)", edited.Source()[1][1])
	assert.Empty(t, engine.Check(edited), "the fix must survive the edit")
}

func TestGroupingChangesTheValueItRendersBackTo(t *testing.T) {
	// `(-2)^2` is 4 and `-2^2` is -4. If a rewrite dropped these parentheses
	// the sheet would silently start computing a different number.
	sheet, err := engine.Parse([]byte("2\t=(-A1)^2"))
	require.NoError(t, err)

	edited := sheet.InsertRow(engine.Address{})

	assert.Equal(t, "=(-A2) ^ 2", edited.Source()[1][1])
	assert.Equal(t, "4", edited.Compute()[1][1])
}

func TestGroupingIsNeverRenderedTwice(t *testing.T) {
	// A parenthesized expression binds as an atom, so the operand helpers must
	// not add a second pair around the pair the author wrote.
	for source, want := range map[string]string{
		"=(A1+2)*3": "=(A2 + 2) * 3",
		"=(A1+2)%":  "=(A2 + 2)%",
		"=-(A1+2)":  "=-(A2 + 2)",
		// A sign binds tighter than +, so these parentheses are not required
		// by precedence — they survive only because the author wrote them.
		"=(-A1)+1":      "=(-A2) + 1",
		"=(A1|len())+1": "=(A2 | len) + 1",
	} {
		sheet, err := engine.Parse([]byte("7\t" + source))
		require.NoError(t, err)

		edited := sheet.InsertRow(engine.Address{})

		assert.Equal(t, want, edited.Source()[1][1], source)

		// Rendering must be idempotent: re-parsing the rewritten text and
		// rendering it again yields the same formula, so repeated edits cannot
		// accrete or shed parentheses.
		rewritten := edited.Source()[1][1]
		reparsed, err := engine.Parse([]byte(rewritten))
		require.NoError(t, err, "a rewritten formula must still parse: %s", source)
		trace, err := engine.Explain(reparsed, engine.Address{})
		require.NoError(t, err)
		assert.Equal(t, strings.TrimPrefix(rewritten, "="), trace.Formula, source)
	}
}
