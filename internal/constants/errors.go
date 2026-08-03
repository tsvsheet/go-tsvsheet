// Package constants declares the tsvsheet engine's sentinel error values. The
// error mechanism (the matchable string type) lives in the shared gomatic/go-error
// library; these values are this package's own.
package constants

// Imported bare (the package is named error); this file declares only sentinels
// and uses no builtin error type, so each declaration reads errs.Const.
import errs "github.com/gomatic/go-error"

// Keep these constants sorted alphabetically.
const (
	ErrCommentCell  errs.Const = "cell text would make the row a comment line"
	ErrDocTooLarge  errs.Const = "document exceeds the resident cell budget"
	ErrEditsAddress errs.Const = "invalid edits address"
	ErrEditsApply   errs.Const = "edits op refused"
	ErrEditsArity   errs.Const = "wrong edits argument count"
	ErrEditsBase    errs.Const = "edits base mismatch"
	ErrEditsBlock   errs.Const = "invalid edits block"
	ErrEditsOp      errs.Const = "unknown edits op"
	ErrInvalidValue errs.Const = "invalid value"
	ErrNotFound     errs.Const = "not found"
	ErrReadInput    errs.Const = "failed to read input"
	ErrSyntax       errs.Const = "syntax error"
	ErrWriteFile    errs.Const = "failed to write file"
)
