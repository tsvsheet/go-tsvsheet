package engine

import (
	"math"
	"strconv"
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
)

// AddressText is spreadsheet-address source text (`A1`, `F4`) accepted by
// ParseAddress. It is exported so callers in other packages can convert their
// string input at the call site.
type AddressText string

// Address is a cell coordinate in spreadsheet notation (`F4`): column letters
// plus a 1-based row. It carries 0-based indices internally.
type Address struct {
	Row int `json:"row"` // 0-based
	Col int `json:"col"` // 0-based
}

// ParseAddress parses spreadsheet notation (`A1`, `F4`, `AA10`) into an
// Address. The column is one or more ASCII uppercase letters, the row a
// positive integer; anything else is constants.ErrInvalidValue.
func ParseAddress(s AddressText) (Address, error) {
	letters, digits := splitLetters(s)
	if letters == "" || digits == "" {
		return Address{}, constants.ErrInvalidValue.With(nil, "address", string(s))
	}
	row, err := strconv.Atoi(digits)
	if err != nil || row < 1 {
		return Address{}, constants.ErrInvalidValue.With(nil, "address", string(s))
	}
	col, ok := lettersToIndex(columnLetters(letters))
	if !ok {
		return Address{}, constants.ErrInvalidValue.With(nil, "address", string(s))
	}
	return Address{Row: row - 1, Col: col}, nil
}

// splitLetters splits a spreadsheet address into its leading uppercase-letter
// run and trailing digit run; a malformed shape (a non-digit in the trailing
// run) leaves the digit part empty.
func splitLetters(s AddressText) (letters, digits string) {
	i := 0
	for i < len(s) && s[i] >= 'A' && s[i] <= 'Z' {
		i++
	}
	rest := string(s[i:])
	for j := 0; j < len(rest); j++ {
		if rest[j] < '0' || rest[j] > '9' {
			return string(s[:i]), ""
		}
	}
	return string(s[:i]), rest
}

// String renders the Address in spreadsheet notation.
func (a Address) String() string {
	return indexToLetters(colIndex(a.Col)) + strconv.Itoa(a.Row+1)
}

// maxColumnMagnitude bounds the bijective base-26 value a column-letter run may
// reach: maxSafeMagnitude, tightened to the platform's int ceiling on a 32-bit
// target so the accumulator's final value always fits an int there too.
const maxColumnMagnitude = min(maxSafeMagnitude, math.MaxInt)

// lettersToIndex converts spreadsheet column letters to a 0-based index
// (A→0, Z→25, AA→26), bijective base-26. ok is false when the letter run's
// value exceeds maxColumnMagnitude: the grammar admits arbitrarily long runs,
// and an unguarded accumulation wraps into a nonsense index that downstream
// allocation math would trust — the refusal has to happen at the conversion,
// for the same reason boundedInt refuses at the float conversion. The
// accumulator is 64-bit on every platform and the bound is checked after every
// step, so it can never exceed 26·maxColumnMagnitude+26 — far below overflow.
func lettersToIndex(letters columnLetters) (int, boolResult) {
	var index int64
	for i := 0; i < len(letters); i++ {
		index = index*26 + int64(letters[i]-'A') + 1
		if index > maxColumnMagnitude {
			return 0, false
		}
	}
	return int(index - 1), true
}

// indexToLetters converts a 0-based column index to spreadsheet letters.
func indexToLetters(index colIndex) string {
	var b strings.Builder
	for n := index + 1; n > 0; n = (n - 1) / 26 {
		_ = b.WriteByte(byte('A' + (n-1)%26))
	}
	return reverse(columnLetters(b.String()))
}

// reverse returns s with its bytes reversed (ASCII column letters only).
func reverse(s columnLetters) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}
