package handlers

import (
	"go-csv-import/internal/logger"
	"go-csv-import/internal/service/phonebook"
	"go-csv-import/internal/utils"
	"go-csv-import/internal/validation"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type JobPublisher interface {
	PublishImportJob(path string, maxRows int) error
}

// Upload contacts file webservice
func Upload(publisher *phonebook.PhonebookHandler) gin.HandlerFunc {
	// TODO: Check file size to limit
	// TODO: handle Go channels to get errors and limit go routine for a lot of files
	return func(c *gin.Context) {
		logger.Info("Call endpoint /upload")
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

		uuid := uuid.New().String()
		totalRows, err := utils.FileCountRowsCsv(dst)
		if err != nil {
			logger.Error("Error counting rows in file", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Send file path to RabbitMQ
		job := &phonebook.FileMessage{
			Uuid:     uuid,
			FilePath: dst,
			MaxRows:  int(publisher.HttpConfig.FileChunkLimit),
		}

		logger.Trace("Publishing message to queue", "message", job)
		if err := publisher.Publish(job); err != nil {
			logger.Error("Error publishing message to queue", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish job"})
			return
		}

		statusUrl := publisher.HttpConfig.Host + publisher.HttpConfig.Port + "/upload/status/" + uuid + "/" + strconv.Itoa(totalRows)
		logger.Info("File is being processed", "file", file.Filename, "uuid", uuid, "status_url", statusUrl)
		c.JSON(http.StatusOK, gin.H{
			"message":    "File is being processed",
			"status_url": statusUrl,
		})
	}
}

/*
UploadStatus returns the status of the file upload process.
It checks how many rows have been processed and calculates the percentage of completion.
It returns the status as "Scheduled", "Processing", or "Completed" based on the number of processed rows.

TODO: Add a timeout to the request to avoid long processing times.
TODO: Add cache to avoid counting rows in the database every time the status is requested.
*/
func UploadStatus(p *phonebook.PhonebookHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.Param("uuid")
		totalRows, err := strconv.Atoi(c.Param("total"))
		if err != nil {
			logger.Error("Error parsing total rows", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid total rows parameter"})
			return
		}
		logger.Info("Call endpoint /upload/status", "uuid", uuid, "total", totalRows)

		pRows, err := p.Uploader.Repository.CountByReqId(uuid)
		if err != nil {
			logger.Error("Error counting inserted rows", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count inserted rows"})
			return
		}

		percentile := utils.MathRound(float64(pRows)/float64(totalRows)*100, 3)
		status := "Processing"
		if pRows == totalRows {
			status = "Completed"
		} else if pRows < totalRows && pRows == 0 {
			status = "Scheduled"
		}

		logger.Info("File is "+status, "uuid", uuid, "status", status, "percentile", percentile, "total_rows", totalRows, "processed_rows", pRows)
		c.JSON(http.StatusOK, gin.H{
			"status":         status,
			"total_rows":     totalRows,
			"processed_rows": pRows,
			"percentile":     percentile,
		})
	}
}
