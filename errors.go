package tsvsheet

import "github.com/tsvsheet/go-tsvsheet/internal/constants"

// Engine error sentinels returned to callers, matchable with errors.Is.
const (
	ErrSyntax       = constants.ErrSyntax
	ErrDocTooLarge  = constants.ErrDocTooLarge
	ErrInvalidValue = constants.ErrInvalidValue
	ErrNotFound     = constants.ErrNotFound
	ErrReadInput    = constants.ErrReadInput
	ErrWriteFile    = constants.ErrWriteFile

	ErrEditsAddress = constants.ErrEditsAddress
	ErrEditsApply   = constants.ErrEditsApply
	ErrEditsArity   = constants.ErrEditsArity
	ErrEditsBase    = constants.ErrEditsBase
	ErrEditsBlock   = constants.ErrEditsBlock
	ErrEditsOp      = constants.ErrEditsOp
)
