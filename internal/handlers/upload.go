package handlers

import (
	"go-csv-import/internal/logger"
	"go-csv-import/internal/service/phonebook"
	"go-csv-import/internal/validation"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type JobPublisher interface {
	PublishImportJob(path string, maxRows int) error
}

// Upload contacts file webservice
func Upload(publisher *phonebook.PhonebookHandler) gin.HandlerFunc {
	// TODO: Check file size to limit
	// TODO: handle Go channels to get errors and limit go routine for a lot of files
	return func(c *gin.Context) {
		logger.Debug("Call endpoint /upload")
		logger.Trace("Get file from form data")
		file, err := c.FormFile("file")
		if err != nil {
			logger.Error("Error uploading file", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing file"})
			return
		}

		logger.Trace("Validating file type", "filename", file.Filename)
		err = validation.IsValidCSV(file.Filename)
		if err != nil {
			logger.Error("Error validating file type is a .csv", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Save uploaded file through shared volume
		dst := filepath.Join("/shared", file.Filename)
		logger.Debug("Saving uploaded file", "filepath", dst)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			logger.Error("Error saving file", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Send file path to RabbitMQ
		job := &phonebook.FileMessage{
			FilePath: dst,
			MaxRows:  int(publisher.HttpConfig.FileChunkLimit),
		}

		logger.Trace("Publishing message to queue", "message", job)
		if err := publisher.Publish(job); err != nil {
			logger.Error("Error publishing message to queue", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish job"})
			return
		}

		logger.Info("File is being processed", "file", file.Filename)
		c.JSON(http.StatusOK, gin.H{"message": "File is being processed"})
	}
}
