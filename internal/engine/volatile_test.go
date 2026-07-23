package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

func TestIsVolatile_OnlyThroughWrapper(t *testing.T) {
	t.Parallel()

	// volatile(…) is the sole marker.
	assert.True(t, parse(t, "=volatile(rand())\n").IsVolatile())
	assert.True(t, parse(t, "=volatile(isnow(\"noon\"))\n").IsVolatile())
	assert.True(t, parse(t, "=volatile(sqrt(2))\n").IsVolatile()) // pointless but legal

	// Bare clock calls are no longer volatile — nothing is, without the wrapper.
	assert.False(t, parse(t, "=now()\n").IsVolatile())
	assert.False(t, parse(t, "=today()\n").IsVolatile())
	assert.False(t, parse(t, "=isnow(\"noon\")\n").IsVolatile())
	assert.False(t, parse(t, "=sum(A1:A2)\n").IsVolatile())
	assert.False(t, parse(t, "5\t=A1 + 1\n").IsVolatile())
}

func TestVolatileSchedules(t *testing.T) {
	t.Parallel()

	assert.Empty(t, parse(t, "=isnow(\"noon\")\n").VolatileSchedules())
	assert.Equal(t, []string{""}, parse(t, "=volatile(rand())\n").VolatileSchedules())
	assert.Equal(t, []string{"5m"}, parse(t, "=volatile(rand(), \"5m\")\n").VolatileSchedules())
	// A wrapped isnow lends its pattern as the cadence.
	assert.Equal(t, []string{"M-F noon"}, parse(t, "=volatile(isnow(\"M-F noon\"))\n").VolatileSchedules())
	// The pipe form desugars to volatile(isnow("noon")) and derives likewise.
	assert.Equal(t, []string{"noon"}, parse(t, "=isnow(\"noon\") | volatile()\n").VolatileSchedules())
	// A non-literal second argument falls back to the derived pattern.
	assert.Equal(t, []string{"noon"}, parse(t, "=volatile(isnow(\"noon\"), A1)\n").VolatileSchedules())

	both := parse(t, "=volatile(rand())\t=volatile(now(), \"1s\")\n").VolatileSchedules()
	assert.ElementsMatch(t, []string{"", "1s"}, both)
}

func TestVolatile_TransparentValue(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		expr string
		want string
	}{
		"scalar passes through":     {"volatile(sqrt(4))", "2"},
		"schedule arg is ignored":   {"volatile(6, \"5m\")", "6"},
		"error propagates":          {"volatile(1/0)", "#DIV/0!"},
		"missing argument is value": {"volatile()", "#VALUE!"},
		"too many arguments":        {"volatile(1, \"5m\", 3)", "#VALUE!"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, formula1(t, tc.expr))
		})
	}
}

func TestVolatile_TransparentArray(t *testing.T) {
	t.Parallel()
	// volatile must not collapse a spilled array to a scalar.
	g := compute(t, "=volatile(sequence(3,1))\n\n\n")
	assert.Equal(t, "1", cellAt(t, g, 0, 0))
	assert.Equal(t, "2", cellAt(t, g, 1, 0))
	assert.Equal(t, "3", cellAt(t, g, 2, 0))
}

func TestCheck_KnowsVolatileAndGenerators(t *testing.T) {
	t.Parallel()
	s := parse(t, "=volatile(rand())\t=randbetween(1,6)\t=randarray(2,2)\t=tick()\t=frame()\n")
	assert.Empty(t, engine.Check(s))
}
