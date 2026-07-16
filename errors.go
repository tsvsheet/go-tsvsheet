package tsvsheet

import "github.com/uplang/go-tsvsheet/internal/constants"

// Engine error sentinels returned to callers, matchable with errors.Is.
const (
	ErrSyntax       = constants.ErrSyntax
	ErrInvalidValue = constants.ErrInvalidValue
	ErrNotFound     = constants.ErrNotFound
	ErrReadInput    = constants.ErrReadInput
	ErrWriteFile    = constants.ErrWriteFile
)
