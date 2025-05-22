package handlers

import (
	"go-csv-import/internal/logger"
	"go-csv-import/internal/validation"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type JobPublisher interface {
	PublishImportJob(path string, maxRows int) error
}

// Upload file webservice
func Upload(publisher JobPublisher) gin.HandlerFunc {
	// TODO: Check file size to limit
	// TODO: handle Go channels to get errors and limit go routine for a lot of files
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			logger.Current.Error("Error uploading file:", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing file"})
			return
		}

		err = validation.IsValidCSV(file.Filename)
		if err != nil {
			logger.Current.Error("Error validating file type is a .csv:", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Save uploaded file through shared volume
		dst := filepath.Join("/shared", file.Filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			logger.Current.Error("Error saving file:", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Send file path to RabbitMQ
		if err := publisher.PublishImportJob(dst, 25000); err != nil {
			logger.Current.Error("Error publishing job to RabbitMQ:", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish job"})
			return
		}

		logger.Current.Info("File is being processed", "file", file.Filename)
		c.JSON(http.StatusOK, gin.H{"message": "File is being processed"})
	}
}
