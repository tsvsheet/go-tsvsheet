package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

func TestSeries_MovingAvgSpillsWithHonestWindows(t *testing.T) {
	t.Parallel()

	// Positions with fewer than span values are #N/A — never partial averages.
	g := compute(t, "1\t=movingavg(A1:A4, 2)\n2\n3\n4\n")
	assert.Equal(t, string(engine.ErrNA), cellAt(t, g, 0, 1))
	assert.Equal(t, "1.5", cellAt(t, g, 1, 1))
	assert.Equal(t, "2.5", cellAt(t, g, 2, 1))
	assert.Equal(t, "3.5", cellAt(t, g, 3, 1))
}

func TestSeries_EmaSeedsWithFirstElement(t *testing.T) {
	t.Parallel()

	// span 3 → α = 0.5: 2, then 0.5·4+0.5·2 = 3, then 0.5·6+0.5·3 = 4.5.
	g := compute(t, "2\t=ema(A1:A3, 3)\n4\n6\n")
	assert.Equal(t, "2", cellAt(t, g, 0, 1))
	assert.Equal(t, "3", cellAt(t, g, 1, 1))
	assert.Equal(t, "4.5", cellAt(t, g, 2, 1))
}

func TestSeries_RollingMinMax(t *testing.T) {
	t.Parallel()

	lo := compute(t, "3\t=rollingmin(A1:A3, 2)\n1\n2\n")
	assert.Equal(t, string(engine.ErrNA), cellAt(t, lo, 0, 1))
	assert.Equal(t, "1", cellAt(t, lo, 1, 1))
	assert.Equal(t, "1", cellAt(t, lo, 2, 1))

	hi := compute(t, "3\t=rollingmax(A1:A3, 2)\n1\n2\n")
	assert.Equal(t, string(engine.ErrNA), cellAt(t, hi, 0, 1))
	assert.Equal(t, "3", cellAt(t, hi, 1, 1))
	assert.Equal(t, "2", cellAt(t, hi, 2, 1))
}

func TestSeries_CumsumRunsTheTotal(t *testing.T) {
	t.Parallel()

	g := compute(t, "1\t=cumsum(A1:A3)\n2\n3\n")
	assert.Equal(t, "1", cellAt(t, g, 0, 1))
	assert.Equal(t, "3", cellAt(t, g, 1, 1))
	assert.Equal(t, "6", cellAt(t, g, 2, 1))
}

func TestSeries_FlattensRowMajorAndComposes(t *testing.T) {
	t.Parallel()

	// A 2-D range flattens row-major: A1:B2 = 1,2,3,4 → cumsum 1,3,6,10.
	g := compute(t, "1\t2\t=cumsum(A1:B2)\n3\t4\n")
	assert.Equal(t, "1", cellAt(t, g, 0, 2))
	assert.Equal(t, "3", cellAt(t, g, 1, 2))
	assert.Equal(t, "6", cellAt(t, g, 2, 2))
	assert.Equal(t, "10", cellAt(t, g, 3, 2))

	// An array argument is consumed exactly like a range (ADR 0004 §2).
	c := compute(t, "=cumsum(sequence(3))\n")
	assert.Equal(t, "1", cellAt(t, c, 0, 0))
	assert.Equal(t, "3", cellAt(t, c, 1, 0))
	assert.Equal(t, "6", cellAt(t, c, 2, 0))
}

func TestSeries_SpanBeyondSeriesIsAllNA(t *testing.T) {
	t.Parallel()

	g := compute(t, "1\t=rollingmin(A1:A2, 5)\n2\n")
	assert.Equal(t, string(engine.ErrNA), cellAt(t, g, 0, 1))
	assert.Equal(t, string(engine.ErrNA), cellAt(t, g, 1, 1))
}

func TestSeries_SpanTruncatesToInteger(t *testing.T) {
	t.Parallel()

	// 2.9 truncates to a span of 2.
	g := compute(t, "1\t=movingavg(A1:A3, 2.9)\n2\n3\n")
	assert.Equal(t, string(engine.ErrNA), cellAt(t, g, 0, 1))
	assert.Equal(t, "1.5", cellAt(t, g, 1, 1))
	assert.Equal(t, "2.5", cellAt(t, g, 2, 1))
}

func TestSeries_BooleansCoerceInTheSeries(t *testing.T) {
	t.Parallel()

	g := compute(t, "=TRUE\t=cumsum(A1:A2)\n=TRUE\n")
	assert.Equal(t, "1", cellAt(t, g, 0, 1))
	assert.Equal(t, "2", cellAt(t, g, 1, 1))
}

func TestSeries_Errors(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		src  string
		want engine.ErrorValue
	}{
		"missing span is wrong arity":  {"5\t=movingavg(A1:A2)\n6\n", engine.ErrValue},
		"cumsum takes no span":         {"5\t=cumsum(A1:A2, 2)\n6\n", engine.ErrValue},
		"zero span":                    {"5\t=movingavg(A1:A2, 0)\n6\n", engine.ErrNum},
		"negative span":                {"5\t=ema(A1:A2, -1)\n6\n", engine.ErrNum},
		"non-numeric span":             {"5\t=rollingmax(A1:A2, \"x\")\n6\n", engine.ErrValue},
		"span error propagates":        {"5\t=movingavg(A1:A2, Z99)\n6\n", engine.ErrRef},
		"empty element breaks density": {"1\t=cumsum(A1:A3)\n\n3\n", engine.ErrValue},
		"text element":                 {"1\t=cumsum(A1:A2)\nfoo\n", engine.ErrValue},
		"error element propagates":     {"=1/0\t=cumsum(A1:A2)\n2\n", engine.ErrDiv},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, string(tc.want), cellAt(t, compute(t, tc.src), 0, 1))
		})
	}
}

func TestCheck_KnowsAdoptedFamilies(t *testing.T) {
	t.Parallel()

	// The static checker and the evaluator agree on the ADR 0011 names.
	s, err := engine.Parse([]byte("=cumsum(A2:A3)\t=digest(A2)\t=jsonlen(\"[]\")\n1\n2\n"))
	require.NoError(t, err)
	assert.Empty(t, engine.Check(s))
}
