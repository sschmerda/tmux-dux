package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigInitWritesEmptyConfigFilesWhenNoConfigExists(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	var out bytes.Buffer
	if err := configInit(&out, nil); err != nil {
		t.Fatalf("configInit returned error: %v", err)
	}

	configPath := filepath.Join(dir, "tmux-commander", "config.toml")
	commandsPath := filepath.Join(dir, "tmux-commander", "commands.toml")
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	commandsBytes, err := os.ReadFile(commandsPath)
	if err != nil {
		t.Fatalf("read commands: %v", err)
	}
	if len(configBytes) != 0 {
		t.Fatal("config init wrote non-empty config.toml")
	}
	if len(commandsBytes) != 0 {
		t.Fatal("config init wrote non-empty commands.toml")
	}
}

func TestConfigInitWritesOnlySelectedConfigFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configPath := filepath.Join(dir, "tmux-commander", "config.toml")
	commandsPath := filepath.Join(dir, "tmux-commander", "commands.toml")

	var out bytes.Buffer
	if err := configInit(&out, []string{"--config"}); err != nil {
		t.Fatalf("configInit --config returned error: %v", err)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if _, err := os.Stat(commandsPath); !os.IsNotExist(err) {
		t.Fatal("config init --config created commands.toml")
	}

	out.Reset()
	if err := configInit(&out, []string{"--commands"}); err != nil {
		t.Fatalf("configInit --commands returned error: %v", err)
	}
	if _, err := os.Stat(commandsPath); err != nil {
		t.Fatalf("stat commands: %v", err)
	}
}

func TestConfigInitRefusesWhenSelectedConfigFileExists(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configPath := filepath.Join(dir, "tmux-commander", "config.toml")
	commandsPath := filepath.Join(dir, "tmux-commander", "commands.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("keep"), 0o600); err != nil {
		t.Fatalf("write existing config: %v", err)
	}
	var out bytes.Buffer
	if err := configInit(&out, []string{"--config"}); err == nil {
		t.Fatal("configInit --config returned nil error with existing config.toml")
	}
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after init: %v", err)
	}
	if string(configBytes) != "keep" {
		t.Fatal("config init overwrote existing config.toml")
	}
	if _, err := os.Stat(commandsPath); !os.IsNotExist(err) {
		t.Fatal("config init created commands.toml even though config.toml existed")
	}

	if err := os.Remove(configPath); err != nil {
		t.Fatalf("remove config: %v", err)
	}
	if err := os.WriteFile(commandsPath, []byte("keep commands"), 0o600); err != nil {
		t.Fatalf("write existing commands: %v", err)
	}
	out.Reset()
	if err := configInit(&out, []string{"--commands"}); err == nil {
		t.Fatal("configInit --commands returned nil error with existing commands.toml")
	}
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatal("config init created config.toml even though commands.toml existed")
	}
}

func TestConfigInitRefusesAllWhenAnySelectedFileExists(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configPath := filepath.Join(dir, "tmux-commander", "config.toml")
	commandsPath := filepath.Join(dir, "tmux-commander", "commands.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("keep"), 0o600); err != nil {
		t.Fatalf("write existing config: %v", err)
	}

	var out bytes.Buffer
	err := configInit(&out, nil)
	if err == nil {
		t.Fatal("configInit returned nil error with existing config.toml")
	}
	if !strings.Contains(err.Error(), configPath) {
		t.Fatalf("config init error = %q, want existing config path", err.Error())
	}
	if _, err := os.Stat(commandsPath); !os.IsNotExist(err) {
		t.Fatal("config init created commands.toml even though config.toml existed")
	}
}

func TestConfigInitRejectsUnknownFlags(t *testing.T) {
	var out bytes.Buffer
	if err := configInit(&out, []string{"--bad"}); err == nil {
		t.Fatal("configInit returned nil error with unknown flag")
	}
}
