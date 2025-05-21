package main

import (
	"fmt"
	"go-csv-import/internal/logger"
	"go-csv-import/internal/queue"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type JobPublisher interface {
	PublishImportJob(path string, maxRows int) error
}

type RabbitPublisher struct{}

func (r *RabbitPublisher) PublishImportJob(path string, maxRows int) error {
	return queue.PublishImportJob(path, maxRows)
}

// Check if the file has a ".csv" extension
func validateFileType(fileName string) error {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == ".csv" {
		return nil
	} else {
		return fmt.Errorf("invalid file csv type: %s. expected a .csv file", ext)
	}
}

// Upload file webservice
func handleUpload(publisher JobPublisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			logger.Current.Info("Error uploading file:", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing file"})
			return
		}

		err = validateFileType(file.Filename)
		if err != nil {
			logger.Current.Info("Error validating file type is a .csv:", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Save uploaded file through shared volume
		dst := filepath.Join("/shared", file.Filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			logger.Current.Info("Error saving file:", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Send file path to RabbitMQ
		if err := publisher.PublishImportJob(dst, 25000); err != nil {
			logger.Current.Info("Error publishing job to RabbitMQ:", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish job"})
			return
		}

		logger.Current.Info("File is being processed", "file", file.Filename)
		c.JSON(http.StatusOK, gin.H{"message": "File is being processed"})
	}
}

func main() {
	if err := logger.InitCurrent("api", false); err != nil {
		panic(err)
	}

	r := gin.Default()

	publisher := &RabbitPublisher{}
	r.POST("/upload", handleUpload(publisher))

	logger.Current.Info("API Server runs on localhost:8080")
	r.Run(":8080")
}
