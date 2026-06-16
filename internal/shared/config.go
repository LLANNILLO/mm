package shared

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type LoggingConfig struct {
	Level string    `mapstructure:"level"`
	Seq   SeqConfig `mapstructure:"seq"`
}

type SeqConfig struct {
	Endpoint string `mapstructure:"endpoint"`
}
