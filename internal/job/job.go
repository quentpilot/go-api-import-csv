package job

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// TempFileJob is an interface for jobs that create temporary files
type TempFileJob interface {
	Remove() error // Removes temporary file
}

// Job is a structure to store job information through RabbitMQ
type ImportJob struct {
	FilePath string `json:"filepath"` // Uploaded file
	MaxRows  int    `json:"max_rows"` // Max number of file rows to process by worker
}

// Safe removes temporary file by checking that file exists and is not a directory
func (j *ImportJob) Remove() error {
	i, err := os.Stat(j.FilePath)
	if err != nil {
		return err
	}

	if i.IsDir() {
		return errors.New("cannot remove directory")
	}

	return os.Remove(j.FilePath)
}

// JobStat is a structure to store job statistics
type JobStat struct {
	FilePath    string        // File to treat
	TotalRows   int           // Total number of rows in the file
	ProcessTime time.Duration // Time taken to process the file
}

// Safe removes temporary file by checking that file exists and is not a directory
func (j *JobStat) Remove() error {
	i, err := os.Stat(j.FilePath)
	if err != nil {
		return err
	}

	if i.IsDir() {
		return errors.New("cannot remove directory")
	}

	return os.Remove(j.FilePath)
}

// PrintStat returns a formatted string with job statistics
func (j *JobStat) PrintStat() string {
	return fmt.Sprintf("File %s has been treated in %0.3f sec with a total of %d rows", j.FilePath, j.ProcessTime.Seconds(), j.TotalRows)
}
