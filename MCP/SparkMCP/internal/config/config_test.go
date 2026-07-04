package config_test

import (
	"testing"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Namespace != "default" {
		t.Errorf("expected default namespace, got %s", cfg.Namespace)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected info log level, got %s", cfg.LogLevel)
	}
	if cfg.ServerName != "spark-debug-mcp" {
		t.Errorf("unexpected server name: %s", cfg.ServerName)
	}
	if cfg.RetryMax != 3 {
		t.Errorf("expected retry max 3, got %d", cfg.RetryMax)
	}
}

func TestLoadKubeconfigDefault(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.KubeConfig == "" {
		t.Error("expected default kubeconfig path")
	}
}
