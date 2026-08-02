package engine_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

func TestParseAddress(t *testing.T) {
	t.Parallel()

	cases := map[string]engine.Address{
		"A1":   {Row: 0, Col: 0},
		"B2":   {Row: 1, Col: 1},
		"Z1":   {Row: 0, Col: 25},
		"AA1":  {Row: 0, Col: 26},
		"F4":   {Row: 3, Col: 5},
		"AB10": {Row: 9, Col: 27},
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			t.Parallel()
			got, err := engine.ParseAddress(engine.AddressText(src))
			require.NoError(t, err)
			assert.Equal(t, want, got)
		})
	}
}

func TestParseAddress_Invalid(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"no letters":    "1",
		"no digits":     "A",
		"row zero":      "A0",
		"trailing junk": "A1x",
		"empty":         "",
		"lowercase":     "a1",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := engine.ParseAddress(engine.AddressText(src))
			require.Error(t, err)
			assert.ErrorIs(t, err, constants.ErrInvalidValue)
		})
	}
}

// TestParseAddress_ColumnOverflow pins that a column-letter run long enough to
// overflow bijective base-26 accumulation is refused at the conversion rather
// than wrapping into a nonsense index that downstream allocation math trusts.
func TestParseAddress_ColumnOverflow(t *testing.T) {
	t.Parallel()

	for _, letters := range []string{
		strings.Repeat("Z", 9),  // first all-Z width past the safe bound (≈5.6e12 > 2^40)
		strings.Repeat("Z", 14), // past int64 — accumulation would wrap without the guard
		strings.Repeat("A", 64), // wraps through positive and negative alike
	} {
		t.Run(letters, func(t *testing.T) {
			t.Parallel()
			_, err := engine.ParseAddress(engine.AddressText(letters + "1"))
			require.Error(t, err)
			assert.ErrorIs(t, err, constants.ErrInvalidValue)
		})
	}
}

// TestMaxColumnMagnitudeIsTheLastAddressableColumn sits on the boundary
// itself: EFWFSWHTP is the run whose accumulator is exactly 2^40 — the 64-bit
// maxColumnMagnitude — so it is the LAST addressable column (index 2^40−1) and
// EFWFSWHTQ, one past it, is the first refused. An off-by-one in the guard's
// comparison moves exactly this line and nothing the width-based cases above
// can see. On a 32-bit platform maxColumnMagnitude always tightens to the int
// ceiling instead, so the accepting half is 64-bit-only; the refusing half
// holds everywhere.
func TestMaxColumnMagnitudeIsTheLastAddressableColumn(t *testing.T) {
	t.Parallel()

	if strconv.IntSize == 64 {
		got, err := engine.ParseAddress(engine.AddressText("EFWFSWHTP1"))
		require.NoError(t, err)
		assert.Equal(t, engine.Address{Row: 0, Col: 1<<40 - 1}, got)
	}
	_, err := engine.ParseAddress(engine.AddressText("EFWFSWHTQ1"))
	assert.ErrorIs(t, err, constants.ErrInvalidValue)
}

func TestAddress_String(t *testing.T) {
	t.Parallel()

	cases := map[string]engine.Address{
		"A1":   {Row: 0, Col: 0},
		"Z1":   {Row: 0, Col: 25},
		"AA1":  {Row: 0, Col: 26},
		"F4":   {Row: 3, Col: 5},
		"AB10": {Row: 9, Col: 27},
	}
	for want, addr := range cases {
		t.Run(want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, addr.String())
		})
	}
}

func TestAddress_RoundTrip(t *testing.T) {
	t.Parallel()

	for _, s := range []string{"A1", "Z26", "AA100", "ZZ1", "AAA5"} {
		addr, err := engine.ParseAddress(engine.AddressText(s))
		require.NoError(t, err)
		assert.Equal(t, s, addr.String())
	}
}

// TestLettersToIndexRefusesARunItsAccumulatorCannotCarry pins that a column-letter run past the
// addressable bound is #REF! in every reference shape — the grammar admits the
// syntax, and before the lettersToIndex guard the wrapped index fed allocation
// math (the range shapes here allocated from a nonsense rectangle).
func TestLettersToIndexRefusesARunItsAccumulatorCannotCarry(t *testing.T) {
	t.Parallel()

	huge := strings.Repeat(
		"Z",
		9,
	) // first all-Z width past the safe bound
	assert.Equal(t, string(engine.ErrRef), formula1(t, huge+"1"))                   // single cell
	assert.Equal(t, string(engine.ErrRef), formula1(t, "sum(A1:"+huge+"1)"))        // range, cells path
	assert.Equal(t, string(engine.ErrRef), formula1(t, "sum("+huge+"1:"+huge+"2)")) // both endpoints
	assert.Equal(t, string(engine.ErrRef), formula1(t, "index(A1:"+huge+"1, 1)"))   // range, matrix path
}
