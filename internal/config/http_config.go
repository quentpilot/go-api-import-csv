package config

type HttpConfig struct {
	Port             string // Log level (default: ":8080")
	MaxContentLength int64  // Max request size for a request in byte (default: "10485760" - 10 Mo)
	FileChunkLimit   int    // Split uploaded file after reached number of rows limit (default: "25000")
}

func (c *HttpConfig) Load() {
	LoadEnv()

	port := Get("HTTP_PORT", "8080")
	c.Port = ":" + port
	c.MaxContentLength = GetInt("HTTP_MAX_CONTENT_LENGTH", 10<<20)
	c.FileChunkLimit = int(GetInt("FILE_CHUNK_LIMIT", 25000))
}
