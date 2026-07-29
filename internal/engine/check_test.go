package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// parse is a test helper that parses a sheet, failing on error.
func parse(t *testing.T, src string) engine.Sheet {
	t.Helper()
	s, err := engine.Parse([]byte(src))
	require.NoError(t, err)
	return s
}

func TestCheck_Clean(t *testing.T) {
	t.Parallel()

	assert.Empty(t, engine.Check(parse(t, "1\t2\t=A1 + B1\n")))
}

func TestCheck_UnknownFunction(t *testing.T) {
	t.Parallel()

	diags := engine.Check(parse(t, "1\t=bogus(A1)\n"))
	require.Len(t, diags, 1)
	assert.Equal(t, "B1", diags[0].Cell)
	assert.Contains(t, diags[0].Message, "bogus")
	assert.False(t, diags[0].IsFatal)
}

func TestCheck_NumberFormulaHasNoRefs(t *testing.T) {
	t.Parallel()

	// A formula with no calls yields no diagnostics (the walker no-ops).
	assert.Empty(t, engine.Check(parse(t, "=1 + 2\n")))
}

func TestCheck_KnownFunctionsClean(t *testing.T) {
	t.Parallel()

	// A conditional (`if`), an inspector (`isnumber`), a table function
	// (`index`), a criteria function (`countif`), an array function (`unique`),
	// and an eager function (`sum`) are all known — no diagnostics.
	assert.Empty(
		t,
		engine.Check(parse(t, "1\t2\t=if(isnumber(A1), countif(unique(A1:B1), 1), index(A1:B1, 1, 1))\n")),
	)
}

func TestCheck_TextFunctionsKnown(t *testing.T) {
	t.Parallel()

	// The lazily-dispatched text builtins (REPT, bounded by the byte budget at
	// compute time) are known — Check must not flag what the evaluator computes.
	assert.Empty(t, engine.Check(parse(t, "3\t=rept(\"█\", A1)\n")))
}

func TestCheck_ImportFunctionsKnown(t *testing.T) {
	t.Parallel()

	// The lazily-dispatched IMPORT* functions are known builtins — Check must not
	// report them as unknown functions (they resolve to #IMPORT! at compute time
	// only when no fetcher is injected, which is a value, not a static error).
	for _, fn := range []string{"importcell", "importrow", "importcolumn", "importrange", "importsheet"} {
		t.Run(fn, func(t *testing.T) {
			t.Parallel()
			assert.Empty(t, engine.Check(parse(t, "="+fn+`("https://x/v")`+"\n")))
		})
	}
}

func TestCheck_WalksIntoUnaryPercentBinaryAndCall(t *testing.T) {
	t.Parallel()

	// Each wrapper form must be walked to reach the unknown call inside it.
	for _, src := range []string{"=-bogus(A1)", "=bogus(A1)%", "=bogus(A1) + 1", "=abs(bogus(A1))"} {
		t.Run(src, func(t *testing.T) {
			t.Parallel()
			diags := engine.Check(parse(t, "1\t"+src+"\n"))
			require.Len(t, diags, 1)
			assert.Contains(t, diags[0].Message, "bogus")
		})
	}
}
