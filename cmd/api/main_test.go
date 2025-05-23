package main

import (
	"bytes"
	"go-csv-import/internal/app"
	"go-csv-import/internal/bootstrap"
	"go-csv-import/internal/handlers"
	"go-csv-import/internal/middleware"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockPublisher struct{}

func (m mockPublisher) PublishImportJob(path string, maxRows int) error {
	return nil
}

func boot() {
	bootstrap.Init(app.AppConfig{LoggerName: "api"})
}

func TestHandleUpload_ValidCSVFile(t *testing.T) {
	boot()

	// Set up Gin router
	router := gin.Default()
	router.POST("/upload", handlers.Upload(mockPublisher{}))

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
	router.POST("/upload", handlers.Upload(mockPublisher{}))

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
	router.POST("/upload", handlers.Upload(mockPublisher{}))

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

func TestHandleUpload_TooLargeFile(t *testing.T) {
	// Set up Gin router with a limit of 1 MB
	router := gin.Default()
	router.POST("/upload",
		middleware.LimitRequestSize(1<<20),
		handlers.Upload(mockPublisher{}),
	)

	// Simulate a file too large
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	part, err := writer.CreateFormFile("file", "big.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Append 2 MB in the file
	content := strings.Repeat("A", 2<<20)
	part.Write([]byte(content))

	assert.Equal(t, int64(2<<20), int64(len(content)))

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "request body is too large") {
		t.Errorf("unexpected response body: %s", rec.Body.String())
	}
}

func TestHandleUpload_NotTooLargeFile(t *testing.T) {
	// Set up Gin router with a limit of 2 MB
	router := gin.Default()
	router.POST("/upload",
		middleware.LimitRequestSize(2<<20),
		handlers.Upload(mockPublisher{}),
	)

	// Simulate a file smaller than the limit
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	part, err := writer.CreateFormFile("file", "big.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Append 1 MB to the file
	content := strings.Repeat("A", 1<<20)
	part.Write([]byte(content))

	assert.Equal(t, int64(1<<20), int64(len(content)))

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "File is being processed") {
		t.Errorf("unexpected response body: %s", rec.Body.String())
	}
}
