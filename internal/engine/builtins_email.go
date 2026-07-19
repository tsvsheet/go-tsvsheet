package engine

import (
	"net/mail"
	"strings"
)

// parsedEmail parses a bare RFC 5322 address: valid iff it parses and the
// parsed address round-trips the input exactly, which rejects display-name
// and angle-bracket forms. ok is false otherwise. Validity is syntactic only —
// never DNS, which would be I/O (ADR 0011 §4).
func parsedEmail(text textVal) (string, boolResult) {
	addr, err := mail.ParseAddress(string(text))
	if err != nil || addr.Address != string(text) {
		return "", false
	}
	return addr.Address, true
}

// fnEmailValid is TRUE iff the argument is a syntactically valid bare address.
func fnEmailValid(args []Value) Value {
	_, ok := parsedEmail(textVal(argText(args, 0)))
	return boolValue(ok)
}

// emailPart lifts a side of the address split — the local part (user) or the
// domain, split at the last "@" — into an eager impl; an invalid address is
// #VALUE!.
func emailPart(isDomain boolResult) func(args []Value) Value {
	return func(args []Value) Value {
		addr, ok := parsedEmail(textVal(argText(args, 0)))
		if !ok {
			return errorValue(ErrValue)
		}
		at := strings.LastIndex(addr, "@")
		if isDomain {
			return stringValue(textVal(addr[at+1:]))
		}
		return stringValue(textVal(addr[:at]))
	}
}
