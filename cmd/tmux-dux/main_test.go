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

	configPath := filepath.Join(dir, "tmux-dux", "config.toml")
	commandsPath := filepath.Join(dir, "tmux-dux", "commands.toml")
	scriptsPath := filepath.Join(dir, "tmux-dux", "scripts")
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
	scriptsInfo, err := os.Stat(scriptsPath)
	if err != nil {
		t.Fatalf("stat scripts dir: %v", err)
	}
	if !scriptsInfo.IsDir() {
		t.Fatal("config init created scripts path as non-directory")
	}
}

func TestConfigInitWritesOnlySelectedConfigFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configPath := filepath.Join(dir, "tmux-dux", "config.toml")
	commandsPath := filepath.Join(dir, "tmux-dux", "commands.toml")
	scriptsPath := filepath.Join(dir, "tmux-dux", "scripts")

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
	if _, err := os.Stat(scriptsPath); !os.IsNotExist(err) {
		t.Fatal("config init --commands created scripts directory")
	}

	out.Reset()
	if err := configInit(&out, []string{"--script_dir"}); err != nil {
		t.Fatalf("configInit --script_dir returned error: %v", err)
	}
	scriptsInfo, err := os.Stat(scriptsPath)
	if err != nil {
		t.Fatalf("stat scripts dir: %v", err)
	}
	if !scriptsInfo.IsDir() {
		t.Fatal("config init --script_dir created scripts path as non-directory")
	}
}

func TestConfigInitRefusesWhenSelectedTargetExists(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configPath := filepath.Join(dir, "tmux-dux", "config.toml")
	commandsPath := filepath.Join(dir, "tmux-dux", "commands.toml")
	scriptsPath := filepath.Join(dir, "tmux-dux", "scripts")
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

	if err := os.Remove(commandsPath); err != nil {
		t.Fatalf("remove commands: %v", err)
	}
	if err := os.Mkdir(scriptsPath, 0o755); err != nil {
		t.Fatalf("create existing scripts dir: %v", err)
	}
	out.Reset()
	if err := configInit(&out, []string{"--script_dir"}); err == nil {
		t.Fatal("configInit --script_dir returned nil error with existing scripts directory")
	}
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatal("config init created config.toml even though scripts directory existed")
	}
	if _, err := os.Stat(commandsPath); !os.IsNotExist(err) {
		t.Fatal("config init created commands.toml even though scripts directory existed")
	}
}

func TestConfigInitRefusesAllWhenAnySelectedFileExists(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configPath := filepath.Join(dir, "tmux-dux", "config.toml")
	commandsPath := filepath.Join(dir, "tmux-dux", "commands.toml")
	scriptsPath := filepath.Join(dir, "tmux-dux", "scripts")
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
	if _, err := os.Stat(scriptsPath); !os.IsNotExist(err) {
		t.Fatal("config init created scripts directory even though config.toml existed")
	}
}

func TestConfigInitRejectsUnknownFlags(t *testing.T) {
	var out bytes.Buffer
	if err := configInit(&out, []string{"--bad"}); err == nil {
		t.Fatal("configInit returned nil error with unknown flag")
	}
}
