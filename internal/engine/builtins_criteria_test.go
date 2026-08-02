package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// A wholesale range refusal must propagate through the criteria family — the
// lazily-dispatched consumers that step over per-cell errors by design. Without
// the out-of-band refusal, a refused range flattens to one non-matching cell
// and the aggregate launders the refusal into a plausible answer: countif over
// an unreadable range answering 0, averageif answering #DIV/0!.

// TestCriteria_RefusedRangePropagates pins refusal propagation for every
// criteria shape: the over-budget span (#LIMIT!), the unaddressable range
// (#REF!), the unresolvable `"file"!` target (#REF!), and the sum-range slot.
func TestCriteria_RefusedRangePropagates(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`countif(A1:A50000000000, ">0")`:         string(engine.ErrLimit),
		`sumif(A1:A50000000000, ">0")`:           string(engine.ErrLimit),
		`averageif(A1:A50000000000, ">0")`:       string(engine.ErrLimit),
		`sumif(A2:D2, ">0", A1:A50000000000)`:    string(engine.ErrLimit), // refused SUM range
		`countif(ZZZZZZZZZ1:ZZZZZZZZZ9, ">0")`:   string(engine.ErrRef),   // unaddressable column
		`countif("missing"!A1:B2, ">0")`:         string(engine.ErrRef),   // no loader → unresolvable
		`sumif(A2:D2, ">0", "missing"!A1:B2)`:    string(engine.ErrRef),
		`averageif(ZZZZZZZZZ1:ZZZZZZZZZ9, ">0")`: string(engine.ErrRef),
		`countif(sort(A1:A50000000000), ">0")`:   string(engine.ErrLimit), // refusal through a nested lazy call
		`countif(sum(A1:A50000000000), ">0")`: string(
			engine.ErrLimit,
		), // refusal riding through a nested EAGER call
		`countif(if(TRUE, A0, 0), ">0")`: string(
			engine.ErrRef,
		), // single-cell refusal through a dispatcher branch
		`countif(if(TRUE, ZZZZZZZZZ1, 0), ">0")`: string(engine.ErrRef),
		`countif(if(TRUE, "missing"!A1, 0), ">0")`: string(
			engine.ErrRef,
		), // foreign single-cell refusal keeps its marker
		`countif(A1:DEJTLX1, ">0")`:               string(engine.ErrLimit),
		`sumif(A1:A50000000000, ">0", A2:D2)`:     string(engine.ErrLimit), // refused CELL range, sane sum range
		`averageif(A1:A50000000000, ">0", A2:D2)`: string(engine.ErrLimit),
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, formula1(t, expr))
		})
	}
}

// TestCriteria_PerCellErrorsAreSteppedOverNotPropagated pins the other half of
// the refusal distinction: an error held by one cell WITHIN a resolved range is
// stepped over — the criteria family's hole tolerance for ragged grids.
func TestCriteria_PerCellErrorsAreSteppedOverNotPropagated(t *testing.T) {
	t.Parallel()

	// B2 is out of the ragged grid: countif's range resolves with a #REF! hole.
	ragged := "1\t2\t=countif(A1:B2, \">0\")\n3\n"
	assert.Equal(t, "3", cellAt(t, compute(t, ragged), 0, 2))

	// A1 holds the literal error; the 1×1 range RESOLVES, so nothing propagates.
	literal := "#LIMIT!\t=countif(A1:A1, \">0\")\n"
	assert.Equal(t, "0", cellAt(t, compute(t, literal), 0, 1))
}
