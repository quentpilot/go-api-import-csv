package importer

import (
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestProcessFile_ValidCSV(t *testing.T) {
	// Create a temporary CSV file
	tempFile, err := os.CreateTemp("", "testfile-*.csv")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write valid CSV content to the file
	_, err = tempFile.WriteString("name,age\nJohn,30\nJane,25\n")
	assert.NoError(t, err)
	tempFile.Close()

	// Call ProcessFile
	err = ProcessFile(tempFile.Name())
	assert.NoError(t, err)
}

func TestProcessFile_EmptyCSV(t *testing.T) {
	// Create a temporary empty CSV file
	tempFile, err := os.CreateTemp("", "testfile-*.csv")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write only headers to the file
	_, err = tempFile.WriteString("name,age\n")
	assert.NoError(t, err)
	tempFile.Close()

	// Call ProcessFile
	err = ProcessFile(tempFile.Name())
	assert.NoError(t, err)
}

func TestProcessFile_InvalidCSV(t *testing.T) {
	// Create a temporary invalid CSV file
	tempFile, err := os.CreateTemp("", "testfile-*.csv")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write invalid CSV content to the file
	_, err = tempFile.WriteString("name,age\nJohn\nJane,25\n")
	assert.NoError(t, err)
	tempFile.Close()

	// Call ProcessFile
	err = ProcessFile(tempFile.Name())
	assert.Error(t, err)
	color.Green("Error caught: %v", err)
}

func TestProcessFile_FileNotFound(t *testing.T) {
	// Call ProcessFile with a non-existent file
	err := ProcessFile("non_existent_file.csv")
	assert.Error(t, err)
	color.Green("Error caught:", err, "\n")
}
