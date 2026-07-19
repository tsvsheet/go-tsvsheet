package engine

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
)

// jsonKind tags a parsed JSON node's shape.
type jsonKind int

const (
	jsonNull jsonKind = iota
	jsonBool
	jsonNumber
	jsonString
	jsonArray
	jsonObject
)

// jsonNode is a parsed JSON document node. Object members keep document order
// and numbers keep their literal text, so JSONSET re-renders with a minimal
// diff (ADR 0011 §4).
type jsonNode struct {
	str     string // string value, or a number's literal text
	arr     []jsonNode
	members []jsonMember
	kind    jsonKind
	isTrue  bool
}

// jsonMember is one ordered object member.
type jsonMember struct {
	key   string
	value jsonNode
}

// pathText is a JSON path argument in the dotted/indexed form `a.b[0].c`.
type pathText string

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

// jsonStep is one parsed path step: an object key or an array index.
type jsonStep struct {
	key     string
	index   int
	isIndex bool
}

// parsePath parses the `a.b[0].c` path form: dot-separated keys, each
// optionally followed by bracketed indices, with a bare leading index run
// (`[0].name`) indexing the root. The empty path is the root; ok is false for
// a malformed path.
func parsePath(path pathText) ([]jsonStep, bool) {
	if path == "" {
		return nil, true
	}
	var steps []jsonStep
	for i, segment := range strings.Split(string(path), ".") {
		segSteps, ok := parseSegment(pathText(segment), i == 0)
		if !ok {
			return nil, false
		}
		steps = append(steps, segSteps...)
	}
	return steps, true
}

// parseSegment parses one dot-separated segment: `key`, `key[i]…`, or — for
// the first segment only — a bare `[i]…`.
func parseSegment(segment pathText, isFirst boolResult) ([]jsonStep, bool) {
	key, indices := splitBracket(segment)
	if key == "" && (!isFirst || indices == "") {
		return nil, false
	}
	var steps []jsonStep
	if key != "" {
		steps = append(steps, jsonStep{key: string(key)})
	}
	return appendIndexSteps(steps, indices)
}

// splitBracket splits a segment at its first bracket: `key[0][1]` becomes
// (`key`, `[0][1]`).
func splitBracket(segment pathText) (pathText, pathText) {
	if bracket := strings.Index(string(segment), "["); bracket >= 0 {
		return segment[:bracket], segment[bracket:]
	}
	return segment, ""
}

// appendIndexSteps parses a run of bracketed indices (`[0][2]`) onto steps.
func appendIndexSteps(steps []jsonStep, indices pathText) ([]jsonStep, bool) {
	for indices != "" {
		end := strings.Index(string(indices), "]")
		if !strings.HasPrefix(string(indices), "[") || end < 0 {
			return nil, false
		}
		n, err := strconv.Atoi(string(indices[1:end]))
		if err != nil || n < 0 {
			return nil, false
		}
		steps = append(steps, jsonStep{index: n, isIndex: true})
		indices = indices[end+1:]
	}
	return steps, true
}

// walkPath descends node step by step; found is false when a step does not
// resolve — a missing key, an out-of-range index, or a step into the wrong
// shape — which callers map to #N/A (ADR 0011 §4).
func walkPath(node jsonNode, steps []jsonStep) (jsonNode, bool) {
	for _, step := range steps {
		next, found := stepInto(node, step)
		if !found {
			return jsonNode{}, false
		}
		node = next
	}
	return node, true
}

// stepInto resolves one step against a node.
func stepInto(node jsonNode, step jsonStep) (jsonNode, bool) {
	if step.isIndex {
		if node.kind != jsonArray || step.index >= len(node.arr) {
			return jsonNode{}, false
		}
		return node.arr[step.index], true
	}
	return memberValue(node, textVal(step.key))
}

// memberValue is the value of an object's member by key.
func memberValue(node jsonNode, key textVal) (jsonNode, bool) {
	if node.kind != jsonObject {
		return jsonNode{}, false
	}
	for _, m := range node.members {
		if m.key == string(key) {
			return m.value, true
		}
	}
	return jsonNode{}, false
}

// jsonArgs parses the document (argument 0) and a path: a malformed document
// or path is #VALUE! (bad); a path that does not resolve reports found=false.
func jsonArgs(args []Value, path pathText) (at jsonNode, isFound boolResult, bad Value) {
	doc, ok := parseJSON(textVal(argText(args, 0)))
	if !ok {
		return jsonNode{}, false, errorValue(ErrValue)
	}
	steps, ok := parsePath(path)
	if !ok {
		return jsonNode{}, false, errorValue(ErrValue)
	}
	var found bool
	at, found = walkPath(doc, steps)
	return at, boolResult(found), Value{}
}

// optionalPath is the path argument when present, else the root path.
func optionalPath(args []Value) pathText {
	if len(args) < 2 {
		return ""
	}
	return pathText(argText(args, 1))
}

// fnJSONGet reads the value at a path: a JSON scalar maps into the value model
// (string→text, number→number, boolean→bool, null→empty) and a container
// returns its compact JSON text. Malformed JSON or a malformed path is
// #VALUE!; a path that does not resolve is #N/A.
func fnJSONGet(args []Value) Value {
	at, found, bad := jsonArgs(args, pathText(argText(args, 1)))
	if bad.isError() {
		return bad
	}
	if !found {
		return errorValue(ErrNA)
	}
	return jsonScalar(at)
}

// jsonScalar maps a JSON node into the value model; a container keeps its
// compact JSON text.
func jsonScalar(node jsonNode) Value {
	switch node.kind {
	case jsonNull:
		return emptyValue()
	case jsonBool:
		return boolValue(boolResult(node.isTrue))
	case jsonNumber:
		return jsonNumberValue(textVal(node.str))
	case jsonString:
		return stringValue(textVal(node.str))
	default: // a container
		return stringValue(textVal(renderJSON(node)))
	}
}

// jsonNumberValue parses a JSON number literal; one that overflows a float is
// #NUM!.
func jsonNumberValue(literal textVal) Value {
	n, err := strconv.ParseFloat(string(literal), 64)
	if err != nil {
		return errorValue(ErrNum)
	}
	return numberValue(floatVal(n))
}

// fnJSONType names the shape at a path: "null", "boolean", "number",
// "string", "array", or "object".
func fnJSONType(args []Value) Value {
	at, found, bad := jsonArgs(args, optionalPath(args))
	if bad.isError() {
		return bad
	}
	if !found {
		return errorValue(ErrNA)
	}
	return stringValue(textVal(typeNameOf(at.kind)))
}

// typeNameOf is the JSONTYPE name of a node kind.
func typeNameOf(kind jsonKind) string {
	switch kind {
	case jsonNull:
		return "null"
	case jsonBool:
		return "boolean"
	case jsonNumber:
		return "number"
	case jsonString:
		return "string"
	case jsonArray:
		return "array"
	default: // jsonObject
		return "object"
	}
}

// fnJSONLen is an array's element count or an object's member count; a scalar
// at the path is #VALUE!.
func fnJSONLen(args []Value) Value {
	at, found, bad := jsonArgs(args, optionalPath(args))
	if bad.isError() {
		return bad
	}
	if !found {
		return errorValue(ErrNA)
	}
	switch at.kind {
	case jsonArray:
		return numberValue(floatVal(len(at.arr)))
	case jsonObject:
		return numberValue(floatVal(len(at.members)))
	default:
		return errorValue(ErrValue)
	}
}

// fnJSONKeys spills an object's keys in document order as a column; a
// non-object at the path is #VALUE! and an empty object is #N/A (no rows —
// FILTER's no-match convention).
func fnJSONKeys(args []Value) Value {
	at, found, bad := jsonArgs(args, optionalPath(args))
	if bad.isError() {
		return bad
	}
	if !found {
		return errorValue(ErrNA)
	}
	if at.kind != jsonObject {
		return errorValue(ErrValue)
	}
	if len(at.members) == 0 {
		return errorValue(ErrNA)
	}
	return arrayValue(keysColumn(at.members))
}

// keysColumn shapes an object's keys as the N×1 array that spills.
func keysColumn(members []jsonMember) [][]Value {
	rows := make([][]Value, len(members))
	for i, m := range members {
		rows[i] = []Value{stringValue(textVal(m.key))}
	}
	return rows
}

// fnJSONSet writes a value at a path and returns the document's compact text —
// a pure text-to-text transform preserving member order, with a new key
// appended. A missing object key along the path materializes as an object;
// an array index past the last element, or a step into a scalar, is #N/A.
func fnJSONSet(args []Value) Value {
	doc, ok := parseJSON(textVal(argText(args, 0)))
	if !ok {
		return errorValue(ErrValue)
	}
	steps, ok := parsePath(pathText(argText(args, 1)))
	if !ok {
		return errorValue(ErrValue)
	}
	updated, bad := setPath(doc, steps, jsonFromValue(args[2]))
	if bad.isError() {
		return bad
	}
	return stringValue(textVal(renderJSON(updated)))
}

// jsonFromValue maps a cell value to a JSON node: number→number, bool→boolean,
// empty→null, and anything textual (text, date) → its canonical text as a
// string (ADR 0011 §4).
func jsonFromValue(v Value) jsonNode {
	switch v.kind {
	case kindNumber:
		return jsonNode{kind: jsonNumber, str: v.String()}
	case kindBool:
		return jsonNode{kind: jsonBool, isTrue: v.num != 0}
	case kindEmpty:
		return jsonNode{kind: jsonNull}
	default: // kindString, kindDate
		return jsonNode{kind: jsonString, str: v.String()}
	}
}

// setPath returns node with value written at steps (immutably — the input
// node is never modified).
func setPath(node jsonNode, steps []jsonStep, value jsonNode) (jsonNode, Value) {
	if len(steps) == 0 {
		return value, Value{}
	}
	if steps[0].isIndex {
		return setIndex(node, steps, value)
	}
	return setKey(node, steps, value)
}

// setIndex writes through an array index; a non-array node or an index past
// the last element is #N/A (indices never extend an array).
func setIndex(node jsonNode, steps []jsonStep, value jsonNode) (jsonNode, Value) {
	i := steps[0].index
	if node.kind != jsonArray || i >= len(node.arr) {
		return jsonNode{}, errorValue(ErrNA)
	}
	child, bad := setPath(node.arr[i], steps[1:], value)
	if bad.isError() {
		return jsonNode{}, bad
	}
	arr := append([]jsonNode(nil), node.arr...)
	arr[i] = child
	return jsonNode{kind: jsonArray, arr: arr}, Value{}
}

// setKey writes through an object key, materializing a missing key as an
// appended empty object; keying into a non-object is #N/A.
func setKey(node jsonNode, steps []jsonStep, value jsonNode) (jsonNode, Value) {
	if node.kind != jsonObject {
		return jsonNode{}, errorValue(ErrNA)
	}
	members := append([]jsonMember(nil), node.members...)
	at := memberIndex(members, textVal(steps[0].key))
	if at < 0 {
		members = append(members, jsonMember{key: steps[0].key, value: jsonNode{kind: jsonObject}})
		at = len(members) - 1
	}
	child, bad := setPath(members[at].value, steps[1:], value)
	if bad.isError() {
		return jsonNode{}, bad
	}
	members[at].value = child
	return jsonNode{kind: jsonObject, members: members}, Value{}
}

// memberIndex is the position of key among members, or -1.
func memberIndex(members []jsonMember, key textVal) int {
	for i, m := range members {
		if m.key == string(key) {
			return i
		}
	}
	return -1
}
