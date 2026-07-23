package engine

import (
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// IsVolatile reports whether any formula wraps an expression in volatile(…),
// the sole marker that a cell's computed value changes over time and a frontend
// should recompute. The clock functions today/now/isnow are volatile only when
// wrapped — nothing is volatile without volatile().
func (s Sheet) IsVolatile() bool {
	return len(s.VolatileSchedules()) > 0
}

// VolatileSchedules returns one refresh-cadence spec per volatile(…) call across
// the sheet: an explicit string second argument if present, else the pattern of
// a wrapped isnow("pattern"), else "" (the frontend's default cadence). A
// frontend unions the set to the soonest next instant.
func (s Sheet) VolatileSchedules() []string {
	var schedules []string
	s.eachFormula(func(at Address) {
		walkCalls(s.cells[at.Row][at.Col].formula, func(call tsvt.Call) {
			if strings.ToLower(call.Name) == fnVolatile {
				schedules = append(schedules, volatileSchedule(call))
			}
		})
	})
	return schedules
}

// volatileSchedule derives one volatile call's cadence: an explicit literal
// second argument wins; otherwise a wrapped isnow("pattern") lends its pattern;
// otherwise "" for the default cadence.
func volatileSchedule(call tsvt.Call) string {
	if spec, ok := explicitSchedule(call.Args); ok {
		return spec
	}
	if spec, ok := derivedSchedule(call.Args); ok {
		return spec
	}
	return ""
}

// explicitSchedule reads a literal string second argument of
// volatile(expr, "spec").
func explicitSchedule(args []tsvt.Expr) (string, boolResult) {
	if len(args) < 2 {
		return "", false
	}
	lit, ok := args[1].(tsvt.StringLit)
	return lit.Value, boolResult(ok)
}

// derivedSchedule reads the pattern of a wrapped volatile(isnow("pattern")).
func derivedSchedule(args []tsvt.Expr) (string, boolResult) {
	if len(args) < 1 {
		return "", false
	}
	call, ok := args[0].(tsvt.Call)
	if !ok || strings.ToLower(call.Name) != fnIsnow || len(call.Args) < 1 {
		return "", false
	}
	lit, ok := call.Args[0].(tsvt.StringLit)
	return lit.Value, boolResult(ok)
}
