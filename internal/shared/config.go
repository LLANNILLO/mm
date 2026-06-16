package shared

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}
