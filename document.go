// Package tsvsheet's document facade: a parsed .tsvt with its physical line
// layout retained, and the view its own directives declare.
package tsvsheet

import "github.com/tsvsheet/go-tsvsheet/internal/engine"

// Document is a parsed .tsvt file with its physical line layout retained, so
// comment and shebang lines — which the grid drops — survive editing and are
// written back in position by Text. Document is immutable: every editing
// operation returns a new Document. It is the one sanctioned way to serialize
// a .tsvt; frontends must never rebuild a file from a grid.
type Document = engine.Document

// ParseDocument reads a .tsvt file like Parse, additionally recording the
// physical line layout so comment and shebang lines are preserved by Text.
func ParseDocument(src []byte) (Document, error) { return engine.ParseDocument(src) }

// View is what a viewport does with a grid, as the sheet's own `#.` directives
// declare it: which rows and columns are hidden, carry headers, or stay
// anchored while the rest scrolls. Every field is a set of 1-based positions,
// resolved against the sheet's extent, so a host renders it without deriving
// anything itself.
type View = engine.View
