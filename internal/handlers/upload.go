package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-csv-import/internal/cache"
	"go-csv-import/internal/handlers/worker"
	"go-csv-import/internal/logger"
	"go-csv-import/internal/service/phonebook"
	"go-csv-import/internal/validation"
	"net/http"
	"os"
	"path/filepath"
	"time"

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
			c.JSON(http.StatusBadRequest, gin.H{"message": "Missing File"})
			return
		}

		logger.Trace("Validating file type", "filename", file.Filename)
		err = validation.IsValidCSV(file.Filename)
		if err != nil {
			logger.Error("Error validating file type is a .csv", "error", err)
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"message": err.Error()})
			return
		}

		// Save uploaded file through shared volume
		dst := filepath.Join("/shared", file.Filename)
		logger.Debug("Saving uploaded file", "filepath", dst)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			logger.Error("Error saving file", "message", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot save file"})
			return
		}

		uuid := uuid.New().String()

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

		statusUrl := publisher.HttpConfig.Host + publisher.HttpConfig.Port + "/upload/status/" + uuid
		logger.Info("File is being processed", "file", file.Filename, "uuid", uuid, "status_url", statusUrl)
		c.JSON(http.StatusAccepted, gin.H{
			"message":    "File is being processed",
			"status_url": statusUrl,
			"uuid":       uuid,
		})
	}
}

/*
UploadStatus returns the status of the file upload process.
It checks how many rows have been processed and calculates the percentage of completion.
It returns the status as "Scheduled", "Processing", or "Completed" based on the number of processed rows.
*/
func UploadStatus(p *phonebook.PhonebookHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.Param("uuid")
		logger.Info("Call endpoint /upload/status", "uuid", uuid)

		// Check cache data
		if cached, found := cache.CacheApiUploadStatus.Get(uuid); found {
			logger.Info("Send progress status from cache")
			c.JSON(http.StatusOK, cached)
			return
		}

		// Call worker API
		url := fmt.Sprintf("http://worker:9090/upload/status/%s", uuid)
		logger.Debug("Call worker API to get progress", "url", url)

		client := http.Client{
			Timeout: 2 * time.Second,
		}

		resp, err := client.Get(url)
		if err != nil {
			if os.IsTimeout(err) {
				logger.Error("Timeout error getting progress from worker", "error", err)
				c.JSON(http.StatusGatewayTimeout, gin.H{"message": "Request to worker timed out"})
				return
			} else if errors.Is(err, context.DeadlineExceeded) {
				logger.Error("Deadline exceeded error getting progress from worker", "error", err)
				c.JSON(http.StatusGatewayTimeout, gin.H{"message": "Request to worker deadline exceeded"})
				return
			} else {
				logger.Error("Error getting progress from worker", "error", err, "status_code", resp.StatusCode)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get progress status from worker"})
				return
			}
		}
		if resp.StatusCode >= http.StatusBadRequest {
			logger.Error("Error status from worker", "error", err, "status_code", resp.StatusCode, "body", resp.Body)
			c.JSON(resp.StatusCode, gin.H{"message": "Progess Status Not Found"})
			return
		}

		defer resp.Body.Close()
		logger.Debug("Worker response received", "status_code", resp.StatusCode)

		// Parse server API response to send client API response
		ps := &worker.MessageProgressResponse{}
		if err := json.NewDecoder(resp.Body).Decode(&ps); err != nil {
			logger.Error("Error decoding progress response", "error", err)
			c.JSON(http.StatusBadGateway, gin.H{"message": "Corrupted progress status data"})
			return
		}

		cache.CacheApiUploadStatus.Set(uuid, ps, 0)
		logger.Info("Progress status received", "status", ps.Status, "total_rows", ps.Total, "processed_rows", ps.Inserted, "percentile", ps.Percentile)

		c.JSON(resp.StatusCode, ps)
	}
}

func HtmlUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload.html", nil)
	}
}
