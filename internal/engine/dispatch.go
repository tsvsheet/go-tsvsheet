// Package engine's lazy-dispatch families: the calls whose arguments are NOT
// evaluated eagerly — clock readings, conditionals that must not evaluate the
// untaken branch, kind inspectors, and the text family.
package engine

import (
	isnow "github.com/tsvsheet/go-isnow"

	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

func (r resolver) evalClock(name funcName, args []tsvt.Expr) (Value, boolResult) {
	switch name {
	case fnToday:
		return clockResult(argCount(len(args)), dateValue(daySerial(r.comp.now))), true
	case fnNow:
		return clockResult(argCount(len(args)), dateValue(datetimeSerial(r.comp.now))), true
	case fnIsnow:
		return r.evalIsnow(args), true
	default:
		return Value{}, false
	}
}

// evalIsnow tests whether an isnow date/time pattern (tsvsheet/isnow) holds at the
// compute clock: ISNOW("M-F noon") is TRUE when the pattern matches the current
// pass instant. A malformed pattern, or the wrong arity, is #VALUE!.
func (r resolver) evalIsnow(args []tsvt.Expr) Value {
	if len(args) != 1 {
		return errorValue(ErrValue)
	}
	pattern := r.eval(args[0])
	if pattern.isError() {
		return pattern
	}
	holds, err := isnow.Is(isnow.PatternText(pattern.String()), r.comp.now)
	if err != nil {
		return errorValue(ErrValue)
	}
	return boolValue(boolResult(holds))
}

// clockResult returns v for a no-argument call, else #VALUE!.
func clockResult(argc argCount, v Value) Value {
	if argc != 0 {
		return errorValue(ErrValue)
	}
	return v
}

// evalConditional handles the selectively-lazy conditionals, which evaluate
// only the arguments they need. ok is false for a non-conditional name.
func (r resolver) evalConditional(name funcName, args []tsvt.Expr) (Value, boolResult) {
	switch name {
	case "if":
		return r.evalIf(args), true
	case "ifs":
		return r.evalIfs(args), true
	case "iferror":
		return r.evalIferror(args, false), true
	case "ifna":
		return r.evalIferror(args, true), true
	case "switch":
		return r.evalSwitch(args), true
	default:
		return Value{}, false
	}
}

// isConditional reports whether name is one of the lazy conditional builtins.
func isConditional(name funcName) boolResult { return isAmong(name, conditionalFns) }

// conditionalFns are the lazily-evaluated branch builtins.
var conditionalFns = []funcName{"if", "ifs", "iferror", "ifna", "switch"}

// evalInspector handles the single-argument inspectors (`IS*`, `N`, `TYPE`): it
// evaluates the argument (observing an error or empty result) and applies the
// pure inspector function.
func (r resolver) evalInspector(name funcName, args []tsvt.Expr) (Value, boolResult) {
	fn, ok := inspectors[string(name)]
	if !ok {
		return Value{}, false
	}
	if len(args) != 1 {
		return errorValue(ErrValue), true
	}
	return fn(r.eval(args[0])), true
}

// inspectors are the pure single-argument value functions behind the `IS*`,
// `N`, and `TYPE` builtins. They take an already-evaluated value, so this map
// holds no reference back into evalCall and stays a cycle-free var initializer.
var inspectors = map[string]func(v Value) Value{
	"isblank":   func(v Value) Value { return boolValue(v.kind == kindEmpty) },
	"iserror":   func(v Value) Value { return boolValue(boolResult(v.isError())) },
	"iserr":     func(v Value) Value { return boolValue(boolResult(v.isError()) && v.str != string(ErrNA)) },
	"isna":      func(v Value) Value { return boolValue(boolResult(v.isError()) && v.str == string(ErrNA)) },
	"isnumber":  func(v Value) Value { return boolValue(v.kind == kindNumber) },
	"istext":    func(v Value) Value { return boolValue(v.kind == kindString) },
	"isnontext": func(v Value) Value { return boolValue(v.kind != kindString) },
	"islogical": func(v Value) Value { return boolValue(v.kind == kindBool) },
	"iseven":    func(v Value) Value { return parityIs(v, false) },
	"isodd":     func(v Value) Value { return parityIs(v, true) },
	"n":         inspectN,
	"type":      func(v Value) Value { return numberValue(floatVal(typeCode(v))) },
}

// evalText dispatches the text builtins that must read an injected resource
// limit: REPT, whose count multiplies, and the two folds — CONCAT and TEXTJOIN
// — whose operands are flattened ranges, so a whole column is one argument.
// Each bounds its result by the byte budget rather than building it first
// (R16: a budget binds every door, and a refusal that first allocates what it
// refuses is no bound at all). ok is false for any other name.
func (r resolver) evalText(name funcName, args []tsvt.Expr) (Value, boolResult) {
	switch name {
	case "rept":
		return r.evalRept(args), true
	case "concat":
		return r.evalConcat(args), true
	case fnTextjoin:
		return r.evalTextJoin(args), true
	default:
		return Value{}, false
	}
}

// isText reports whether name is a lazily-dispatched text builtin — the set
// evalText owns. Check consults it so the checker and the evaluator agree.
func isText(name funcName) boolResult { return isAmong(name, textLazyFns) }

// textLazyFns are the byte-budgeted text builtins dispatched lazily.
// fnTextjoin names the delimited join builtin.
const fnTextjoin = "textjoin"

var textLazyFns = []funcName{"rept", "concat", fnTextjoin}

// evalConcat evaluates CONCAT(text…) — every operand's text, flattened ranges
// included, joined with no separator — under the byte budget.
func (r resolver) evalConcat(args []tsvt.Expr) Value {
	if len(args) < 1 {
		return errorValue(ErrValue)
	}
	values := r.argValues(args, cellsRest)
	if bad, found := firstError(values); found {
		return bad
	}
	return joinText(values, separator(""), skipEmpties(false), byteBudget(r.comp.limits.ResultBytes))
}

// evalTextJoin evaluates TEXTJOIN(delimiter, ignore_empty, text…) under the
// byte budget: the leading two operands are scalars, the rest flatten.
func (r resolver) evalTextJoin(args []tsvt.Expr) Value {
	if len(args) < 3 {
		return errorValue(ErrValue)
	}
	values := r.argValues(args, twoScalarsThenCells)
	if bad, found := firstError(values); found {
		return bad
	}
	isSkippingEmpties, _ := values[1].truthy()
	return joinText(values[2:], separator(values[0].String()), skipEmpties(isSkippingEmpties),
		byteBudget(r.comp.limits.ResultBytes))
}

// evalRept evaluates REPT(text, count) lazily so it can bound its result by the
// injected byte limit. Its observable behavior matches the former eager path: a
// wrong arity or a propagated argument error short-circuits, then repeatText
// applies the count and byte-budget checks.
func (r resolver) evalRept(args []tsvt.Expr) Value {
	if len(args) != 2 {
		return errorValue(ErrValue)
	}
	values := r.argValues(args, paramModes{})
	if bad, found := firstError(values); found {
		return bad
	}
	return repeatText(values, byteBudget(r.comp.limits.ResultBytes))
}
