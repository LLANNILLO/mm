package shared

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Cache    CacheConfig    `mapstructure:"cache"`
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

type CacheConfig struct {
	Address string `mapstructure:"address"`
}
