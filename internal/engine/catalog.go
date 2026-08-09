package engine

import (
	"slices"
	"strings"
)

// The enumerable function catalog. Every builtin the engine answers to is
// listed by the same declarations its dispatch predicates read — the eager
// registry, the value inspectors, and each lazy family's name list — so the
// catalog cannot drift from what evaluates: a name added to a family is in
// both by the one edit, and TestCatalogAgreesWithIsKnownFunc holds the two
// surfaces together from the other side.

// FunctionName is one builtin's canonical (lowercase) name.
type FunctionName string

// isAmong reports membership of a name in a family's declared list.
func isAmong(name funcName, family []funcName) boolResult {
	return boolResult(slices.Contains(family, name))
}

// lazyFamilies are the declared name lists behind the lazy-dispatch
// predicates, gathered for enumeration.
func lazyFamilies() [][]funcName {
	return [][]funcName{
		conditionalFns, clockFns, volatileFns, randomFns, tableFns,
		criteriaFns, arrayFns, seriesNames(), digestFns, textLazyFns,
		embedFns, importNames(),
	}
}

// seriesNames enumerates the timeseries registry.
func seriesNames() []funcName {
	names := make([]funcName, 0, len(seriesFuncs))
	for name := range seriesFuncs {
		names = append(names, funcName(name))
	}
	return names
}

// importNames enumerates the IMPORT* registry.
func importNames() []funcName {
	names := make([]funcName, 0, len(importMedia))
	for name := range importMedia {
		names = append(names, funcName(name))
	}
	return names
}

// Functions lists every callable builtin, lowercase and sorted, with no
// duplicates. The meta function `named` is deliberately absent: it is not
// callable — it exists only in the `|@` clause (SPECIFICATION §5.6) — and a
// completion or catalog that offered it as a function would teach the
// misplacement Check refuses.
func Functions() []FunctionName {
	seen := map[funcName]boolResult{}
	for name := range functions {
		seen[funcName(name)] = true
	}
	for name := range inspectors {
		seen[funcName(name)] = true
	}
	for _, family := range lazyFamilies() {
		for _, name := range family {
			seen[funcName(strings.ToLower(string(name)))] = true
		}
	}
	names := make([]FunctionName, 0, len(seen))
	for name := range seen {
		names = append(names, FunctionName(name))
	}
	slices.Sort(names)
	return names
}
