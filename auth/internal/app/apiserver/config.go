package apiserver

// Config ...
type Config struct {
	BindAddr      string `toml:"bind_addr"`
	LogLevel      string `toml:"log_level"`
	DatabaseURL   string `toml:"database_url"`
	apiGatewayUrl string `toml:"api_gateway_url"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		BindAddr:      ":8000",
		LogLevel:      "info",
		DatabaseURL:   "host=localhost user=postgres dbname=users password=postgres sslmode=disable",
		apiGatewayUrl: "http://127.0.0.1:7000",
	}
}
