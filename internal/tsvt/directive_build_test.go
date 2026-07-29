package tsvt

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
)

// TestBuild_RejectionsCarryTheRightSentinel pins WHICH error each rejection is,
// not merely that one occurred. The two are not interchangeable: a caller
// distinguishes "you wrote something the grammar does not admit" (ErrSyntax,
// fix the spelling) from "you wrote a well-formed thing that means nothing"
// (ErrInvalidValue, fix the value), and a directive linter reports them
// differently. errors.Is is the assertion, so a rewrapping that broke the chain
// would fail here rather than pass on a matching message.
func TestBuild_RejectionsCarryTheRightSentinel(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want  error
		value ValueText
	}{
		"unknown item":        {constants.ErrInvalidValue, "rows(nope(1))"},
		"span runs backwards": {constants.ErrInvalidValue, "rows(range(9:2))"},
		"lettered row":        {constants.ErrInvalidValue, "rows(range(A:C))"},
		"numbered column":     {constants.ErrInvalidValue, "cols(range(A:9))"},
		"offset out of range": {constants.ErrInvalidValue, "rows(range(99999999999999999999:1))"},
		"item is not a call":  {constants.ErrSyntax, "rows(3)"},
		"unclosed call":       {constants.ErrSyntax, "rows(range(1:2)"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseDirectiveValue(tc.value)
			require.Error(t, err)
			assert.True(t, errors.Is(err, tc.want), "want %v, got %v", tc.want, err)
		})
	}
}

// TestBuild_AcceptsTheFormsTheHintsRecommend is the other half of the contract:
// every spelling the rejection hints above name as the fix must actually parse.
// A hint that recommends an unparseable form is worse than no hint.
func TestBuild_AcceptsTheFormsTheHintsRecommend(t *testing.T) {
	t.Parallel()

	for _, value := range []ValueText{
		"rows(range(20:31))",
		"rows(count(3))",
		"cols(range(B:M))",
		"rows(range(3:3))",
	} {
		_, err := ParseDirectiveValue(value)
		assert.NoError(t, err, string(value))
	}
}

// TestParseDirectiveValue_SyntaxSentinel names constants.ErrSyntax directly, so
// the sentinel's identity is asserted rather than inferred from a table.
func TestParseDirectiveValue_SyntaxSentinel(t *testing.T) {
	t.Parallel()

	_, err := ParseDirectiveValue("rows(3)")
	require.Error(t, err)
	assert.True(t, errors.Is(err, constants.ErrSyntax))
}

// TestParseDirectiveValue_InvalidValueSentinel is the same for the other half of
// the pair: well-formed syntax that names nothing meaningful.
func TestParseDirectiveValue_InvalidValueSentinel(t *testing.T) {
	t.Parallel()

	_, err := ParseDirectiveValue("rows(nope(1))")
	require.Error(t, err)
	assert.True(t, errors.Is(err, constants.ErrInvalidValue))
}

// TestAscending_MixedAnchorsAlwaysRunForwards pins the non-obvious half of the
// rule: two offsets compare directly only when anchored to the same end, and a
// positive start with a negative end is forwards BY CONSTRUCTION — it ends at
// or before the last line, which is at or after the start. Comparing the raw
// numbers would call `2:-1` backwards and reject a span that is perfectly
// well-formed.
func TestAscending_MixedAnchorsAlwaysRunForwards(t *testing.T) {
	t.Parallel()

	assert.True(t, ascending(2, -1), "positive start, negative end is forwards")
	assert.True(t, ascending(1, -1), "the whole axis")
	assert.True(t, ascending(2, 4), "same anchor, increasing")
	assert.True(t, ascending(-4, -2), "both from the end, increasing")
	assert.True(t, ascending(3, 3), "a single line is not backwards")
	assert.False(t, ascending(4, 2), "same anchor, decreasing")
	assert.False(t, ascending(-2, -4), "both from the end, decreasing")
}

// TestColumnOffset_MatchesTheRowAxisNumbering pins the claim that a column
// compares and shifts exactly like a row: the letters must land on the same
// 1-based scale, or an insert at column B would shift the wrong columns.
func TestColumnOffset_MatchesTheRowAxisNumbering(t *testing.T) {
	t.Parallel()

	cases := map[endpointText]Offset{
		"A": 1, "B": 2, "Z": 26,
		"AA": 27, "AB": 28, "AZ": 52, "BA": 53, "ZZ": 702,
	}
	for letters, want := range cases {
		assert.Equal(t, want, columnOffset(letters), string(letters))
	}
}
