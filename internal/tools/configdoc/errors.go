package configdoc

import (
	"errors"
	"fmt"
)

// ErrEmptySpec is returned when the input spec has no fields.
var ErrEmptySpec = errors.New("configdoc: spec is empty")

// ErrOutputDir is returned when the output directory cannot be created or written to.
type ErrOutputDir struct {
	Path string
	Err  error
}

// Error returns the error message.
func (e *ErrOutputDir) Error() string {
	return fmt.Sprintf("configdoc: failed to write to %s: %v", e.Path, e.Err)
}

// Unwrap returns the underlying error.
func (e *ErrOutputDir) Unwrap() error {
	return e.Err
}
