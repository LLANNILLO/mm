package shared

type Config struct {
	Database       DatabaseConfig       `mapstructure:"database"`
	Logging        LoggingConfig        `mapstructure:"logging"`
	Cache          CacheConfig          `mapstructure:"cache"`
	Authentication AuthenticationConfig `mapstructure:"authentication"`
	Users          UsersConfig          `mapstructure:"users"`
}

type AuthenticationConfig struct {
	IssuerURL string `mapstructure:"issuer_url"`
	Audience  string `mapstructure:"audience"`
}

type UsersConfig struct {
	Keycloak KeycloakConfig `mapstructure:"keycloak"`
}

type KeycloakConfig struct {
	AdminURL                 string `mapstructure:"admin_url"`
	TokenURL                 string `mapstructure:"token_url"`
	ConfidentialClientID     string `mapstructure:"confidential_client_id"`
	ConfidentialClientSecret string `mapstructure:"confidential_client_secret"`
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
