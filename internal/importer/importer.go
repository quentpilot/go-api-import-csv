package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

func ProcessFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)

	// Skip header
	headers, err := reader.Read()
	if err != nil {
		return err
	}
	fmt.Printf("CSV Headers : %v\n", headers)

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read record: %w", err)
		}

		// Print each line
		fmt.Printf("Read current line : %v\n", record)
		count++
	}

	fmt.Printf("File %s treated with %d rows\n", path, count)
	return nil
}
