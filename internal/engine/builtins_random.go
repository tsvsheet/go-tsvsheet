package engine

import "github.com/tsvsheet/go-tsvsheet/internal/tsvt"

// Tick is a recompute-pass ordinal, injected for tick()/frame(): a frontend
// re-rendering a volatile sheet passes an incrementing value each pass.
type Tick int

// prngSeed seeds the pass generator (derived from the pass clock).
type prngSeed uint64

// passRNG is the spec'd per-pass generator (splitmix64), defined here in full so
// the draw sequence is identical on every host (one engine, R1) and independent
// of any standard-library version. Its state is a pointer so the generator is
// shared when the computer is copied by value, and so its advancing methods stay
// value receivers.
type passRNG struct{ state *uint64 }

// newPassRNG seeds the generator.
func newPassRNG(seed prngSeed) passRNG {
	s := uint64(seed)
	return passRNG{state: &s}
}

// splitmix constants: the increment and the two multiplicative mixers.
const (
	splitmixGamma = 0x9E3779B97F4A7C15
	splitmixMulA  = 0xBF58476D1CE4E5B9
	splitmixMulB  = 0x94D049BB133111EB
)

// next advances the generator by one splitmix64 step.
func (p passRNG) next() uint64 {
	*p.state += splitmixGamma
	z := *p.state
	z = (z ^ (z >> 30)) * splitmixMulA
	z = (z ^ (z >> 27)) * splitmixMulB
	return z ^ (z >> 31)
}

// float64 returns a draw in [0, 1) from the top 53 bits.
func (p passRNG) float64() float64 {
	return float64(p.next()>>11) / float64(uint64(1)<<53)
}

// intN returns a draw in [0, n) for n > 0.
func (p passRNG) intN(n int64) int64 {
	return int64(p.next() % uint64(n))
}

// Volatility and generator builtin names. volatile is the sole marker that a
// cell should re-evaluate over time; the generators are pure per pass (they draw
// from the pass PRNG or read the pass ordinal) and animate only when wrapped in
// volatile().
const (
	fnVolatile    = "volatile"
	fnRand        = "rand"
	fnRandom      = "random"
	fnRandbetween = "randbetween"
	fnRandarray   = "randarray"
	fnTick        = "tick"
	fnFrame       = "frame"
)

// isVolatileFn reports whether name is the volatile wrapper.
func isVolatileFn(name funcName) boolResult {
	return name == fnVolatile
}

// evalVolatile evaluates volatile(expr[, schedule]): it returns expr's value
// verbatim — transparent to scalars, arrays, and errors alike — so its only
// effect is marking the cell volatile (see Sheet.VolatileSchedules). The
// optional second argument is a static cadence hint, not evaluated here.
func (r resolver) evalVolatile(name funcName, args []tsvt.Expr) (Value, boolResult) {
	if !isVolatileFn(name) {
		return Value{}, false
	}
	if len(args) < 1 || len(args) > 2 {
		return errorValue(ErrValue), true
	}
	return r.eval(args[0]), true
}

// isRandom reports whether name is one of the pass-state generator builtins.
func isRandom(name funcName) boolResult {
	switch name {
	case fnRand, fnRandom, fnRandbetween, fnRandarray, fnTick, fnFrame:
		return true
	default:
		return false
	}
}

// evalRandom dispatches the generator builtins, which read the pass PRNG or the
// pass ordinal. ok is false for any other name.
func (r resolver) evalRandom(name funcName, args []tsvt.Expr) (Value, boolResult) {
	switch name {
	case fnRand, fnRandom:
		return r.randFloat(args), true
	case fnRandbetween:
		return r.randBetween(args), true
	case fnRandarray:
		return r.randArray(args), true
	case fnTick, fnFrame:
		return tickValue(argCount(len(args)), r.comp.tick), true
	default:
		return Value{}, false
	}
}

// randFloat evaluates rand()/random(): a float in [0,1) from the pass PRNG.
func (r resolver) randFloat(args []tsvt.Expr) Value {
	if len(args) != 0 {
		return errorValue(ErrValue)
	}
	return numberValue(floatVal(r.comp.rng.float64()))
}

// randBetween evaluates randbetween(lo, hi): a uniform integer in [lo, hi].
// Bounds truncate toward zero; hi < lo is #NUM!; a non-numeric bound is #VALUE!.
func (r resolver) randBetween(args []tsvt.Expr) Value {
	if len(args) != 2 {
		return errorValue(ErrValue)
	}
	lo, bad := r.eval(args[0]).asNumber()
	if bad.isError() {
		return bad
	}
	hi, bad := r.eval(args[1]).asNumber()
	if bad.isError() {
		return bad
	}
	loI, hiI := int64(lo), int64(hi)
	if hiI < loI {
		return errorValue(ErrNum)
	}
	return numberValue(floatVal(loI + r.comp.rng.intN(hiI-loI+1)))
}

// randArray evaluates randarray(rows, [cols]): a rows×cols block of [0,1) draws
// that spills, mirroring SEQUENCE's dimension and cell-budget handling.
func (r resolver) randArray(args []tsvt.Expr) Value {
	if len(args) < 1 || len(args) > 2 {
		return errorValue(ErrValue)
	}
	rows, cols, bad := r.seqDims(args)
	if bad.isError() {
		return bad
	}
	if rows < 1 || cols < 1 {
		return errorValue(ErrValue)
	}
	if r.comp.limits.tooManyCells(resultDim(rows), resultDim(cols)) {
		return errorValue(ErrValue) // result exceeds the cell budget
	}
	return arrayValue(r.randMatrix(rows, cols))
}

// randMatrix builds a rows×cols grid of [0,1) draws from the pass PRNG.
func (r resolver) randMatrix(rows, cols charPos) [][]Value {
	m := make([][]Value, rows)
	for i := range m {
		m[i] = make([]Value, cols)
		for j := range m[i] {
			m[i][j] = numberValue(floatVal(r.comp.rng.float64()))
		}
	}
	return m
}

// tickValue evaluates tick()/frame(): the injected recompute-pass ordinal, or
// #VALUE! for any argument.
func tickValue(argc argCount, tick Tick) Value {
	if argc != 0 {
		return errorValue(ErrValue)
	}
	return numberValue(floatVal(tick))
}
