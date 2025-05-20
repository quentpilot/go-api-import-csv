package importer

import (
	"encoding/csv"
	"fmt"
	"go-csv-import/internal/logger"
	"io"
	"log"
	"os"
	"time"
)

func ProcessFile(path string) error {
	if err := logger.InitLogger("logs/worker.log"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	start := time.Now()

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
	log.Println("CSV Headers:", headers)

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("failed to read record: %w", err)
			return fmt.Errorf("failed to read record: %w", err)
		}

		// Print each line
		log.Println("Read current line:", record)
		count++
	}

	duration := time.Since(start)
	log.Printf("File %s treated with %d rows in %v\n", path, count, duration)
	return nil
}
