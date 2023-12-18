package config

type Configuration interface {
	GetString(key string, value string) string
	GetInt(key string, value int) int
}

var (
	Configs Configuration
)
