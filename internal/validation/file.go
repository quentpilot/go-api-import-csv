package validation

import (
	"fmt"
	"go-csv-import/internal/handlers/worker"
	"net/http"
	"path/filepath"
	"strings"
)

// Checks if the file has a ".csv" extension
func IsValidCSV(fileName string) error {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == ".csv" {
		return nil
	} else {
		return fmt.Errorf("invalid file type: %s. expected a .csv file", ext)
	}
}

// Checks if file progress status indicated that contacts can be deleted.
func IsSafeDeletable(ps *worker.MessageProgressResponse, statusCode int) bool {
	// Delete anyway because worker memory may be reset
	if ps == nil {
		return true
	}

	// Clean case
	if ps.Status == string(worker.StatusCompleted) {
		return true
	}

	// Delete anyway because error doesn't means that contacts are not inserted
	if statusCode == http.StatusMultiStatus {
		return true
	}
	return false
}
