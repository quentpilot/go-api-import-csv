package config

import (
	"time"
)

type ApmqConfig struct {
	Dsn      string        // AMQP connection string (default: "amqp://guest:guest@rabbitmq:5672/")
	Queue    string        // AMQP queue name (default: "import_queue")
	Lifetime time.Duration // AMQP message lifetime process in seconds (default: 60)
}

func (c *ApmqConfig) Load() {
	LoadEnv()

	c.Dsn = Get("AMQP_DSN", "amqp://guest:guest@rabbitmq:5672/")
	c.Queue = Get("AMQP_QUEUE", "import_queue")
	c.Lifetime = time.Duration(int(GetUint("AMQP_LIFETIME", 60))) * time.Second

	c.Validate()
}

func (c *ApmqConfig) Validate() {
	if c.Dsn == "" {
		panicInvalidConfig("ENV var AMQP_DSN must not be empty")
	}
	if c.Queue == "" {
		panicInvalidConfig("ENV var AMQP_QUEUE must not be empty")
	}
	if c.Lifetime <= 0 {
		panicInvalidConfig("ENV var AMQP_LIFETIME must be greater than zero")
	}
}
