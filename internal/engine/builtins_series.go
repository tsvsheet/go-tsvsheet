package engine

import (
	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// windowSpan is a windowed series function's window size in elements.
type windowSpan int

// seriesSpec describes one windowed timeseries builtin (ADR 0011 §4): how it
// builds the output series, and whether it takes a window-size argument.
type seriesSpec struct {
	build     func(nums []float64, span windowSpan) []Value
	isSpanned boolResult
}

// seriesFuncs is the windowed timeseries registry: each reads a range as a
// dense numeric series and produces an N×1 column that spills.
var seriesFuncs = map[string]seriesSpec{
	"movingavg":  {build: windowed(meanOf), isSpanned: true},
	"ema":        {build: emaSeries, isSpanned: true},
	"rollingmin": {build: windowed(least), isSpanned: true},
	"rollingmax": {build: windowed(greatest), isSpanned: true},
	"cumsum":     {build: cumsumSeries},
}

// isSeries reports whether name is a windowed timeseries builtin.
func isSeries(name funcName) boolResult {
	_, ok := seriesFuncs[string(name)]
	return boolResult(ok)
}

// evalSeries dispatches the windowed timeseries builtins, which read their
// range argument's cells positionally and spill the derived series. ok is
// false for any other name.
func (r resolver) evalSeries(name funcName, args []tsvt.Expr) (Value, boolResult) {
	fn, owned := seriesFuncs[string(name)]
	if !owned {
		return Value{}, false
	}
	return r.seriesCall(fn, args), true
}

// seriesCall evaluates one timeseries builtin: the range flattens row-major
// into a dense numeric series, and the result is a same-length column.
func (r resolver) seriesCall(fn seriesSpec, args []tsvt.Expr) Value {
	if wrongSeriesArity(fn, argCount(len(args))) {
		return errorValue(ErrValue)
	}
	span := windowSpan(1)
	if fn.isSpanned {
		s, bad := r.spanArg(args[1])
		if bad.isError() {
			return bad
		}
		span = s
	}
	nums, bad := seriesNumbers(flatten1D(r.argMatrix(args[0])))
	if bad.isError() {
		return bad
	}
	return arrayValue(seriesColumn(fn.build(nums, span)))
}

// wrongSeriesArity reports whether n mismatches the builtin's arity: the range
// plus, for the windowed forms, the window size.
func wrongSeriesArity(fn seriesSpec, n argCount) bool {
	if fn.isSpanned {
		return n != 2
	}
	return n != 1
}

// spanArg reads the window-size argument: numeric, truncated to an integer; a
// span below 1 is #NUM!.
func (r resolver) spanArg(arg tsvt.Expr) (windowSpan, Value) {
	n, bad := intArg(r.argScalar(arg))
	if bad.isError() {
		return 0, bad
	}
	if n < 1 {
		return 0, errorValue(ErrNum)
	}
	return windowSpan(n), Value{}
}

// seriesNumbers reads flattened range cells as a dense numeric series: unlike
// the aggregate family (which skips empties), positions are meaningful in a
// window, so an empty or non-numeric element is #VALUE!; an error element
// propagates (ADR 0011 §4).
func seriesNumbers(cells []Value) ([]float64, Value) {
	nums := make([]float64, len(cells))
	for i, cell := range cells {
		if cell.kind == kindEmpty {
			return nil, errorValue(ErrValue)
		}
		n, bad := cell.asNumber()
		if bad.isError() {
			return nil, bad
		}
		nums[i] = n
	}
	return nums, Value{}
}

// seriesColumn shapes a series as the N×1 array that spills down a column.
func seriesColumn(series []Value) [][]Value {
	rows := make([][]Value, len(series))
	for i, v := range series {
		rows[i] = []Value{v}
	}
	return rows
}

// windowed lifts a fold over each trailing window of span elements; a position
// with fewer than span values is #N/A — honest windows, no partials (ADR 0011
// §4).
func windowed(fold func(window []float64) float64) func([]float64, windowSpan) []Value {
	return func(nums []float64, span windowSpan) []Value {
		out := make([]Value, len(nums))
		for i := range nums {
			if i+1 < int(span) {
				out[i] = errorValue(ErrNA)
				continue
			}
			out[i] = numberValue(floatVal(fold(nums[i+1-int(span) : i+1])))
		}
		return out
	}
}

// least is the minimum of a non-empty window.
func least(window []float64) float64 {
	m := window[0]
	for _, n := range window[1:] {
		m = min(m, n)
	}
	return m
}

// greatest is the maximum of a non-empty window.
func greatest(window []float64) float64 {
	m := window[0]
	for _, n := range window[1:] {
		m = max(m, n)
	}
	return m
}

// emaSeries is the exponential moving average with smoothing α = 2/(span+1),
// seeded with the first element and defined at every position (ADR 0011 §4).
func emaSeries(nums []float64, span windowSpan) []Value {
	alpha := 2 / (float64(span) + 1)
	out := make([]Value, len(nums))
	prev := nums[0]
	for i, n := range nums {
		if i > 0 {
			prev = alpha*n + (1-alpha)*prev
		}
		out[i] = numberValue(floatVal(prev))
	}
	return out
}

// cumsumSeries is the running total; the span is unused (CUMSUM is unspanned).
func cumsumSeries(nums []float64, _ windowSpan) []Value {
	out := make([]Value, len(nums))
	total := 0.0
	for i, n := range nums {
		total += n
		out[i] = numberValue(floatVal(total))
	}
	return out
}
