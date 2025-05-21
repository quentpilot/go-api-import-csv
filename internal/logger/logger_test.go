package logger

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitLogger(t *testing.T) {
	err := Init("test_logger")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	logFile, _ := os.Getwd()
	logFile = filepath.Join(logFile, "..", "..", "logs", "test_logger.log")

	// Check if the log file is created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatalf("Log file was not created: %v", err)
	}

	// Clean up
	os.Remove(logFile)
}

func TestInitLoggerColor(t *testing.T) {
	err := Init("test_logger")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	logFile, _ := os.Getwd()
	logFile = filepath.Join(logFile, "..", "..", "logs", "test_logger.log")

	// Check if the log file is created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatalf("Log file was not created: %v", err)
	}

	// Clean up
	os.Remove(logFile)
}

/* func TestWriteFile(t *testing.T) {
	err := Init("test_api_logger")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	logFile, _ := os.Getwd()
	logFile = filepath.Join(logFile, "..", "..", "logs", "test_api_logger.log")

	// Check if the log file is created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatalf("Log file was not created: %v", err)
	}

	log.Println("Test log message")
	log.SetOutput(os.Stdout)
	time.Sleep(1 * time.Second) // Wait for the log to be written

	file, err := os.OpenFile(logFile, os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("Log file was not opened: %v", err)
	}

	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	defer file.Close()
	// Clean up
	defer os.Remove(logFile)

	assert.Contains(t, line, "Test log message", "Log file does not contain the expected log message")

} */
