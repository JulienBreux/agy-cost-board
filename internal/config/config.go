package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// DefaultConfigFileName is the standard configuration file name searched in the workspace.
const DefaultConfigFileName = ".agy-cost-board.yaml"

// FileConfig represents the structure of the persisted configuration file.
type FileConfig struct {
	ProjectID      string `yaml:"project_id"`
	TelemetryTable string `yaml:"telemetry_table"`
	BillingTable   string `yaml:"billing_table"`
	SinkName       string `yaml:"sink_name"`
	SeatQuota      int    `yaml:"seat_quota"`
}

// Load reads and unmarshals the YAML config file from the given path.
// If the file does not exist, it returns an empty FileConfig without error.
func Load(path string) (*FileConfig, error) {
	if path == "" {
		path = DefaultConfigFileName
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if _, legacyErr := os.Stat(".agy-ge-board.yaml"); legacyErr == nil {
				path = ".agy-ge-board.yaml"
			}
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &FileConfig{}, nil
		}
		return nil, err
	}

	var cfg FileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save marshals and writes the configuration to the specified file path.
func Save(path string, cfg *FileConfig) error {
	if path == "" {
		path = DefaultConfigFileName
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ResolveSetting resolves a string value adhering to precedence:
// 1. Explicit CLI Flag (if non-empty)
// 2. Environment Variable (if set and non-empty)
// 3. Persisted Configuration Value (if non-empty)
// 4. Fallback Default Value
func ResolveSetting(flagVal, envKey, configVal, defaultVal string) string {
	if flagVal != "" {
		return flagVal
	}
	if envKey != "" {
		if envVal := os.Getenv(envKey); envVal != "" {
			return envVal
		}
	}
	if configVal != "" {
		return configVal
	}
	return defaultVal
}

// ResolveIntSetting resolves an integer value adhering to precedence:
// 1. Explicit CLI Flag (if non-zero)
// 2. Persisted Configuration Value (if non-zero)
// 3. Fallback Default Value
func ResolveIntSetting(flagVal, configVal, defaultVal int) int {
	if flagVal != 0 {
		return flagVal
	}
	if configVal != 0 {
		return configVal
	}
	return defaultVal
}
