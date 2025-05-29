package config

import "time"

type HttpConfig struct {
	Port             string        // Log level (default: ":8080")
	MaxContentLength int64         // Max request size for a request in byte (default: "10485760" - 10 Mo)
	FileChunkLimit   uint          // Split uploaded file after reached number of rows limit (default: "25000")
	BatchInsert      uint          // Number of contact rows inserted by query (default: "5000")
	FileTimeout      time.Duration // Lifetime in seconds for file processing (default: 30)
}

func (c *HttpConfig) Load() {
	LoadEnv()

	port := Get("HTTP_PORT", "8080")
	c.Port = ":" + port
	c.MaxContentLength = int64(GetUint("HTTP_MAX_CONTENT_LENGTH", 10<<20))
	c.FileChunkLimit = uint(GetUint("FILE_CHUNK_LIMIT", 25000))
	c.BatchInsert = uint(GetUint("BATCH_INSERT", 5000))
	c.FileTimeout = time.Duration(GetUint("FILE_UPLOAD_TIMEOUT", 30)) * time.Second

	c.validate()
}

func (c *HttpConfig) validate() {
	if c.Port == "" {
		panicInvalidConfig("ENV var HTTP_PORT must not be empty")
	}
	if c.MaxContentLength < 100 {
		panicInvalidConfig("ENV var HTTP_MAX_CONTENT_LENGTH must be greater than 100")
	}
	if c.FileChunkLimit == 0 {
		panicInvalidConfig("ENV var FILE_CHUNK_LIMIT must be greater than zero")
	}
	if c.BatchInsert == 0 {
		panicInvalidConfig("ENV var BATCH_INSERT must be greater than zero")
	}
	if c.FileTimeout <= 0 {
		panicInvalidConfig("ENV var FILE_UPLOAD_TIMEOUT must be greater than zero")
	}

}
