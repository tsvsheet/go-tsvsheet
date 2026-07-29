// Package engine's regex family: the three REGEX* builtins over Go's RE2
// syntax, kept apart from the plain text builtins because a bad pattern is a
// value error rather than a text operation.
package engine

import "regexp"

// fnRegexMatch reports whether text matches a regular expression.
func fnRegexMatch(args []Value) Value {
	re, bad := compileRegex(textVal(argText(args, 1)))
	if bad.isError() {
		return bad
	}
	return boolValue(boolResult(re.MatchString(argText(args, 0))))
}

// fnRegexExtract returns the first match of a regular expression; no match is
// #N/A.
func fnRegexExtract(args []Value) Value {
	re, bad := compileRegex(textVal(argText(args, 1)))
	if bad.isError() {
		return bad
	}
	subject := argText(args, 0)
	if !re.MatchString(subject) {
		return errorValue(ErrNA)
	}
	return stringValue(textVal(re.FindString(subject)))
}

// fnRegexReplace replaces every match of a regular expression.
func fnRegexReplace(args []Value) Value {
	re, bad := compileRegex(textVal(argText(args, 1)))
	if bad.isError() {
		return bad
	}
	return stringValue(textVal(re.ReplaceAllString(argText(args, 0), argText(args, 2))))
}

// compileRegex compiles a pattern, reporting an invalid pattern as #VALUE!.
func compileRegex(pattern textVal) (*regexp.Regexp, Value) {
	re, err := regexp.Compile(string(pattern))
	if err != nil {
		return nil, errorValue(ErrValue)
	}
	return re, Value{}
}
