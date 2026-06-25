package main

import (
	"bytes"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	cmd := newVersionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("spark-debug-mcp")) {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestRootCommand(t *testing.T) {
	cmd := newRootCmd()
	if cmd.Use != "spark-debug-mcp" {
		t.Errorf("unexpected use: %s", cmd.Use)
	}
	if cmd.Commands()[0].Use != "serve" {
		t.Errorf("expected serve subcommand")
	}
}

func TestInitLogger(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn"} {
		logger, err := initLogger(level)
		if err != nil {
			t.Fatalf("level %s: %v", level, err)
		}
		if logger == nil {
			t.Fatalf("nil logger for level %s", level)
		}
	}
}
