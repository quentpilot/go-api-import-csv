package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockPublisher struct{}

func (m mockPublisher) PublishImportJob(path string) error {
	return nil
}

func TestHandleUpload_ValidCSVFile(t *testing.T) {
	// Set up Gin router
	router := gin.Default()
	router.POST("/upload", handleUpload(mockPublisher{}))

	// Create a temporary CSV file
	tempFile, err := os.CreateTemp("", "testfile-*.csv")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write some content to the file
	_, err = tempFile.WriteString("name,age\nJohn,30\n")
	assert.NoError(t, err)
	tempFile.Close()

	// Create a new HTTP request with the file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", tempFile.Name())
	assert.NoError(t, err)

	file, err := os.Open(tempFile.Name())
	assert.NoError(t, err)
	defer file.Close()

	_, err = io.Copy(part, file)
	assert.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "File is being processed")
}

func TestHandleUpload_InvalidFileType(t *testing.T) {
	// Set up Gin router
	router := gin.Default()
	router.POST("/upload", handleUpload(mockPublisher{}))

	// Create a temporary non-CSV file
	tempFile, err := os.CreateTemp("", "testfile-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write some content to the file
	_, err = tempFile.WriteString("This is a test file.")
	assert.NoError(t, err)
	tempFile.Close()

	// Create a new HTTP request with the file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", tempFile.Name())
	assert.NoError(t, err)

	file, err := os.Open(tempFile.Name())
	assert.NoError(t, err)
	defer file.Close()

	_, err = io.Copy(part, file)
	assert.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid file csv type")
}

func TestHandleUpload_MissingFile(t *testing.T) {
	// Set up Gin router
	router := gin.Default()
	router.POST("/upload", handleUpload(mockPublisher{}))

	// Create a new HTTP request without a file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing file")
}
