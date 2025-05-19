package main

import (
	"fmt"
	"go-csv-import/internal/queue"
)

func main() {
	fmt.Println("Worker is listening...")
	queue.ConsumeImportJobs()
}
