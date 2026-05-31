package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExampleConfigAndCommandsLoad(t *testing.T) {
	root := filepath.Join("..", "..")
	configBytes, err := os.ReadFile(filepath.Join(root, "examples", "default-config.toml"))
	if err != nil {
		t.Fatalf("read default config example: %v", err)
	}
	commandsBytes, err := os.ReadFile(filepath.Join(root, "examples", "default-commands.toml"))
	if err != nil {
		t.Fatalf("read default commands example: %v", err)
	}

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(configPath, configBytes, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "commands.toml"), commandsBytes, 0o600); err != nil {
		t.Fatalf("write commands: %v", err)
	}

	cfg, err := LoadFile(configPath)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}
	if len(cfg.Commands) < 40 {
		t.Fatalf("example command count = %d, want at least 40", len(cfg.Commands))
	}
}
