package phonebook

import (
	"errors"
	"go-csv-import/internal/logger"
	"os"
)

// FileMessage is a structure to transport message information through RabbitMQ
type FileMessage struct {
	FilePath string `json:"filepath"` // Uploaded file
	MaxRows  int    `json:"max_rows"` // Max number of file rows to process by worker
}

// Safe removes temporary file by checking that file exists and is not a directory
func (j *FileMessage) Remove() error {
	i, err := os.Stat(j.FilePath)
	if err != nil {
		return err
	}

	if i.IsDir() {
		return errors.New("cannot remove directory")
	}

	logger.Debug("Removing FileMessage", "file", j.FilePath)
	return os.Remove(j.FilePath)
}
