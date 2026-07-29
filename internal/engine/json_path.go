// Package engine's JSON path layer: the dotted/indexed path syntax (a.b[0].c)
// compiled to steps, and the walk that resolves them against a document.
package engine

import (
	"strconv"
	"strings"
)

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
