package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

var colorEnabled bool

type MultiLogger interface {
	/*
		Initialize multiple output logger. Write into "{PROJECT_DIR}/logs/{name}.log" file and print onto STDOUT.
		The name must be the filename without extension.
	*/
	Init() error
}

type Logger struct {
}

type LoggerColor struct {
}

func Init(name string) error {
	/* filePath, err := getLogFilePath(name)
	if err != nil {
		return err
	} */

	filePath := filepath.Join("logs", name+".log")

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

func _Init(name string) error {
	filePath, err := getLogFilePath(name)
	if err != nil {
		return err
	}

	fmt.Println("[logger] Initialize log file path:", filePath)

	fWriter, err := createLogFile(filePath)
	if err != nil {
		return err
	}

	setWriters(os.Stdout, fWriter)

	return nil
}

func (l *LoggerColor) Init(name string, useColor bool) error {
	filePath, err := getLogFilePath(name)
	if err != nil {
		return err
	}

	fmt.Println("[logger] Initialize log file path:", filePath)

	fWriter, err := createLogFile(filePath)
	if err != nil {
		return err
	}

	colorEnabled = useColor
	var oWriter io.Writer
	oWriter = os.Stdout
	if useColor {
		oWriter = color.Output
	}

	setWriters(oWriter, fWriter)

	return nil
}

// Get the absolute path of the project directory
func getLogDir() (string, error) {
	execPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	logsDir := filepath.Join(execPath, "..", "..", "logs")

	return logsDir, nil
}

// Get the absolute file name where to write the logs.
// The file name is the name of the file without extension
func getLogFilePath(name string) (string, error) {
	dir, err := getLogDir()
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(dir, name+".log")

	return filePath, nil
}

// Create the log file and ensure the /logs/ directory exists
func createLogFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f, nil
}

// Set the writers for the logger
func setWriters(writers ...io.Writer) {
	m := io.MultiWriter(writers...)

	log.SetOutput(m)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

/*
Initialize multiple output logger. Write into ./logs/{name}.log and print onto STDOUT.

The name must be the filename without extension.
*/
func InitLogger(name string) error {
	filePath, err := getLogFilePath(name)
	if err != nil {
		return err
	}

	fmt.Println("[logger] Initialize log file path:", filePath)

	fWriter, err := createLogFile(filePath)
	if err != nil {
		return err
	}

	setWriters(os.Stdout, fWriter)

	return nil
}

func Inf(v ...any) {
	prefix := "[INFO]"
	if colorEnabled {
		log.Println(color.HiCyanString(prefix), fmt.Sprint(v...))
	} else {
		log.Println(prefix, v)
	}
}
