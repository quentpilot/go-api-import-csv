package config

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

/* func TestFallback(t *testing.T) {
	ch := &HttpConfig{}
	ch.Load()
	assert.Equal(t, ":8080", ch.Port)

	ca := &ApmqConfig{}
	ca.Load()
	assert.Equal(t, "amqp://guest:guest@rabbitmq:5672/", ca.Dsn)

	cl := &LoggerConfig{}
	cl.Load()
	assert.Equal(t, "info", cl.Level)
} */

func TestGet_WithFallback(t *testing.T) {
	os.Unsetenv("MISSING_KEY")
	assert.Equal(t, "default", Get("MISSING_KEY", "default"))

	os.Setenv("EXISTING_KEY", "value")
	assert.Equal(t, "value", Get("EXISTING_KEY", "default"))
}

func TestGetBool_WithFallback(t *testing.T) {
	os.Unsetenv("DEBUG")
	assert.Equal(t, true, GetBool("DEBUG", true))

	os.Setenv("DEBUG", "false")
	assert.Equal(t, false, GetBool("DEBUG", true))
}

func TestGetInt_WithFallback(t *testing.T) {
	os.Unsetenv("LIMIT")
	assert.Equal(t, int64(10), GetInt("LIMIT", 10))

	os.Setenv("LIMIT", "42")
	assert.Equal(t, int64(42), GetInt("LIMIT", 10))
}

func TestGetFloat_WithFallback(t *testing.T) {
	os.Unsetenv("THRESHOLD")
	assert.Equal(t, 1.5, GetFloat("THRESHOLD", 1.5))

	os.Setenv("THRESHOLD", "3.14")
	assert.Equal(t, 3.14, GetFloat("THRESHOLD", 1.5))
}

func TestLoadEnv_LoadsVariables(t *testing.T) {
	content := "TEST_ENV_VAR=hello_test\n"
	tmpFile, err := os.CreateTemp("", "test_env")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	tmpFile.Close()

	err = godotenv.Load(tmpFile.Name())
	assert.NoError(t, err)

	val := os.Getenv("TEST_ENV_VAR")
	assert.Equal(t, "hello_test", val)
}
