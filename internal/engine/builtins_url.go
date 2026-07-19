package engine

import (
	"net/url"
	"unicode/utf8"
)

// parsedURL parses an absolute URL; unparsable or scheme-less text is #VALUE!
// (ADR 0011 §4: component extraction requires an absolute URL).
func parsedURL(text textVal) (*url.URL, Value) {
	u, err := url.Parse(string(text))
	if err != nil || !u.IsAbs() {
		return nil, errorValue(ErrValue)
	}
	return u, Value{}
}

// urlPart lifts a component extractor into an eager impl over an absolute URL.
func urlPart(part func(u *url.URL) string) func(args []Value) Value {
	return func(args []Value) Value {
		u, bad := parsedURL(textVal(argText(args, 0)))
		if bad.isError() {
			return bad
		}
		return stringValue(textVal(part(u)))
	}
}

// The URL components: scheme, host (without any port), decoded path, and
// decoded fragment.
func schemeOf(u *url.URL) string   { return u.Scheme }
func hostOf(u *url.URL) string     { return u.Hostname() }
func pathOf(u *url.URL) string     { return u.Path }
func fragmentOf(u *url.URL) string { return u.Fragment }

// fnURLQuery is the first value of a query parameter; a missing key is #N/A
// and a malformed query string is #VALUE!.
func fnURLQuery(args []Value) Value {
	u, bad := parsedURL(textVal(argText(args, 0)))
	if bad.isError() {
		return bad
	}
	query, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return errorValue(ErrValue)
	}
	values, present := query[argText(args, 1)]
	if !present {
		return errorValue(ErrNA)
	}
	return stringValue(textVal(values[0]))
}

// fnURLEncode percent-encodes every byte outside RFC 3986's unreserved set
// (ALPHA / DIGIT / "-" / "." / "_" / "~") — a space is %20, never "+" —
// matching Sheets' ENCODEURL (ADR 0011 §4).
func fnURLEncode(args []Value) Value {
	text := argText(args, 0)
	out := make([]byte, 0, len(text))
	for i := 0; i < len(text); i++ {
		out = appendEncoded(out, octet(text[i]))
	}
	return stringValue(textVal(string(out)))
}

// upperHex is the percent-encoding digit alphabet (RFC 3986 uses uppercase).
const upperHex = "0123456789ABCDEF"

// octet is one byte of text under percent-encoding.
type octet byte

// appendEncoded appends b verbatim when unreserved, else percent-encoded.
func appendEncoded(out []byte, b octet) []byte {
	if unreservedByte(b) {
		return append(out, byte(b))
	}
	return append(out, '%', upperHex[b>>4], upperHex[b&0xF])
}

// unreservedByte reports whether b is in RFC 3986's unreserved set.
func unreservedByte(b octet) bool {
	switch {
	case 'A' <= b && b <= 'Z', 'a' <= b && b <= 'z', '0' <= b && b <= '9':
		return true
	case b == '-' || b == '.' || b == '_' || b == '~':
		return true
	default:
		return false
	}
}

// fnURLDecode reverses percent-encoding: %XX sequences only — a "+" stays a
// "+" — and a malformed sequence or a non-UTF-8 result is #VALUE!.
func fnURLDecode(args []Value) Value {
	text, err := url.PathUnescape(argText(args, 0))
	if err != nil || !utf8.ValidString(text) {
		return errorValue(ErrValue)
	}
	return stringValue(textVal(text))
}
