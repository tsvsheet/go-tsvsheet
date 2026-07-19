package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

func TestURL_Components(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`urlscheme("https://example.com/a?q=1#f")`:      "https",
		`urlhost("https://example.com:8080/a")`:         "example.com", // the port is not the host
		`urlpath("https://example.com/a/b")`:            "/a/b",
		`urlpath("https://example.com")`:                "", // no path component
		`urlfragment("https://example.com/#frag")`:      "frag",
		`urlquery("https://example.com/?q=1&r=2", "r")`: "2",
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, formula1(t, expr))
		})
	}
}

func TestURL_ComponentErrors(t *testing.T) {
	t.Parallel()

	cases := map[string]engine.ErrorValue{
		`urlscheme("/relative/only")`:                 engine.ErrValue, // not absolute
		`urlhost("http://exa mple.com/")`:             engine.ErrValue, // unparsable
		`urlquery("https://example.com/?q=1", "zz")`:  engine.ErrNA,    // missing key
		`urlquery("https://example.com/?%zz=1", "q")`: engine.ErrValue, // malformed query
		`urlquery("/relative", "q")`:                  engine.ErrValue, // not absolute
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, string(want), formula1(t, expr))
		})
	}
}

func TestURL_EncodeIsRFC3986(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`urlencode("a b/~")`: "a%20b%2F~", // space is %20, never "+"
		`urlencode("é")`:     "%C3%A9",    // per UTF-8 byte
		`urlencode("A9-._")`: "A9-._",     // the unreserved set passes through
		`urlencode(5)`:       "5",         // numbers encode their text form
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, formula1(t, expr))
		})
	}
}

func TestURL_DecodeIsPercentOnly(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "a b", formula1(t, `urldecode("a%20b")`))
	assert.Equal(t, "a+b", formula1(t, `urldecode("a+b")`)) // "+" is not a space
	assert.Equal(t, string(engine.ErrValue), formula1(t, `urldecode("%zz")`))
	assert.Equal(t, string(engine.ErrValue), formula1(t, `urldecode("%ff")`)) // not UTF-8
}
