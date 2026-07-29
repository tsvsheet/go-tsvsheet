// Package engine's argument-shaping layer: how a call's operands are reduced to
// the values a builtin receives — a scalar slot, a flattened range, or the
// mixed shapes the recurring parameter specs describe.
package engine

import "github.com/tsvsheet/go-tsvsheet/internal/tsvt"

// argMode selects how an eager parameter slot consumes its operand (ADR 0004
// §2): a scalar slot yields exactly one value, a cells slot flattens a range
// or array operand row-major into the argument list.
type argMode int

const (
	modeScalar argMode = iota
	modeCells
)

// paramModes declares a function's parameter slots: lead modes bind the first
// arguments, tail modes the last, and rest every slot between. The zero value
// is all-scalar — the default for positional (scalar-parameter) functions.
type paramModes struct {
	lead []argMode
	tail []argMode
	rest argMode
}

// mode is the declared mode of argument slot i in a call of n arguments.
func (p paramModes) mode(i argIndex, n argCount) argMode {
	if int(i) < len(p.lead) {
		return p.lead[i]
	}
	if tailAt := int(n) - len(p.tail); int(i) >= tailAt {
		return p.tail[int(i)-tailAt]
	}
	return p.rest
}

// The recurring parameter shapes: an aggregate flattens every slot (SUM,
// AND, …), LARGE/SMALL flatten their values but keep the trailing k scalar,
// and NPV keeps its leading rate scalar ahead of the flattened cashflows.
var (
	cellsRest       = paramModes{rest: modeCells}
	cellsThenK      = paramModes{rest: modeCells, tail: []argMode{modeScalar}}
	scalarThenCells = paramModes{lead: []argMode{modeScalar}, rest: modeCells}
)

// argValues materializes call arguments per the declared parameter modes: a
// cells slot contributes every cell of a range or array operand (§11.3), a
// scalar slot exactly one value — so a multi-cell operand can never shift the
// arguments that follow it (go-tsvsheet#2).
func (r resolver) argValues(args []tsvt.Expr, spec paramModes) []Value {
	values := make([]Value, 0, len(args))
	for i, arg := range args {
		if spec.mode(argIndex(i), argCount(len(args))) == modeCells {
			values = append(values, r.argCells(arg)...)
			continue
		}
		values = append(values, r.argScalar(arg))
	}
	return values
}

// argScalar evaluates one argument in scalar context (ADR 0004 §2): eval
// already reduces a multi-cell range to #VALUE! (cellset.scalar), and an
// array reduces to its top-left element per the pinned no-broadcasting rule.
func (r resolver) argScalar(arg tsvt.Expr) Value {
	v := r.eval(arg)
	if v.kind == kindArray {
		return v.arr[0][0]
	}
	return v
}

// argCells expands one argument: a bare reference contributes all its resolved
// cells (so `sum(A:H)` sees the whole range), and an expression that evaluates
// to an array contributes its elements row-major — consumed exactly like a
// range, so `sum(sort(A1:A3))` aggregates (ADR 0004 §2 array-valued
// arguments). Any other expression is one scalar value.
func (r resolver) argCells(arg tsvt.Expr) []Value {
	if ref, ok := arg.(tsvt.RefOperand); ok {
		return r.resolveOperand(ref.Ref).values
	}
	v := r.eval(arg)
	if v.kind == kindArray {
		return flatten1D(v.arr)
	}
	return []Value{v}
}
