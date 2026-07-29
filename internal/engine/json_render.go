// Package engine's JSON renderer: a jsonNode tree back to compact JSON text,
// preserving member order and number literals exactly as they were read.
package engine

import (
	"encoding/json"
	"strings"
)

// renderJSON re-encodes a node compactly, preserving member order and number
// literals verbatim.
func renderJSON(node jsonNode) string {
	switch node.kind {
	case jsonNull:
		return "null"
	case jsonBool:
		return boolLiteral(boolResult(node.isTrue))
	case jsonNumber:
		return node.str
	case jsonString:
		return jsonQuote(textVal(node.str))
	case jsonArray:
		return renderArray(node.arr)
	default: // jsonObject
		return renderObject(node.members)
	}
}

// boolLiteral is the JSON boolean literal.
func boolLiteral(isTrue boolResult) string {
	if isTrue {
		return "true"
	}
	return "false"
}

// jsonQuote encodes s as a JSON string literal without HTML escaping (JSONSET
// output is a document, not markup — "<" stays "<").
func jsonQuote(s textVal) string {
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(string(s)) // encoding a plain string never fails
	return strings.TrimSuffix(b.String(), "\n")
}

// renderArray re-encodes an array's elements.
func renderArray(items []jsonNode) string {
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = renderJSON(item)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// renderObject re-encodes an object's members in document order.
func renderObject(members []jsonMember) string {
	parts := make([]string, len(members))
	for i, m := range members {
		parts[i] = jsonQuote(textVal(m.key)) + ":" + renderJSON(m.value)
	}
	return "{" + strings.Join(parts, ",") + "}"
}
