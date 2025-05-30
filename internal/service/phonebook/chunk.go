package phonebook

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

/*
Check if the file has more than maximum rows configured.

This determines if the file should be chunked or not.

As each file will be processed in a separate goroutine,
we need to check if the file has more than the maximum number of rows defined by file.MaxRows
to get better performance and avoid memory issues.
*/
func (i *ContactUploader) mustChunkFile(file *FileMessage) (bool, error) {
	f, err := os.Open(file.FilePath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	reached := false
	for scanner.Scan() {
		count++
		if count > file.MaxRows {
			reached = true
			break
		}
	}

	return reached, nil
}

/*
Splits the file into smaller chunks.

Each chunk will have a maximum number of rows defined by file.MaxRows.

The chunk files will be created in the /tmp directory.

The chunk files will be named as <original_file_name>-part-<chunk_index>.csv.

The chunk files will contain the same header as the original file.

The chunk files will be returned as a slice of strings.

The original file will not be modified.
*/
func (i *ContactUploader) chunkFile(file *FileMessage) ([]FilePart, error) {
	f, err := os.Open(file.FilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	// Reads once the first line to get csv headers
	if !scanner.Scan() {
		return nil, NewFileError(file.FilePath, fmt.Errorf("failed to read first line: %w", scanner.Err()))
	}
	header := scanner.Text()

	var chunkFiles []FilePart
	var out *os.File
	var writer *bufio.Writer
	var currentLine int
	chunkIndex := 1

	createNewChunk := func() error {
		if writer != nil {
			writer.Flush()
			out.Close()
		}

		// Create a new chunked file
		filename := fmt.Sprintf("%v-part-%d.csv", filepath.Base(file.FilePath), chunkIndex)
		filename = filepath.Join("/tmp", filename)
		chunkIndex++
		fo, err := os.Create(filename)
		if err != nil {
			return err
		}
		out = fo
		writer = bufio.NewWriter(out)

		// Write the header to the new chunked file
		if _, err := writer.WriteString(header + "\n"); err != nil {
			return err
		}

		filePart := FilePart{FilePath: filename, Uuid: file.Uuid, TotalRows: 0, ProcessTime: 0}
		chunkFiles = append(chunkFiles, filePart)
		currentLine = 0
		return nil
	}

	if err := createNewChunk(); err != nil {
		return nil, err
	}

	for scanner.Scan() {
		line := scanner.Text()

		if currentLine >= file.MaxRows {
			if err := createNewChunk(); err != nil {
				return nil, err
			}
		}

		if _, err := writer.WriteString(line + "\n"); err != nil {
			return nil, err
		}
		currentLine++
	}

	if writer != nil {
		writer.Flush()
		out.Close()
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return chunkFiles, nil
}
