package config

type LoggerConfig struct {
	Level string // Log level (default: "info")
}

func (c *LoggerConfig) Load() {
	LoadEnv()

	c.Level = Get("LOG_LEVEL", "info")
}
