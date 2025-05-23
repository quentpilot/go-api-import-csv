package config

type ApmqConfig struct {
	Dsn   string // AMQP connection string (default: "amqp://guest:guest@rabbitmq:5672/")
	Queue string // AMQP queue name (default: "import_queue")
}

func (c *ApmqConfig) Load() {
	LoadEnv()

	c.Dsn = Get("AMQP_DSN", "amqp://guest:guest@rabbitmq:5672/")
	c.Queue = Get("AMQP_QUEUE", "import_queue")
}
