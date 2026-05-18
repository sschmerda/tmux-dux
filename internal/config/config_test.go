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
	if got := len(cfg.Commands); got != 16 {
		t.Fatalf("default command count = %d, want 16", got)
	}
	if cfg.UI.Width != "75%" || cfg.UI.PopupWidth != "80%" {
		t.Fatalf("unexpected default UI: %#v", cfg.UI)
	}
	if cfg.UI.Theme != "shades-of-purple" {
		t.Fatalf("theme = %q, want shades-of-purple", cfg.UI.Theme)
	}
}

func TestLoadFileParsesTOMLCommands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	input := `
[ui]
width = "60%"
theme = "custom"

[custom_theme]
background = "#111111"
title = "#eeeeee"
prompt = "#aaaaaa"
query = "#bbbbbb"
empty = "#cccccc"
selected_bg = "#333333"

[[commands]]
title = "Logs"
description = "Tail logs"
category = "Tools"
aliases = ["log", "tail"]
icon = "file"
popup = "tail -f app.log"
popup_width = "95%"
popup_height = "85%"
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
	if cfg.UI.Theme != "custom" {
		t.Fatalf("theme = %q, want custom", cfg.UI.Theme)
	}
	if cfg.CustomTheme.Background != "#111111" || cfg.CustomTheme.Title != "#eeeeee" || cfg.CustomTheme.Prompt != "#aaaaaa" || cfg.CustomTheme.Query != "#bbbbbb" || cfg.CustomTheme.Empty != "#cccccc" || cfg.CustomTheme.SelectedBG != "#333333" {
		t.Fatalf("custom theme = %#v", cfg.CustomTheme)
	}
	if len(cfg.Commands) != 2 {
		t.Fatalf("command count = %d, want 2", len(cfg.Commands))
	}
	if cfg.Commands[0].Popup != "tail -f app.log" {
		t.Fatalf("popup = %q", cfg.Commands[0].Popup)
	}
	if cfg.Commands[0].PopupWidth != "95%" || cfg.Commands[0].PopupHeight != "85%" {
		t.Fatalf("popup size = %q x %q", cfg.Commands[0].PopupWidth, cfg.Commands[0].PopupHeight)
	}
	if cfg.Commands[1].Internal != InternalThemePreview {
		t.Fatalf("internal command = %q, want %q", cfg.Commands[1].Internal, InternalThemePreview)
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
		"Preview Themes":   false,
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
