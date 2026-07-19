package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// jsonDoc is the reference document exercising every JSON shape; it lives in
// A1 and the formula under test in B1.
const jsonDoc = `{"a":1,"b":[true,null,2.5],"c":{"d":"x"},"e":""}`

// jsonFormula computes one formula in B1 with jsonDoc in A1.
func jsonFormula(t *testing.T, expr string) string {
	t.Helper()
	return cellAt(t, compute(t, jsonDoc+"\t="+expr+"\n"), 0, 1)
}

func TestJSON_GetScalarsMapIntoTheValueModel(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`jsonget(A1, "a")`:    "1",    // number
		`jsonget(A1, "b[0]")`: "TRUE", // boolean
		`jsonget(A1, "b[1]")`: "",     // null → empty
		`jsonget(A1, "b[2]")`: "2.5",  // number in an array
		`jsonget(A1, "c.d")`:  "x",    // nested string
		`jsonget(A1, "e")`:    "",     // empty string
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, jsonFormula(t, expr))
		})
	}
}

func TestJSON_GetContainersKeepCompactText(t *testing.T) {
	t.Parallel()

	assert.Equal(t, `{"d":"x"}`, jsonFormula(t, `jsonget(A1, "c")`))
	assert.Equal(t, `[true,null,2.5]`, jsonFormula(t, `jsonget(A1, "b")`))
	// The empty path is the whole document, re-rendered compactly.
	assert.Equal(t, `{"a":1,"b":[true,null,2.5],"c":{"d":"x"},"e":""}`, jsonFormula(t, `jsonget(A1, "")`))

	// Both boolean literals survive a container round-trip.
	bools := compute(t, "[true,false]\t=jsonget(A1, \"\")\n")
	assert.Equal(t, "[true,false]", cellAt(t, bools, 0, 1))
}

func TestJSON_NumberLiteralsSurviveRoundTrips(t *testing.T) {
	t.Parallel()

	// 1.50 must not normalize to 1.5 in a text round-trip.
	g := compute(t, `{"n":1.50}`+"\t=jsonget(A1, \"\")\n")
	assert.Equal(t, `{"n":1.50}`, cellAt(t, g, 0, 1))
}

func TestJSON_StringEscapesRoundTrip(t *testing.T) {
	t.Parallel()

	g := compute(t, `{"s":"a\"b"}`+"\t=jsonget(A1, \"\")\n")
	assert.Equal(t, `{"s":"a\"b"}`, cellAt(t, g, 0, 1))
}

func TestJSON_RootArrayIndexing(t *testing.T) {
	t.Parallel()

	g := compute(t, "[10,20]\t=jsonget(A1, \"[1]\")\n")
	assert.Equal(t, "20", cellAt(t, g, 0, 1))
}

func TestJSON_NumberOverflowIsNum(t *testing.T) {
	t.Parallel()

	g := compute(t, "1e999\t=jsonget(A1, \"\")\n")
	assert.Equal(t, string(engine.ErrNum), cellAt(t, g, 0, 1))
}

func TestJSON_GetMissesAreNA(t *testing.T) {
	t.Parallel()

	cases := []string{
		`jsonget(A1, "z")`,    // missing key
		`jsonget(A1, "b[9]")`, // index out of range
		`jsonget(A1, "a.b")`,  // stepping into a scalar
		`jsonget(A1, "e[0]")`, // indexing a string
	}
	for _, expr := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, string(engine.ErrNA), jsonFormula(t, expr))
		})
	}
}

func TestJSON_MalformedDocumentsAreValueErrors(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"open object":     "{",
		"open array":      "[1,",
		"nested truncate": `[{"a":1`,
		"bad member":      `{"a":}`,
		"non-string key":  "{1:2}",
		"trailing input":  "1 2",
		"empty document":  "",
	}
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := compute(t, doc+"\t=jsontype(A1)\n")
			assert.Equal(t, string(engine.ErrValue), cellAt(t, g, 0, 1))
		})
	}

	// Every reader rejects a malformed document the same way.
	g := compute(t, "{\t=jsonlen(A1)\t=jsonkeys(A1)\n")
	assert.Equal(t, string(engine.ErrValue), cellAt(t, g, 0, 1))
	assert.Equal(t, string(engine.ErrValue), cellAt(t, g, 0, 2))
}

func TestJSON_MalformedPathsAreValueErrors(t *testing.T) {
	t.Parallel()

	cases := []string{"a..b", "a[", "a[x]", "a[-1]", ".a", "a.", "a[0]x", "a.[0]"}
	for _, path := range cases {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, string(engine.ErrValue), jsonFormula(t, `jsonget(A1, "`+path+`")`))
		})
	}
}

func TestJSON_TypeNamesEveryShape(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`jsontype(A1, "b[1]")`: "null",
		`jsontype(A1, "b[0]")`: "boolean",
		`jsontype(A1, "a")`:    "number",
		`jsontype(A1, "c.d")`:  "string",
		`jsontype(A1, "b")`:    "array",
		`jsontype(A1, "c")`:    "object",
		`jsontype(A1)`:         "object", // no path → the root
		`jsontype(A1, "zz")`:   string(engine.ErrNA),
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, jsonFormula(t, expr))
		})
	}
}

func TestJSON_LenCountsContainers(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`jsonlen(A1, "b")`:  "3",
		`jsonlen(A1)`:       "4",                     // the root object's member count
		`jsonlen(A1, "a")`:  string(engine.ErrValue), // scalar
		`jsonlen(A1, "zz")`: string(engine.ErrNA),
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, jsonFormula(t, expr))
		})
	}
}

func TestJSON_KeysSpillInDocumentOrder(t *testing.T) {
	t.Parallel()

	g := compute(t, jsonDoc+"\t=jsonkeys(A1)\n")
	assert.Equal(t, "a", cellAt(t, g, 0, 1))
	assert.Equal(t, "b", cellAt(t, g, 1, 1))
	assert.Equal(t, "c", cellAt(t, g, 2, 1))
	assert.Equal(t, "e", cellAt(t, g, 3, 1))

	nested := compute(t, jsonDoc+"\t=jsonkeys(A1, \"c\")\n")
	assert.Equal(t, "d", cellAt(t, nested, 0, 1))
}

func TestJSON_KeysErrors(t *testing.T) {
	t.Parallel()

	empty := compute(t, "{}\t=jsonkeys(A1)\n")
	assert.Equal(t, string(engine.ErrNA), cellAt(t, empty, 0, 1)) // no rows — FILTER's convention

	assert.Equal(t, string(engine.ErrValue), jsonFormula(t, `jsonkeys(A1, "b")`)) // non-object
	assert.Equal(t, string(engine.ErrNA), jsonFormula(t, `jsonkeys(A1, "zz")`))   // missing
}

func TestJSON_SetPreservesOrderAndAppends(t *testing.T) {
	t.Parallel()

	g := compute(t, `{"a":1,"b":2}`+"\t=jsonset(A1, \"a\", 5)\t=jsonset(A1, \"z\", \"w\")\n")
	assert.Equal(t, `{"a":5,"b":2}`, cellAt(t, g, 0, 1))
	assert.Equal(t, `{"a":1,"b":2,"z":"w"}`, cellAt(t, g, 0, 2))
}

func TestJSON_SetMaterializesMissingObjects(t *testing.T) {
	t.Parallel()

	g := compute(t, "{}\t=jsonset(A1, \"a.b\", 1)\n")
	assert.Equal(t, `{"a":{"b":1}}`, cellAt(t, g, 0, 1))
}

func TestJSON_SetWritesArrayElements(t *testing.T) {
	t.Parallel()

	g := compute(t, "[1,2]\t=jsonset(A1, \"[1]\", 9)\n")
	assert.Equal(t, "[1,9]", cellAt(t, g, 0, 1))
}

func TestJSON_SetValueKinds(t *testing.T) {
	t.Parallel()

	// Boolean, empty (null), date (canonical ISO text), and text — with no
	// HTML escaping of the text. B2 is an in-grid empty cell.
	g := compute(t, "{}\t=jsonset(A1, \"k\", TRUE)\t=jsonset(A1, \"k\", B2)\n\t\n")
	assert.Equal(t, `{"k":true}`, cellAt(t, g, 0, 1))
	assert.Equal(t, `{"k":null}`, cellAt(t, g, 0, 2))

	d := compute(t, "{}\t=jsonset(A1, \"k\", date(2026, 1, 19))\t=jsonset(A1, \"k\", \"<b>\")\n")
	assert.Equal(t, `{"k":"2026-01-19"}`, cellAt(t, d, 0, 1))
	assert.Equal(t, `{"k":"<b>"}`, cellAt(t, d, 0, 2))
}

func TestJSON_SetReplacesTheRoot(t *testing.T) {
	t.Parallel()

	g := compute(t, `{"a":1}`+"\t=jsonset(A1, \"\", 7)\n")
	assert.Equal(t, "7", cellAt(t, g, 0, 1))
}

func TestJSON_SetErrors(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		doc  string
		expr string
		want engine.ErrorValue
	}{
		"index past the end":      {"[1]", `jsonset(A1, "[5]", 0)`, engine.ErrNA},
		"index at length":         {"[1]", `jsonset(A1, "[1]", 0)`, engine.ErrNA},
		"index into an object":    {`{"a":1}`, `jsonset(A1, "[0]", 0)`, engine.ErrNA},
		"key into a scalar":       {`{"a":1}`, `jsonset(A1, "a.b", 0)`, engine.ErrNA},
		"key into an array":       {`[[1]]`, `jsonset(A1, "[0].a", 0)`, engine.ErrNA},
		"deep index out of range": {`{"a":[1]}`, `jsonset(A1, "a[5]", 0)`, engine.ErrNA},
		"malformed document":      {"{", `jsonset(A1, "a", 0)`, engine.ErrValue},
		"malformed path":          {`{"a":1}`, `jsonset(A1, "a..b", 0)`, engine.ErrValue},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := compute(t, tc.doc+"\t="+tc.expr+"\n")
			assert.Equal(t, string(tc.want), cellAt(t, g, 0, 1))
		})
	}
}
