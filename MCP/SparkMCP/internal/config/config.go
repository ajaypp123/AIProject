// Package config handles application configuration via viper.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	defaultKubeConfig = ""
	envPrefix         = "SPARK_DEBUG"
)

// Config holds all server configuration.
type Config struct {
	KubeConfig     string        `mapstructure:"kubeconfig"`
	InCluster      bool          `mapstructure:"inCluster"`
	Namespace      string        `mapstructure:"namespace"`
	LogLevel       string        `mapstructure:"logLevel"`
	OutputFormat   string        `mapstructure:"outputFormat"`
	RetryMax       int           `mapstructure:"retryMax"`
	RetryDelay     time.Duration `mapstructure:"retryDelay"`
	ServerName     string        `mapstructure:"serverName"`
	ServerVersion  string        `mapstructure:"serverVersion"`
	SparkAppLabel  string        `mapstructure:"sparkAppLabel"`
	DriverLabelKey string        `mapstructure:"driverLabelKey"`
	ExecLabelKey   string        `mapstructure:"executorLabelKey"`
}

// Load reads configuration from file, environment, and defaults.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName(".spark-debug")
	v.SetConfigType("yaml")
	v.AddConfigPath("$HOME")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/spark-debug-mcp")

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.KubeConfig == "" {
		if home, err := os.UserHomeDir(); err == nil {
			cfg.KubeConfig = home + "/.kube/config"
		}
	}

	return cfg, nil
}

// BindFlags binds cobra flags to viper.
func BindFlags(v *viper.Viper) {
	v.BindEnv("kubeconfig", "KUBECONFIG")
	v.BindEnv("inCluster", "SPARK_DEBUG_IN_CLUSTER")
	v.BindEnv("namespace", "SPARK_DEBUG_NAMESPACE")
	v.BindEnv("logLevel", "SPARK_DEBUG_LOG_LEVEL")
	v.BindEnv("outputFormat", "SPARK_DEBUG_OUTPUT_FORMAT")
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("kubeconfig", defaultKubeConfig)
	v.SetDefault("inCluster", false)
	v.SetDefault("namespace", "default")
	v.SetDefault("logLevel", "info")
	v.SetDefault("outputFormat", "json")
	v.SetDefault("retryMax", 3)
	v.SetDefault("retryDelay", "500ms")
	v.SetDefault("serverName", "spark-debug-mcp")
	v.SetDefault("serverVersion", "1.0.0")
	v.SetDefault("sparkAppLabel", "sparkoperator.k8s.io/app-name")
	v.SetDefault("driverLabelKey", "spark-role")
	v.SetDefault("driverLabelValue", "driver")
	v.SetDefault("executorLabelKey", "spark-role")
	v.SetDefault("executorLabelValue", "executor")
}
