package importer

/* import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/bootstrap"
	"go-csv-import/internal/job"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func boot() {
	bootstrap.Init(app.AppConfig{LoggerName: "worker"})
}

func TestProcessFile_ValidCSV(t *testing.T) {
	boot()

	// Create a temporary CSV file
	tempFile, err := os.CreateTemp("", "testfile-*.csv")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write valid CSV content to the file
	_, err = tempFile.WriteString("name,age\nJohn,30\nJane,25\n")
	assert.NoError(t, err)
	tempFile.Close()

	job := job.ImportJob{
		FilePath: tempFile.Name(),
		MaxRows:  5000,
	}

	// Call ProcessFile
	err = ProcessFile(job)
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

	job := job.ImportJob{
		FilePath: tempFile.Name(),
		MaxRows:  5000,
	}

	// Call ProcessFile
	err = ProcessFile(job)
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

	job := job.ImportJob{
		FilePath: tempFile.Name(),
		MaxRows:  5000,
	}

	// Call ProcessFile
	err = ProcessFile(job)
	assert.Error(t, err)
	color.Green("Error caught: %v", err)
}

func TestProcessFile_FileNotFound(t *testing.T) {
	// Call ProcessFile with a non-existent file

	job := job.ImportJob{
		FilePath: "non_existent_file.csv",
		MaxRows:  5000,
	}

	err := ProcessFile(job)
	assert.Error(t, err)
	color.Green("Error caught:", err, "\n")
} */
