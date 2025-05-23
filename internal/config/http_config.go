package config

type HttpConfig struct {
	Port string // Log level (default: ":8080")
}

func (c *HttpConfig) Load() {
	LoadEnv()

	port := Get("API_PORT", "8080")

	c.Port = ":" + port
}
