// The clipboard block: how a copied range travels as plain TSV text.
// ParseBlock is the one definition of how that text becomes a grid again; the
// paste half of the model lives in paste.go.
package engine

import "strings"

// BlockText is clipboard-block source text: TAB-separated cells on
// newline-separated rows, as a copied range travels through a clipboard.
type BlockText string

// ParseBlock reads a clipboard block: CRLF and lone CR normalize to LF,
// exactly one trailing newline is ignored, rows split on LF and cells on TAB.
// Every line is data — a clipboard block has no comment or directive
// semantics, unlike a .tsvt file. An empty text is a single empty cell (the
// TSV serialization of one empty cell IS the empty string), so pasting it
// clears its target.
func ParseBlock(text BlockText) Grid {
	normalized := strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(string(text))
	trimmed := strings.TrimSuffix(normalized, "\n")
	rows := strings.Split(trimmed, "\n")
	out := make(Grid, len(rows))
	for r, row := range rows {
		out[r] = strings.Split(row, tab)
	}
	return out
}
