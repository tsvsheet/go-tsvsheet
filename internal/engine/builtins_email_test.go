package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
