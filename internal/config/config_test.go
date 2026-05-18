package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileMissingReturnsDefaults(t *testing.T) {
	cfg, err := LoadFile(filepath.Join(t.TempDir(), "missing.toml"))
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}
	if got := len(cfg.Commands); got != 15 {
		t.Fatalf("default command count = %d, want 15", got)
	}
	if cfg.UI.Width != "75%" || cfg.UI.PopupWidth != "80%" {
		t.Fatalf("unexpected default UI: %#v", cfg.UI)
	}
}

func TestLoadFileParsesTOMLCommands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	input := `
[ui]
width = "60%"

[[commands]]
title = "Logs"
description = "Tail logs"
category = "Tools"
aliases = ["log", "tail"]
icon = "file"
popup = "tail -f app.log"
`
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}
	if cfg.UI.Width != "60%" {
		t.Fatalf("width = %q, want 60%%", cfg.UI.Width)
	}
	if cfg.UI.Height != "70%" {
		t.Fatalf("height = %q, want default 70%%", cfg.UI.Height)
	}
	if len(cfg.Commands) != 1 {
		t.Fatalf("command count = %d, want 1", len(cfg.Commands))
	}
	if cfg.Commands[0].Popup != "tail -f app.log" {
		t.Fatalf("popup = %q", cfg.Commands[0].Popup)
	}
}

func TestPathUsesXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/config-root")
	path, err := Path()
	if err != nil {
		t.Fatalf("Path returned error: %v", err)
	}
	want := filepath.Join("/tmp/config-root", "tmux-commander", "config.toml")
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestDefaultCommandsContainExpectedInitialSet(t *testing.T) {
	commands := DefaultCommands()
	want := map[string]bool{
		"Find Pane":        false,
		"Split Horizontal": false,
		"Split Vertical":   false,
		"Close Pane":       false,
		"Zoom / Unzoom":    false,
		"New Window":       false,
		"Rename Window":    false,
		"Close Window":     false,
		"Choose Session":   false,
		"New Session":      false,
		"Rename Session":   false,
		"Detach":           false,
		"Reload Config":    false,
		"Lazygit":          false,
		"Btop":             false,
	}
	for _, cmd := range commands {
		if _, ok := want[cmd.Title]; ok {
			want[cmd.Title] = true
		}
	}
	for title, found := range want {
		if !found {
			t.Fatalf("missing default command %q", title)
		}
	}
}
