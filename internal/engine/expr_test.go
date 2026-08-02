package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// TestFormatValueMatchesWhatWriteTSVEmitsByteForByte pins the single rendering
// rule. Two renderers for the same value is one renderer too many: the moment
// they disagree, a value shown in one surface and written to a file in another
// stop being the same value, and the difference surfaces as a diff nobody made.
func TestFormatValueMatchesWhatWriteTSVEmitsByteForByte(t *testing.T) {
	t.Parallel()
	for _, formula := range []string{
		"=1/3", "=0.1+0.2", "=2^60", "=TRUE", "=1/0", "=\"text\"", "=round(-0.4)", "=1=1",
	} {
		sheet, err := engine.Parse([]byte(formula + "\n"))
		require.NoError(t, err)

		written := sheet.Compute()[0][0]
		expr, err := engine.CompileExpr([]byte(formula[1:]))
		require.NoError(t, err)
		formatted := engine.FormatValue(expr.Eval(engine.Grid{{""}}, engine.ComputeOptions{}))

		assert.Equal(t, written, formatted, "%s renders alike wherever it is rendered", formula)
	}
}
