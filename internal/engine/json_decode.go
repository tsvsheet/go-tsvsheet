// Package engine's JSON decoder: a document text becomes the immutable jsonNode
// tree the JSON family navigates. Member order is preserved because jsonset
// must round-trip a document without reordering it.
package engine

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

// parseJSON parses one complete JSON document, preserving member order and
// number literals; ok is false for malformed or trailing input.
func parseJSON(text textVal) (jsonNode, bool) {
	dec := json.NewDecoder(strings.NewReader(string(text)))
	dec.UseNumber()
	node, ok := decodeNode(dec)
	if !ok {
		return jsonNode{}, false
	}
	_, err := dec.Token()
	return node, errors.Is(err, io.EOF)
}

// decodeNode decodes the next value from the token stream.
func decodeNode(dec *json.Decoder) (jsonNode, bool) {
	tok, err := dec.Token()
	if err != nil {
		return jsonNode{}, false
	}
	return nodeFromToken(dec, tok)
}

// nodeFromToken lifts one token into a node, descending into containers.
func nodeFromToken(dec *json.Decoder, tok json.Token) (jsonNode, bool) {
	switch v := tok.(type) {
	case json.Delim:
		return decodeContainer(dec, v)
	case string:
		return jsonNode{kind: jsonString, str: v}, true
	case json.Number:
		return jsonNode{kind: jsonNumber, str: v.String()}, true
	case bool:
		return jsonNode{kind: jsonBool, isTrue: v}, true
	default: // nil — JSON null
		return jsonNode{kind: jsonNull}, true
	}
}

// decodeContainer parses the container opened by delim; the decoder yields
// only '[' or '{' at a value position (a closing delimiter there is a syntax
// error it reports itself).
func decodeContainer(dec *json.Decoder, open json.Delim) (jsonNode, bool) {
	if open == '[' {
		return decodeArray(dec)
	}
	return decodeObject(dec)
}

// decodeArray parses the elements and closing bracket of an opened array.
func decodeArray(dec *json.Decoder) (jsonNode, bool) {
	node := jsonNode{kind: jsonArray}
	for dec.More() {
		child, ok := decodeNode(dec)
		if !ok {
			return jsonNode{}, false
		}
		node.arr = append(node.arr, child)
	}
	_, err := dec.Token() // the closing ']'
	return node, err == nil
}

// decodeObject parses the members and closing brace of an opened object.
func decodeObject(dec *json.Decoder) (jsonNode, bool) {
	node := jsonNode{kind: jsonObject}
	for dec.More() {
		member, ok := decodeMember(dec)
		if !ok {
			return jsonNode{}, false
		}
		node.members = append(node.members, member)
	}
	_, err := dec.Token() // the closing '}'
	return node, err == nil
}

// decodeMember parses one key/value member; the decoder yields only string
// keys, erroring on anything else before the key token arrives.
func decodeMember(dec *json.Decoder) (jsonMember, bool) {
	keyTok, err := dec.Token()
	if err != nil {
		return jsonMember{}, false
	}
	key, isString := keyTok.(string)
	value, ok := decodeNode(dec)
	return jsonMember{key: key, value: value}, isString && ok
}
