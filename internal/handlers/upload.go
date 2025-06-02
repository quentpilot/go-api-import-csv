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

func HtmlUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload.html", nil)
	}
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
		if err := publisher.Publish(job, phonebook.MessageTypeUpload); err != nil {
			logger.Error("Error publishing message to queue", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish job"})
			return
		}

		statusUrl := publisher.HttpConfig.Host + publisher.HttpConfig.Port + "/upload/status/" + uuid
		deleteUrl := publisher.HttpConfig.Host + publisher.HttpConfig.Port + "/delete/" + uuid
		logger.Info("File is being processed", "file", file.Filename, "uuid", uuid, "status_url", statusUrl)
		c.JSON(http.StatusAccepted, gin.H{
			"message":    "File is being processed",
			"status_url": statusUrl,
			"delete_url": deleteUrl,
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

		ps, status, err := internalUploadStatus(uuid)
		if err != nil {
			c.JSON(status, gin.H{"message": err.Error()})
		}

		logger.Info("Progress status received", "status", ps.Status, "total_rows", ps.Total, "processed_rows", ps.Inserted, "percentile", ps.Percentile)

		c.JSON(status, ps)
	}
}

func internalUploadStatus(uuid string) (message *worker.MessageProgressResponse, statusCode int, err error) {
	// Check cache data
	if cached, found := cache.CacheApiUploadStatus.Get(uuid); found {
		logger.Info("Send progress status from cache")
		if msg, ok := cached.(*worker.MessageProgressResponse); ok {
			return msg, http.StatusOK, nil
		} else {
			return nil, http.StatusOK, nil
		}
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
			logger.Warn("Timeout error getting progress from worker", "error", err)
			return nil, http.StatusGatewayTimeout, fmt.Errorf("request to worker timed out")
		} else if errors.Is(err, context.DeadlineExceeded) {
			logger.Warn("Deadline exceeded error getting progress from worker", "error", err)
			return nil, http.StatusGatewayTimeout, fmt.Errorf("request to worker deadline exceeded")
		} else {
			logger.Warn("Error getting progress from worker", "error", err, "status_code", resp.StatusCode)
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to get progress status from worker")
		}
	}
	if resp.StatusCode >= http.StatusBadRequest {
		logger.Warn("Error status from worker", "error", err, "status_code", resp.StatusCode, "body", resp.Body)
		return nil, resp.StatusCode, fmt.Errorf("progess status not found")
	}

	defer resp.Body.Close()
	logger.Debug("Worker response received", "status_code", resp.StatusCode)

	// Parse server API response to send client API response
	ps := &worker.MessageProgressResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&ps); err != nil {
		logger.Error("Error decoding progress response", "error", err)
		return nil, http.StatusBadGateway, fmt.Errorf("corrupted progress status data")
	}

	cache.CacheApiUploadStatus.Set(uuid, ps, 0)

	return ps, resp.StatusCode, nil
}

func Delete(publisher *phonebook.PhonebookHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.Param("uuid")
		logger.Info("Call endpoint /delete", "uuid", uuid)

		ps, status, err := internalUploadStatus(uuid)
		if err != nil || ps == nil {
			logger.Debug("Force delete anyway", "progress", ps, "error", err)
		}

		if !validation.IsSafeDeletable(ps, status) {
			logger.Warn("Cannot delete contacts when upload is not completed")
			c.JSON(http.StatusConflict, gin.H{"message": "Upload is not completed yet"})
			return
		}

		job := &phonebook.FileMessage{
			Uuid: uuid,
		}

		logger.Trace("Publishing message to queue", "message", job)
		if err := publisher.Publish(job, phonebook.MessageTypeDelete); err != nil {
			logger.Error("Error publishing message to queue", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish job"})
			return
		}

		logger.Info("Contacts are being deleted")
		c.JSON(http.StatusOK, gin.H{"message": "Contacts are being deleted"})
	}
}
