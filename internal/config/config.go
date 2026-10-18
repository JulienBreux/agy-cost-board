package config

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// DefaultConfigFileName is the standard configuration file name searched in the workspace.
const DefaultConfigFileName = ".agy-cost-board.yaml"

// Config represents the strongly typed application configuration.
type Config struct {
	ProjectID         string   `mapstructure:"project_id" yaml:"project_id" json:"project_id"`
	MonitoredProjects []string `mapstructure:"monitored_projects" yaml:"monitored_projects,omitempty" json:"monitored_projects,omitempty"`
	TelemetryTable    string   `mapstructure:"telemetry_table" yaml:"telemetry_table" json:"telemetry_table"`
	BillingTable      string   `mapstructure:"billing_table" yaml:"billing_table" json:"billing_table"`
	SinkName          string   `mapstructure:"sink_name" yaml:"sink_name" json:"sink_name"`
	DatasetName       string   `mapstructure:"dataset" yaml:"dataset" json:"dataset"`
	SeatQuota         int      `mapstructure:"seat_quota" yaml:"seat_quota" json:"seat_quota"`
	Demo              bool     `mapstructure:"demo" yaml:"demo" json:"demo"`
	Format            string   `mapstructure:"format" yaml:"format" json:"format"`
	Port              int      `mapstructure:"port" yaml:"port" json:"port"`
	Host              string   `mapstructure:"host" yaml:"host" json:"host"`
	MCPEnabled        bool     `mapstructure:"mcp_enabled" yaml:"mcp_enabled" json:"mcp_enabled"`
	LogLevel          string   `mapstructure:"log_level" yaml:"log_level,omitempty" json:"log_level,omitempty"`
	LogFormat         string   `mapstructure:"log_format" yaml:"log_format,omitempty" json:"log_format,omitempty"`
}

// ParseMonitoredProjects splits and deduplicates comma- or whitespace-separated project IDs.
func ParseMonitoredProjects(items ...string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, raw := range items {
		for _, part := range strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == ' ' || r == '\t' || r == '\n'
		}) {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" && !seen[trimmed] {
				seen[trimmed] = true
				out = append(out, trimmed)
			}
		}
	}
	return out
}

// FileConfig is an alias for Config for backwards compatibility.
type FileConfig = Config

type contextKey struct{}

// WithConfig returns a new Context that carries the provided Config value.
func WithConfig(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, contextKey{}, cfg)
}

// FromContext extracts the Config from ctx, or returns nil if none exists.
func FromContext(ctx context.Context) *Config {
	if ctx == nil {
		return nil
	}
	if cfg, ok := ctx.Value(contextKey{}).(*Config); ok {
		return cfg
	}
	return nil
}

// SetupViper initializes a Viper instance with the AGY_COST_BOARD_ environment
// variable prefix, key replacers, aliases, and sane defaults.
func SetupViper(v *viper.Viper) {
	v.SetEnvPrefix("AGY_COST_BOARD")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()

	// Strictly AGY_COST_BOARD_ prefixed environment variables
	_ = v.BindEnv("project_id", "AGY_COST_BOARD_PROJECT_ID", "AGY_COST_BOARD_PROJECT")
	_ = v.BindEnv("monitored_projects", "AGY_COST_BOARD_MONITORED_PROJECTS")
	_ = v.BindEnv("telemetry_table", "AGY_COST_BOARD_TELEMETRY_TABLE")
	_ = v.BindEnv("billing_table", "AGY_COST_BOARD_BILLING_TABLE")
	_ = v.BindEnv("seat_quota", "AGY_COST_BOARD_SEAT_QUOTA")
	_ = v.BindEnv("sink_name", "AGY_COST_BOARD_SINK_NAME")
	_ = v.BindEnv("dataset", "AGY_COST_BOARD_DATASET")
	_ = v.BindEnv("demo", "AGY_COST_BOARD_DEMO")
	_ = v.BindEnv("format", "AGY_COST_BOARD_FORMAT")
	_ = v.BindEnv("port", "AGY_COST_BOARD_PORT")
	_ = v.BindEnv("host", "AGY_COST_BOARD_HOST")
	_ = v.BindEnv("mcp_enabled", "AGY_COST_BOARD_MCP_ENABLED")
	_ = v.BindEnv("log_level", "AGY_COST_BOARD_LOG_LEVEL")
	_ = v.BindEnv("log_format", "AGY_COST_BOARD_LOG_FORMAT")

	// Aliases for seamless flag name and config key resolution
	v.RegisterAlias("project", "project_id")
	v.RegisterAlias("monitored-projects", "monitored_projects")
	v.RegisterAlias("telemetry-table", "telemetry_table")
	v.RegisterAlias("billing-table", "billing_table")
	v.RegisterAlias("seat-quota", "seat_quota")
	v.RegisterAlias("sink-name", "sink_name")
	v.RegisterAlias("mcp-enabled", "mcp_enabled")
	v.RegisterAlias("log-level", "log_level")
	v.RegisterAlias("log-format", "log_format")

	// Default fallback values
	v.SetDefault("seat_quota", 10)
	v.SetDefault("format", "table")
	v.SetDefault("port", 8080)
	v.SetDefault("host", "0.0.0.0")
	v.SetDefault("sink_name", "agy-inference-sink")
	v.SetDefault("dataset", "antigravity_telemetry")
	v.SetDefault("mcp_enabled", true)
	v.SetDefault("log_level", "info")
	v.SetDefault("log_format", "auto")
}

// LoadViper reads and unmarshals the configuration into a strongly typed Config struct
// using the supplied Viper instance. If path is empty, it searches for .agy-cost-board.yaml/.json
// in the current directory and user home directory.
func LoadViper(v *viper.Viper, path string) (*Config, error) {
	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.AddConfigPath(".")
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			v.AddConfigPath(home)
		}
		v.SetConfigName(".agy-cost-board")
	}

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if len(cfg.MonitoredProjects) > 0 {
		cfg.MonitoredProjects = ParseMonitoredProjects(cfg.MonitoredProjects...)
	} else if raw := v.GetString("monitored_projects"); raw != "" {
		cfg.MonitoredProjects = ParseMonitoredProjects(raw)
	}

	return &cfg, nil
}

// Load reads and unmarshals the YAML/JSON config file from the given path using an isolated Viper instance.
// If the file does not exist, it returns an empty/default Config without error.
func Load(path string) (*Config, error) {
	v := viper.New()
	SetupViper(v)
	return LoadViper(v, path)
}

// Save marshals and writes the configuration to the specified file path.
func Save(path string, cfg *Config) error {
	path = cmp.Or(path, DefaultConfigFileName)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// ResolveSetting resolves a string value adhering to precedence:
// 1. Explicit CLI Flag (if non-empty)
// 2. Environment Variable (if set and non-empty)
// 3. Persisted Configuration Value (if non-empty)
// 4. Fallback Default Value
func ResolveSetting(flagVal, envKey, configVal, defaultVal string) string {
	var envVal string
	if envKey != "" {
		envVal = os.Getenv(envKey)
	}
	return cmp.Or(flagVal, envVal, configVal, defaultVal)
}

// ResolveIntSetting resolves an integer value adhering to precedence:
// 1. Explicit CLI Flag (if non-zero)
// 2. Persisted Configuration Value (if non-zero)
// 3. Fallback Default Value
func ResolveIntSetting(flagVal, configVal, defaultVal int) int {
	return cmp.Or(flagVal, configVal, defaultVal)
}
