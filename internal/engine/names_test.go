package engine_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// TestNames_TheCanonicalSheet is SPECIFICATION §5.6's own example: declare
// where the value lives, refer by sigil, compute as if the names were pinned
// references.
func TestNames_TheCanonicalSheet(t *testing.T) {
	t.Parallel()

	src := "Rate\t=0.08 |@ named(Rate)\n" +
		"Total\t=25 |@ named(Total)\n" +
		"Tax\t=@Total * @Rate\n"
	g := compute(t, src)
	assert.Equal(t, "0.08", cellAt(t, g, 0, 1))
	assert.Equal(t, "25", cellAt(t, g, 1, 1))
	assert.Equal(t, "2", cellAt(t, g, 2, 1))
}

// TestNames_BindingIsInertAtEvaluation states that the clause never changes
// the cell's own value: `=5 |@ named(X)` computes exactly what `=5` does.
func TestNames_BindingIsInertAtEvaluation(t *testing.T) {
	t.Parallel()

	g := compute(t, "=5 |@ named(X)\t=@X\n")
	assert.Equal(t, "5", cellAt(t, g, 0, 0))
	assert.Equal(t, "5", cellAt(t, g, 0, 1))
}

// TestNames_NamesAreCaseInsensitive states name identity follows
// function-name identity: `@RATE`, `@rate`, and `@Rate` are one name, and the
// clause's own function is case-insensitive too.
func TestNames_NamesAreCaseInsensitive(t *testing.T) {
	t.Parallel()

	g := compute(t, "=7 |@ NAMED(Rate)\t=@RATE\t=@rate\n")
	assert.Equal(t, "7", cellAt(t, g, 0, 1))
	assert.Equal(t, "7", cellAt(t, g, 0, 2))
}

// TestNames_UnboundIsName states the unbound answer: `@X` with no binding
// computes #NAME?, unambiguously "no such name".
func TestNames_UnboundIsName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "#NAME?", cellAt(t, compute(t, "=@Nowhere\n"), 0, 0))
}

// TestNames_DuplicateRefusesEveryUse states the duplicate answer: a name bound
// twice refuses every use as #NAME? — a wholesale refusal, never a silently
// chosen winner computing a plausible number from one of two cells.
func TestNames_DuplicateRefusesEveryUse(t *testing.T) {
	t.Parallel()

	src := "=1 |@ named(X)\t=2 |@ named(x)\t=@X\n"
	g := compute(t, src)
	assert.Equal(t, "1", cellAt(t, g, 0, 0), "the binding cells still compute their own values")
	assert.Equal(t, "2", cellAt(t, g, 0, 1))
	assert.Equal(t, "#NAME?", cellAt(t, g, 0, 2), "no duplicate silently wins")
}

// TestNames_SelfReferenceIsCirc states that `@X` inside the cell binding X is
// a self-dependency through the ordinary cycle machinery: #CIRC!.
func TestNames_SelfReferenceIsCirc(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "#CIRC!", cellAt(t, compute(t, "=1 + @Self |@ named(Self)\n"), 0, 0))
}

// TestNames_CycleThroughTwoNamesIsCirc states the transitive case: two cells
// reading each other by name are a cycle exactly as two cell references are.
func TestNames_CycleThroughTwoNamesIsCirc(t *testing.T) {
	t.Parallel()

	g := compute(t, "=@B |@ named(A)\t=@A |@ named(B)\n")
	assert.Equal(t, "#CIRC!", cellAt(t, g, 0, 0))
	assert.Equal(t, "#CIRC!", cellAt(t, g, 0, 1))
}

// TestNames_NameBindsTheCellNotTheStage states the meta clause's whole-cell
// semantics: after `=10/3 | round(2) |@ named(T)`, `@T` is the ROUNDED value —
// the cell's value — not the pre-round intermediate the clause happens to
// follow in the source.
func TestNames_NameBindsTheCellNotTheStage(t *testing.T) {
	t.Parallel()

	g := compute(t, "=10/3 | round(2) |@ named(T)\t=@T\n")
	assert.Equal(t, "3.33", cellAt(t, g, 0, 1))
}

// TestNames_MalformedClauseBindsNothing states that only a well-formed
// declaration binds: an unknown meta function, a bare `|@ named`, and a
// surplus-argument clause each bind nothing, so a use of the would-be name is
// #NAME? — loud, never quietly half-working. Check names each defect
// (check_names_test.go); this is the compute half of that contract.
func TestNames_MalformedClauseBindsNothing(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"unknown meta function": "=1 |@ label(X)\t=@X\n",
		"bare clause":           "=1 |@ named\t=@named\n",
		"surplus arguments":     "=1 |@ named(X, Y)\t=@X\n",
	}
	for name, src := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := compute(t, src)
			assert.Equal(t, "1", cellAt(t, g, 0, 0), "the cell's own value is untouched")
			assert.Equal(t, "#NAME?", cellAt(t, g, 0, 1))
		})
	}
}

// TestNames_MisplacedNamedCallIsName states the evaluator's answer to `named`
// written as an ordinary call or pipe stage: #NAME?, deterministic — the
// function exists only in the meta clause.
func TestNames_MisplacedNamedCallIsName(t *testing.T) {
	t.Parallel()

	g := compute(t, "=named(1, X)\t=1 | named\n")
	assert.Equal(t, "#NAME?", cellAt(t, g, 0, 0))
	assert.Equal(t, "#NAME?", cellAt(t, g, 0, 1))
}

// TestNames_EquivalentToPinnedReferences is acceptance criterion four: a sheet
// with names computes value-for-value what the same sheet computes with each
// clause deleted and each use replaced by the pinned reference it aliases.
func TestNames_EquivalentToPinnedReferences(t *testing.T) {
	t.Parallel()

	named := compute(t, "10\t=A1*2 |@ named(Double)\n=@Double + 1\t=sum(A1:B1)\n")
	pinned := compute(t, "10\t=A1*2\n=$B$1 + 1\t=sum(A1:B1)\n")
	assert.Equal(t, pinned, named)
}

// TestNames_SheetLocalThroughEmbedding states scope: an embedded sheet's
// names are its own. The child resolves its bindings; the parent referencing
// the same spelling has no binding and answers #NAME?.
func TestNames_SheetLocalThroughEmbedding(t *testing.T) {
	t.Parallel()

	g := embedGrid(t, "=sheet(\"child.tsvt\")\t=@Rate\n", map[string]string{
		"child.tsvt": "=6 |@ named(Rate)\t=@Rate * 2\t=output(B1)\n",
	})
	assert.Equal(t, "12", cellAt(t, g, 0, 0), "the child resolves its own names")
	assert.Equal(t, "#NAME?", cellAt(t, g, 0, 1), "the parent sees none of them")
}

// windowedLimits forces the windowed capability with a generous compute
// budget: resident refusal at 100 cells, everything else roomy.
func windowedLimits() engine.Limits {
	return engine.Limits{ResultCells: 1 << 20, GridDim: 1 << 20, ResultBytes: 1 << 20, ResidentCells: 100}
}

// TestNames_WindowedDocumentResolvesNames is R18 at the engine seam: the same
// file must answer the same `@name` at every size, so an over-budget document
// resolves a name bound thousands of rows away from the window that reads it —
// without materializing the document.
func TestNames_WindowedDocumentResolvesNames(t *testing.T) {
	t.Parallel()

	src := "=42 |@ named(Answer)\n" + strings.Repeat("x\ty\tz\n", 5000) + "=@Answer\n"
	_, windowed := open(t, src, windowedLimits())
	require.NotNil(t, windowed, "the document is over budget and windowed")

	g, err := windowed.ComputeRows(5001, 1, engine.ComputeOptions{Limits: windowedLimits()})
	require.NoError(t, err)
	assert.Equal(t, "42", g[0][0], "the use resolves a binding far outside the window")
}

// TestNames_WindowedMarkOverflowRefusesAsLimit states the cap's behaviour: a
// document with more candidate declaration rows than the resident budget
// answers every `@name` use #LIMIT! — the budget refusal, raisable like every
// budget (R16) — never a wrong or partial binding.
func TestNames_WindowedMarkOverflowRefusesAsLimit(t *testing.T) {
	t.Parallel()

	declarations := strings.Repeat("=1 |@ named(x)\n", 200) // 200 marked rows > 100 budget
	src := declarations + "=@x\n"
	_, windowed := open(t, src, windowedLimits())
	require.NotNil(t, windowed)

	g, err := windowed.ComputeRows(200, 1, engine.ComputeOptions{Limits: windowedLimits()})
	require.NoError(t, err)
	assert.Equal(t, "#LIMIT!", g[0][0])
}

// TestNames_WindowedMarkerInsideAStringBindsNothing states the
// over-approximation is filtered by the real parser: a cell whose STRING
// merely contains the marker's bytes declares nothing, and a use of the
// would-be name is #NAME?, exactly as resident.
func TestNames_WindowedMarkerInsideAStringBindsNothing(t *testing.T) {
	t.Parallel()

	src := "=\"a|@b\"\n" + strings.Repeat("x\ty\tz\n", 5000) + "=@b\n"
	_, windowed := open(t, src, windowedLimits())
	require.NotNil(t, windowed)

	g, err := windowed.ComputeRows(5001, 1, engine.ComputeOptions{Limits: windowedLimits()})
	require.NoError(t, err)
	assert.Equal(t, "#NAME?", g[0][0])
}

// TestNames_WindowedBindingReadFailureIsAnError states the failure leg: a
// source that dies while the marked declaration row is fetched fails the
// whole ComputeRows call as ErrReadInput — a broken source is an error, not a
// window computed over partial bindings.
func TestNames_WindowedBindingReadFailureIsAnError(t *testing.T) {
	t.Parallel()

	// The declaration sits in block one; 300 filler rows push the window past
	// the first stride, so arming a failure below byte 4 breaks only the
	// binding fetch.
	doc := "=7 |@ named(Seed)\n" + strings.Repeat("x\n", 300) + "=@Seed\n"
	src := &regionFailSource{data: []byte(doc)}

	_, windowed, err := engine.OpenSheet(engine.ByteSource{ReadAt: src, Size: int64(len(src.data))}, windowedLimits())
	require.NoError(t, err)
	require.NotNil(t, windowed)

	src.failBelow = 4
	_, err = windowed.ComputeRows(301, 1, engine.ComputeOptions{Limits: windowedLimits()})
	assert.ErrorIs(t, err, constants.ErrReadInput)
}

// TestNames_SheetNamesListsDeclarations states the editor-facing list: names
// in binding order, declared spelling, one entry per identity however many
// cells duplicate it.
func TestNames_SheetNamesListsDeclarations(t *testing.T) {
	t.Parallel()

	s, err := engine.Parse([]byte("=1 |@ named(Rate)\t=2 |@ named(zeta)\n=3 |@ named(RATE)\n"))
	require.NoError(t, err)
	assert.Equal(t, []string{"Rate", "zeta"}, s.Names(),
		"binding order, declared spelling, the case-variant duplicate folded")
}
