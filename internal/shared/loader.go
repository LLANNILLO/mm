package shared

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func LoadConfig(env string, modules []string) (*Config, error) {
	const configsDir = "configs"

	v := viper.New()
	v.BindEnv("database.dsn", "DATABASE_URL")        //nolint:errcheck
	v.BindEnv("logging.seq.endpoint", "SEQ_ENDPOINT") //nolint:errcheck

	if err := readInto(v, configsDir, "app"); err != nil {
		return nil, fmt.Errorf("app config: %w", err)
	}
	// optional per-environment app override (e.g. app.development.yaml)
	_ = mergeInto(v, configsDir, fmt.Sprintf("app.%s", env))

	for _, module := range modules {
		if err := mergeInto(v, configsDir, fmt.Sprintf("modules.%s", module)); err != nil {
			return nil, fmt.Errorf("module %s config: %w", module, err)
		}
		// env override is optional — silently skip if absent
		_ = mergeInto(v, configsDir, fmt.Sprintf("modules.%s.%s", module, env))
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	return &cfg, nil
}

// readInto loads the first matching file (yaml, yml, json) into v.
func readInto(v *viper.Viper, dir, name string) error {
	path, err := findConfigFile(dir, name)
	if err != nil {
		return err
	}
	v.SetConfigFile(path)
	return v.ReadInConfig()
}

// mergeInto merges a config file into an existing Viper instance.
func mergeInto(base *viper.Viper, dir, name string) error {
	path, err := findConfigFile(dir, name)
	if err != nil {
		return err
	}
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	return base.MergeConfigMap(v.AllSettings())
}

func findConfigFile(dir, name string) (string, error) {
	for _, ext := range []string{"yaml", "yml", "json"} {
		path := filepath.Join(dir, name+"."+ext)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("config file %q not found in %s (tried yaml, yml, json)", name, dir)
}
