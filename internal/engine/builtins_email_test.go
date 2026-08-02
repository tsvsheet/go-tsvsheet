package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

func TestEmail_ValidIsSyntacticOnly(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`emailvalid("a@b.com")`:                    "TRUE",
		`emailvalid("first.last@sub.example.org")`: "TRUE",
		`emailvalid("Name <a@b.com>")`:             "FALSE", // bare addresses only
		`emailvalid("a@")`:                         "FALSE",
		`emailvalid("@b.com")`:                     "FALSE",
		`emailvalid("not-an-email")`:               "FALSE",
		`emailvalid("")`:                           "FALSE",
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, formula1(t, expr))
		})
	}
}

func TestEmail_PartsSplitAtTheAt(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "first.last", formula1(t, `emailuser("first.last@sub.example.org")`))
	assert.Equal(t, "sub.example.org", formula1(t, `emaildomain("first.last@sub.example.org")`))
	assert.Equal(t, string(engine.ErrValue), formula1(t, `emailuser("nope")`))
	assert.Equal(t, string(engine.ErrValue), formula1(t, `emaildomain("nope")`))
}

// TestParsedEmailAcceptsOnlyAnAddressThatRoundTripsExactly pins what "valid"
// means here: the address must parse AND come back byte-identical, which is
// what rejects the display-name and angle-bracket forms a mail client accepts
// but a cell should not silently rewrite.
func TestParsedEmailAcceptsOnlyAnAddressThatRoundTripsExactly(t *testing.T) {
	t.Parallel()
	source := "a@b.com\t=emailvalid(A1)\t=emailvalid(\"Ann <a@b.com>\")\t=emailvalid(\"<a@b.com>\")\n"
	sheet, err := engine.Parse([]byte(source))
	require.NoError(t, err)
	computed := sheet.Compute()[0]

	assert.Equal(t, "TRUE", computed[1], "a bare address round-trips")
	assert.Equal(t, "FALSE", computed[2], "a display-name form does not")
	assert.Equal(t, "FALSE", computed[3], "nor an angle-bracket form")
}
