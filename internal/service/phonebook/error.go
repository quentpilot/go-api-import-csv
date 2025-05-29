package phonebook

import (
	"context"
	"errors"
	"fmt"
	"go-csv-import/internal/db"
	"log/slog"

	"github.com/hashicorp/go-multierror"
)

// FileError represents an error that occurred while processing a file.
type FileError struct {
	FilePath string // Path to the file that caused the error
	Err      error
}

// NewFileError creates a new FileError instance with the specified file path and error.
func NewFileError(filePath string, err error) *FileError {
	return &FileError{
		FilePath: filePath,
		Err:      err,
	}
}

func (e *FileError) Error() string {
	return fmt.Sprintf("error processing file %s: %v", e.FilePath, e.Err)
}

func (e *FileError) Unwrap() error {
	return e.Err
}

func (q *PhonebookHandler) printTypedErrors(err error, file *FileMessage) {
	if errs, ok := err.(*multierror.Error); ok {
		for _, e := range errs.Errors {
			var fe *FileError
			if errors.As(e, &fe) {
				slog.Error("Error processing file", "file", fe.FilePath, "error", fe)
				continue
			}

			var de *db.DbError
			if errors.As(e, &de) {
				slog.Error("Database error", "error", de)
				continue
			}

			if errors.Is(e, context.Canceled) {
				slog.Error("Interrupted by SIGINT", "error", e)
				continue
			}

			if errors.Is(e, context.DeadlineExceeded) {
				slog.Error("Timeout", "error", e)
				continue
			}

			slog.Error("Unexpected error", "file", file.FilePath, "error", e)
		}
	} else {

		if ie, ok := err.(*FileError); ok {
			slog.Error("Error processing single file", "file", file.FilePath, "error", ie.Err)
		} else if de, ok := err.(*db.DbError); ok {
			slog.Error("Database error for single file", "error", de.Err)
		} else if errors.Is(err, context.Canceled) {
			slog.Error("Interrupted by SIGINT", "error", err)
		} else if errors.Is(err, context.DeadlineExceeded) {
			slog.Error("Timeout", "error", err)
		} else {
			slog.Error("Unexpected error for single file", "file", file.FilePath, "error", err)
		}
	}
}
