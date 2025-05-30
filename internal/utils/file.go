package utils

import (
	"bufio"
	"os"
)

// FileCountRows counts the number of rows in a file.
func FileCountRows(filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}

// FileCountRows counts the number of rows in a file, excluding the header row for CSV files.
func FileCountRowsCsv(filePath string) (int, error) {
	count, err := FileCountRows(filePath)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		count-- // Exclude header row
	}
	return count, nil
}
