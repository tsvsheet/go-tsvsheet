package engine

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFunctionsIsSortedUniqueLowercase states the catalog's shape contract:
// sorted, duplicate-free, and canonical-lowercase throughout, so a consumer
// can binary-search it and render it without normalizing.
func TestFunctionsIsSortedUniqueLowercase(t *testing.T) {
	t.Parallel()

	fns := Functions()
	require.NotEmpty(t, fns)
	assert.True(t, slices.IsSorted(fns))
	assert.Len(t, slices.Compact(slices.Clone(fns)), len(fns), "no duplicates")
	for _, fn := range fns {
		assert.Equal(t, FunctionName(funcName(fn)), fn)
		assert.NotContains(t, string(fn), " ")
	}
}

// TestFunctionsAgreesWithIsKnownFunc holds the catalog and the dispatch
// answer together from both sides reachable by a test: every listed name is
// known, and the family lists the catalog enumerates are the very slices the
// predicates read — so a name added to a family cannot appear in one surface
// and not the other.
func TestFunctionsAgreesWithIsKnownFunc(t *testing.T) {
	t.Parallel()

	for _, fn := range Functions() {
		assert.True(t, bool(isKnownFunc(funcName(fn))), "catalog lists %q; dispatch must know it", fn)
	}
}

// TestFunctionsCarriesEveryFamily spot-pins one member of each contributing
// surface — the eager registry, the inspectors, and every lazy family — so a
// family silently dropped from the enumeration is a named failure.
func TestFunctionsCarriesEveryFamily(t *testing.T) {
	t.Parallel()

	fns := Functions()
	for _, want := range []FunctionName{
		"sum",        // eager registry
		"isblank",    // inspectors
		"if",         // conditionals
		"now",        // clocks
		"volatile",   // the volatility wrapper
		"rand",       // draws
		"vlookup",    // table
		"countif",    // criteria
		"sort",       // arrays
		"movingavg",  // timeseries
		"digest",     // crypto ranges
		"textjoin",   // lazy text
		"sheet",      // embedding
		"importcell", // imports
	} {
		assert.Contains(t, fns, want)
	}
}

// TestFunctionsExcludesTheMetaFunction states the deliberate absence: `named`
// is not callable — it exists only in the |@ clause — and a catalog offering
// it as a function would teach the misplacement Check refuses.
func TestFunctionsExcludesTheMetaFunction(t *testing.T) {
	t.Parallel()

	assert.NotContains(t, Functions(), FunctionName(metaFnNamed))
}
