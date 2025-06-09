package apiserver

type Config struct {
	BindAddr      string `toml:"bind_addr"`
	LogLevel      string `toml:"log_level"`
	DatabaseURL   string `toml:"database_url"`
	storePath     string `toml:"store_path"`
	secretKey     string `toml:"secret_key"`
	apiGatewayUrl string `toml:"api_gateway_url"`
}

func NewConfig() *Config {
	return &Config{
		BindAddr:      ":8080",
		LogLevel:      "info",
		DatabaseURL:   "host=localhost user=postgres dbname=s3 password=postgres sslmode=disable",
		storePath:     "storage",
		secretKey:     "secret",
		apiGatewayUrl: "http://127.0.1:7000",
	}
}
