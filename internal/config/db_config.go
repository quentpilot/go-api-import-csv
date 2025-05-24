package config

import "fmt"

type DbConfig struct {
	Dsn  string // DB connection url (default "appuser:apppass@tcp(localhost:3306)/contactdb?charset=utf8mb4&parseTime=True&loc=Local")
	Host string
	Port int
	Name string
	User string
	Pass string
}

func (c *DbConfig) Load() {
	LoadEnv()

	c.Host = Get("DB_HOST", "localhost")
	c.Port = int(GetInt("DB_PORT", 3306))
	c.Name = Get("DB_NAME", "contactdb")
	c.User = Get("DB_USER", "appuser")
	c.Pass = Get("DB_PASS", "apppass")

	c.Dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.User, c.Pass, c.Host, c.Port, c.Name)
}
