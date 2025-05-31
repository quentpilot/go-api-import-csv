package validation

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Check if the file has a ".csv" extension
func IsValidCSV(fileName string) error {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == ".csv" {
		return nil
	} else {
		return fmt.Errorf("invalid file type: %s. expected a .csv file", ext)
	}
}
