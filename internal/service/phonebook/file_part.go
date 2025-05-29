package phonebook

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"
)

type FilePart struct {
	FilePath    string        // File to treat
	TotalRows   int           // Total number of rows in the file
	ProcessTime time.Duration // Time taken to process the file
	Error       error         // Error that occurred during processing, if any
}

// Safe removes temporary file by checking that file exists and is not a directory
func (f *FilePart) Remove() error {
	i, err := os.Stat(f.FilePath)
	if err != nil {
		return err
	}

	if i.IsDir() {
		return errors.New("cannot remove directory")
	}

	slog.Info("Removing FilePart", "file", f.FilePath)
	return os.Remove(f.FilePath)
}

// PrintStat returns a formatted string with job statistics
func (f *FilePart) PrintStat() string {
	return fmt.Sprintf("FilePart %s has been treated in %0.3f sec with a total of %d rows", f.FilePath, f.ProcessTime.Seconds(), f.TotalRows)
}
