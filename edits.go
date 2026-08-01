// The sheet edit language facade (work order 011): parse an edits document,
// content-address a document, and apply the deterministic fold.
package tsvsheet

import "github.com/tsvsheet/go-tsvsheet/internal/engine"

// RevisionHex is a document's content address: the lowercase-hex SHA-256 of
// its canonical Text bytes. Equal revision ⇔ byte-equal document.
type RevisionHex = engine.RevisionHex

// Edits is a parsed edits document (application/vnd.tsvsheet.edits+tsv): a TSV
// stream of semantic operations. Immutable; Text returns the exact bytes
// parsed.
type Edits = engine.Edits

// Revision content-addresses d.
func Revision(d Document) RevisionHex { return engine.Revision(d) }

// ParseEdits reads an edits document; any malformed line is a specific
// ErrEdits* sentinel carrying its 1-based line number.
func ParseEdits(src []byte) (Edits, error) { return engine.ParseEdits(src) }

// Apply left-folds e's ops over d: atomic (a refused op rejects the whole
// batch), base-checked (ErrEditsBase when e names a revision that is not d's),
// and deterministic — nothing outside (d, e, limits) influences the result.
func Apply(d Document, e Edits, limits Limits) (Document, error) {
	return engine.Apply(d, e, limits)
}
