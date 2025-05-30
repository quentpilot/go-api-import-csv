package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fatih/color"
)

func main() {
	var lines uint
	flag.UintVar(&lines, "lines", 1000, "Number of lines to generate into CSV file")
	flag.Parse()

	createFile(lines)

	os.Exit(0)
}

func createFile(lines uint) {
	filename := fmt.Sprintf("gen_contact_%d.csv", lines)
	filePath := filepath.Join("testdata", filename)

	file, err := os.Create(filePath)
	if err != nil {
		color.Red("Error creating file: %v\n", err)
		os.Exit(1)
	}

	defer file.Close()
	color.Blue("CSV file created: %s\n", filePath)

	writeFile(file, lines, 1000)
}

func writeFile(file *os.File, lines uint, batchSize uint) {
	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Comma = ';'

	batch := make([][]string, 0, batchSize) // Batch size of 1000

	// Write CSV header
	header := []string{"Phone", "Lastname", "Firstname"}
	_ = writer.Write(header)

	total := int(lines)
	start := time.Now()

	color.Yellow("Generating %d lines...\n", lines)
	for i := 1; i <= total; i++ {
		phone := "07" + strconv.Itoa(rand.Intn(100000000))
		first := "Customer " + strconv.Itoa(i)
		last := "Doe " + strconv.Itoa(i)

		row := []string{
			phone, first, last,
		}
		batch = append(batch, row)

		if len(batch) >= int(batchSize) {
			writer.WriteAll(batch)
			batch = batch[:0] // Reset the batch
		}
	}
	if len(batch) > 0 {
		writer.WriteAll(batch)
	}

	color.Green("CSV file generated with %d in %v: %s\n", lines, time.Since(start), file.Name())
}
