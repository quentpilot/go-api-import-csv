package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFallback(t *testing.T) {
	ch := &HttpConfig{}
	ch.Load()
	assert.Equal(t, ":8080", ch.Port)

	ca := &ApmqConfig{}
	ca.Load()
	assert.Equal(t, "amqp://guest:guest@rabbitmq:5672/", ca.Dsn)

	cl := &LoggerConfig{}
	cl.Load()
	assert.Equal(t, "info", cl.Level)
}
