package engine_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

func TestDefaultLimits_Values(t *testing.T) {
	t.Parallel()

	l := engine.DefaultLimits()
	assert.Equal(t, 5_000_000, l.ResultCells)
	assert.Equal(t, 1_000_000, l.GridDim)
	assert.Equal(t, 1<<20, l.ResultBytes)
	assert.Equal(t, 5_000_000, l.SpanCells)
}

func TestBrowserLimits_Values(t *testing.T) {
	t.Parallel()

	l := engine.BrowserLimits()
	assert.Equal(t, 100_000, l.ResultCells)
	assert.Equal(t, 20_000, l.GridDim)
	assert.Equal(t, 64<<10, l.ResultBytes)
	assert.Equal(t, 1_000_000, l.SpanCells)
}

// computeWithCell parses src, computes it with the injected limits, and returns
// cell A1's computed value.
func computeWithCell(t *testing.T, src string, limits engine.Limits) string {
	t.Helper()
	s, err := engine.Parse([]byte(src))
	require.NoError(t, err)
	return s.ComputeWith(engine.ComputeOptions{At: time.Now(), Limits: limits})[0][0]
}

func TestComputeWith_ZeroLimitsFallBackToDefault(t *testing.T) {
	t.Parallel()

	// A zero (unset) Limits is treated as DefaultLimits: a modest array and a
	// modest REPT both compute — a degenerate zero cap would reject them.
	assert.Equal(t, "1", computeWithCell(t, "=sequence(2, 2)\n", engine.Limits{}))
	assert.Equal(t, "aaa", computeWithCell(t, "=rept(\"a\", 3)\n", engine.Limits{}))
}

func TestComputeWith_HonorsInjectedCellBudget(t *testing.T) {
	t.Parallel()

	tiny := engine.Limits{ResultCells: 5, GridDim: 5, ResultBytes: 5}
	assert.Equal(t, "1", computeWithCell(t, "=sequence(2, 2)\n", tiny))                     // 4 cells <= 5
	assert.Equal(t, string(engine.ErrLimit), computeWithCell(t, "=sequence(3, 3)\n", tiny)) // 9 > 5: budget refusal
}

func TestComputeWith_HonorsInjectedByteBudget(t *testing.T) {
	t.Parallel()

	tiny := engine.Limits{ResultCells: 5, GridDim: 5, ResultBytes: 5}
	assert.Equal(t, "aaaaa", computeWithCell(t, "=rept(\"a\", 5)\n", tiny))                 // 5 bytes <= 5
	assert.Equal(t, string(engine.ErrLimit), computeWithCell(t, "=rept(\"a\", 6)\n", tiny)) // 6 > 5: budget refusal
}

// TestComputeWith_RangeSpanOverBudgetIsLimit pins the span budget at the range
// choke point: a reference whose written rectangle holds more cells than the
// budget is #LIMIT! before any allocation or read — a written reference can
// never drive the materialization it names. In-budget ranges keep today's
// semantics exactly (the 6-cell sum below still computes under a 6-cell
// budget), and out-of-grid corners inside the budget stay #REF!.
func TestComputeWith_RangeSpanOverBudgetIsLimit(t *testing.T) {
	t.Parallel()

	grid := "1\t2\t3\n4\t5\t6\n=sum(A1:C2)\t=sum(A1:C2)\t=index(A1:C2, 1, 1)\n"
	sixCells := engine.Limits{ResultCells: 6, GridDim: 10, ResultBytes: 100}
	fiveCells := engine.Limits{ResultCells: 5, GridDim: 10, ResultBytes: 100}

	s, err := engine.Parse([]byte(grid))
	require.NoError(t, err)
	within := s.ComputeWith(engine.ComputeOptions{At: time.Now(), Limits: sixCells})
	assert.Equal(t, "21", within[2][0])
	assert.Equal(t, "1", within[2][2])

	over := s.ComputeWith(engine.ComputeOptions{At: time.Now(), Limits: fiveCells})
	assert.Equal(t, string(engine.ErrLimit), over[2][0], "cells path (sum) refuses the 6-cell rectangle")
	assert.Equal(t, string(engine.ErrLimit), over[2][2], "matrix path (index) refuses the same rectangle")
}

// TestSpanAndResultBudgetsAreSeparateCeilings pins that SpanCells and
// ResultCells bound different costs: with a 5-cell result budget and a 10-cell
// span budget, reading a 6-cell rectangle computes (it is a READ the span
// budget authorises) while a 9-cell SEQUENCE result is still refused (it
// WRITES cells the result budget bounds). Sharing one ceiling broke the
// shipped BrowserLimits: a grid its GridDim explicitly permits could not sum
// its own columns.
func TestSpanAndResultBudgetsAreSeparateCeilings(t *testing.T) {
	t.Parallel()

	grid := "1\t2\t3\n4\t5\t6\n=sum(A1:C2)\t=sequence(3, 3)\n"
	split := engine.Limits{ResultCells: 5, GridDim: 10, ResultBytes: 100, SpanCells: 10}

	s, err := engine.Parse([]byte(grid))
	require.NoError(t, err)
	g := s.ComputeWith(engine.ComputeOptions{At: time.Now(), Limits: split})
	assert.Equal(t, "21", g[2][0], "a 6-cell read is within the 10-cell span budget")
	assert.Equal(t, string(engine.ErrLimit), g[2][1], "a 9-cell result still exceeds the 5-cell result budget")
}

// TestSpanBudgetTreatsANonPositiveSpanCellsAsUnsetNeverAsRefuseEverything pins
// spanBudget's fallback contract: zero or negative SpanCells means "unset, use
// ResultCells", so a pre-SpanCells caller (or a nonsense negative) keeps a
// working ceiling instead of a budget that refuses every reference.
func TestSpanBudgetTreatsANonPositiveSpanCellsAsUnsetNeverAsRefuseEverything(t *testing.T) {
	t.Parallel()

	for name, spanCells := range map[string]int{"zero": 0, "negative": -1} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			l := engine.Limits{ResultCells: 10, GridDim: 10, ResultBytes: 100, SpanCells: spanCells}
			got := computeWithCell(t, "=sum(A2:C3)\t\n1\t2\t3\n4\t5\t6\n", l)
			assert.Equal(t, "21", got, "a 6-cell span computes under the 10-cell fallback")
		})
	}
}

// TestCompute_GiantWrittenRangeIsLimitNotOOM pins the work-order acceptance
// (015): the literal rectangles below — 50 billion rows; DEJTLX is column
// ~50 million in bijective base-26 — used to be materialized value-by-value
// before any check, OOM-killing the process from a single formula on a tiny
// file. The span budget refuses them from the corners alone under
// DefaultLimits, in O(1) memory, on both resolution paths. Each compute runs
// under a deadline so a regression FAILS with a diagnosis instead of
// exhausting the machine and presenting as a runner death.
func TestCompute_GiantWrittenRangeIsLimitNotOOM(t *testing.T) {
	t.Parallel()

	for name, expr := range map[string]string{
		"tall, cells path": "sum(A1:A50000000000)",
		"wide, cells path": "sum(A1:DEJTLX1)",
		"matrix path":      "index(A1:ZZ99999999, 1, 1)",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := make(chan string, 1)
			go func() {
				s, err := engine.Parse([]byte("2\t=" + expr + "\t3\t4\n"))
				if err != nil {
					got <- "parse: " + err.Error()
					return
				}
				got <- s.Compute()[0][1]
			}()
			select {
			case v := <-got:
				assert.Equal(t, string(engine.ErrLimit), v)
			case <-time.After(10 * time.Second):
				t.Fatal("no refusal within 10s — the written rectangle is being materialized")
			}
		})
	}
}

func TestSet_HonorsInjectedGridLimit(t *testing.T) {
	t.Parallel()

	s, err := engine.Parse([]byte("1\n"))
	require.NoError(t, err)

	tiny := engine.Limits{ResultCells: 5, GridDim: 5, ResultBytes: 5}
	// Within the grid dimension the edit grows the grid; at or beyond it the edit
	// is rejected before growing (the OOM guard).
	_, err = s.Set(engine.Address{Row: 4, Col: 0}, "x", tiny)
	require.NoError(t, err)
	_, err = s.Set(engine.Address{Row: 5, Col: 0}, "x", tiny)
	assert.ErrorIs(t, err, constants.ErrInvalidValue)
}

// TestMaxSafeMagnitudeRefusesWhatIntegerArithmeticCannotCarry pins the guard
// that stands between a stored number and five different panics. A cell holding
// 9.3e18 is not exotic — a nanosecond timestamp multiplied by anything, or a
// scientific export, reaches it — and Go's out-of-range float-to-int conversion
// saturates rather than failing, so every range check downstream passed on a
// value that had already become nonsense.
func TestMaxSafeMagnitudeRefusesWhatIntegerArithmeticCannotCarry(t *testing.T) {
	t.Parallel()
	for _, formula := range []string{
		`=mid("abc",A1,A1)`, `=sequence(A1,A1)`, `=randarray(A1,A1)`,
		`=rept("ab",A1)`, `=randbetween(-A1,A1)`, `=left("abc",A1)`, `=right("abc",A1)`,
	} {
		sheet, err := engine.Parse([]byte("9.3e18\n" + formula + "\n"))
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			// A saturated magnitude is refused at the conversion (#VALUE!/#NUM!)
			// or, where the conversion admits it, by the cell budget (#LIMIT!).
			assert.Contains(t, []string{"#VALUE!", "#NUM!", "#LIMIT!"}, sheet.Compute()[1][0], formula)
		}, formula)
	}
}

// TestTooManyCellsCountsADimensionThatAloneExceedsTheBudget pins the overflow
// its own doc comment used to deny. Two dimensions of maxInt multiply to 1 in
// int64, which is under every budget, so a request for 9.2 quintillion cells
// was authorised and the allocation panicked.
func TestTooManyCellsCountsADimensionThatAloneExceedsTheBudget(t *testing.T) {
	t.Parallel()
	sheet, err := engine.Parse([]byte("9.3e18\n=sequence(A1,A1)\n"))
	require.NoError(t, err)

	assert.NotEqual(t, "", sheet.Compute()[1][0], "a refusal, not an allocation")
}

// TestEffectiveResidentCellsMatchesTheLoadPolicy pins the exported ceiling
// against the policy the loads actually apply: zero value → the default,
// dedicated budget wins, ResultCells is the single-ceiling fallback.
func TestEffectiveResidentCellsMatchesTheLoadPolicy(t *testing.T) {
	t.Parallel()

	assert.Equal(t, int64(engine.DefaultLimits().ResidentCells), engine.Limits{}.EffectiveResidentCells())
	assert.Equal(t, int64(7), engine.Limits{ResidentCells: 7, ResultCells: 9}.EffectiveResidentCells())
	assert.Equal(t, int64(9), engine.Limits{ResultCells: 9}.EffectiveResidentCells())
}
