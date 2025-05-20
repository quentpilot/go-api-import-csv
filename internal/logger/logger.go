package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// Initialize multiple output logger. Write into filePath and print onto STDOUT
func InitLogger(filePath string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	m := io.MultiWriter(os.Stdout, f)

	log.SetOutput(m)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return nil
}
