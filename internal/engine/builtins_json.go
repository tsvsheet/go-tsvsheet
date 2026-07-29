package engine

import (
	"strconv"
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
