package engine_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixedClock is an arbitrary but stable instant; ComputeAt(fixedClock) makes the
// pass PRNG reproducible.
func fixedClock() time.Time { return time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC) }

// numAt evaluates =expr in A1 at the fixed clock and returns the parsed float.
func numAt(t *testing.T, expr string) float64 {
	t.Helper()
	s := parse(t, "="+expr+"\n")
	n, err := strconv.ParseFloat(s.ComputeAt(fixedClock())[0][0], 64)
	require.NoError(t, err)
	return n
}

func TestRand_RangeAndDeterminism(t *testing.T) {
	t.Parallel()
	s := parse(t, "=rand()\n")

	first := s.ComputeAt(fixedClock())[0][0]
	again := s.ComputeAt(fixedClock())[0][0]
	assert.Equal(t, first, again, "same clock reproduces the draw (one engine, R1)")

	n, err := strconv.ParseFloat(first, 64)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, n, 0.0)
	assert.Less(t, n, 1.0)

	// "random" is an alias of "rand".
	assert.Equal(t, first, parse(t, "=random()\n").ComputeAt(fixedClock())[0][0])
}

func TestRand_EachDrawDiffers(t *testing.T) {
	t.Parallel()
	// Two draws in one pass pull successive PRNG values, so they differ.
	g := parse(t, "=rand()\t=rand()\n").ComputeAt(fixedClock())
	assert.NotEqual(t, g[0][0], g[0][1])
}

func TestRand_Arity(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "#VALUE!", formula1(t, "rand(1)"))
}

func TestRandbetween(t *testing.T) {
	t.Parallel()

	n := numAt(t, "randbetween(1, 6)")
	assert.GreaterOrEqual(t, n, 1.0)
	assert.LessOrEqual(t, n, 6.0)
	assert.Equal(t, n, float64(int64(n)), "an integer")

	assert.Equal(t, 5.0, numAt(t, "randbetween(5, 5)"))

	cases := map[string]struct{ expr, want string }{
		"hi below lo":    {"randbetween(6, 1)", "#NUM!"},
		"non-numeric lo": {"randbetween(\"x\", 6)", "#VALUE!"},
		"non-numeric hi": {"randbetween(1, \"y\")", "#VALUE!"},
		"one argument":   {"randbetween(1)", "#VALUE!"},
		"three argument": {"randbetween(1, 2, 3)", "#VALUE!"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, formula1(t, tc.expr))
		})
	}
}

func TestRandarray_Spills(t *testing.T) {
	t.Parallel()
	g := parse(t, "=randarray(2, 3)\n\n\n").ComputeAt(fixedClock())
	require.GreaterOrEqual(t, len(g), 2)
	for row := range 2 {
		require.GreaterOrEqual(t, len(g[row]), 3)
		for col := range 3 {
			n, err := strconv.ParseFloat(g[row][col], 64)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, n, 0.0)
			assert.Less(t, n, 1.0)
		}
	}
}

func TestRandarray_Errors(t *testing.T) {
	t.Parallel()
	cases := map[string]struct{ expr, want string }{
		"zero rows":        {"randarray(0, 1)", "#VALUE!"},
		"negative cols":    {"randarray(1, -2)", "#VALUE!"},
		"no arguments":     {"randarray()", "#VALUE!"},
		"too many":         {"randarray(1, 2, 3)", "#VALUE!"},
		"exceeds budget":   {"randarray(3000, 3000)", "#VALUE!"},
		"non-numeric rows": {"randarray(\"x\", 2)", "#VALUE!"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, formula1(t, tc.expr))
		})
	}
}

func TestRandarray_DefaultsToOneColumn(t *testing.T) {
	t.Parallel()
	g := parse(t, "=randarray(2)\n\n\n").ComputeAt(fixedClock())
	require.GreaterOrEqual(t, len(g), 2)
	n, err := strconv.ParseFloat(g[0][0], 64)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, n, 0.0)
	assert.Less(t, n, 1.0)
}

func TestTick(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "7", parse(t, "=tick()\n").ComputeAtTick(fixedClock(), 7)[0][0])
	assert.Equal(t, "7", parse(t, "=frame()\n").ComputeAtTick(fixedClock(), 7)[0][0])
	// ComputeAt injects ordinal 0.
	assert.Equal(t, "0", parse(t, "=tick()\n").ComputeAt(fixedClock())[0][0])
	// tick takes no arguments.
	assert.Equal(t, "#VALUE!", formula1(t, "tick(1)"))
}
