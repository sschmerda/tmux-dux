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
	if got := len(cfg.Commands); got != 19 {
		t.Fatalf("default command count = %d, want 19", got)
	}
	if cfg.UI.Width != "40%" || cfg.UI.PopupWidth != "80%" {
		t.Fatalf("unexpected default UI: %#v", cfg.UI)
	}
	if cfg.UI.Theme != "shades-of-purple" {
		t.Fatalf("theme = %q, want shades-of-purple", cfg.UI.Theme)
	}
	if !cfg.UI.ShowDescription {
		t.Fatal("show_description = false, want true")
	}
	if !cfg.UI.RecentCommands || cfg.UI.RecentLimit != 10 {
		t.Fatalf("recent defaults = %v %d, want true 10", cfg.UI.RecentCommands, cfg.UI.RecentLimit)
	}
}

func TestLoadFileParsesTOMLCommands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	input := `
[ui]
width = "60%"
theme = "custom"
glyphs = false
show_description = false
recent_commands = false
recent_limit = 5

[custom_theme]
background = "#111111"
title = "#eeeeee"
commander_border = "#ddddff"
prompt_border = "#ccccff"
prompt = "#aaaaaa"
query = "#bbbbbb"
search_bg = "#444444"
search_fg = "#eeeeff"
empty = "#cccccc"
chip_bg = "#222222"
selected_chip = "#ffccaa"
selected_chip_bg = "#332211"
glyph = "#dddddd"
match_fg = "#ffeeaa"
selected_match_fg = "#aaffee"
selected_bg = "#333333"

[[commands]]
title = "Logs"
description = "Tail logs"
category = "Tools"
aliases = ["log", "tail"]
icon = "file"
action = "popup"
command = "tail -f app.log"
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
	if cfg.UI.Height != "80%" {
		t.Fatalf("height = %q, want default 80%%", cfg.UI.Height)
	}
	if cfg.UI.Theme != "custom" {
		t.Fatalf("theme = %q, want custom", cfg.UI.Theme)
	}
	if cfg.UI.Glyphs {
		t.Fatal("glyphs = true, want false")
	}
	if cfg.UI.ShowDescription {
		t.Fatal("show_description = true, want false")
	}
	if cfg.UI.RecentCommands {
		t.Fatal("recent_commands = true, want false")
	}
	if cfg.UI.RecentLimit != 5 {
		t.Fatalf("recent_limit = %d, want 5", cfg.UI.RecentLimit)
	}
	if cfg.CustomTheme.Background != "#111111" || cfg.CustomTheme.Title != "#eeeeee" || cfg.CustomTheme.CommanderBorder != "#ddddff" || cfg.CustomTheme.PromptBorder != "#ccccff" || cfg.CustomTheme.Prompt != "#aaaaaa" || cfg.CustomTheme.Query != "#bbbbbb" || cfg.CustomTheme.SearchBG != "#444444" || cfg.CustomTheme.SearchFG != "#eeeeff" || cfg.CustomTheme.Empty != "#cccccc" || cfg.CustomTheme.ChipBG != "#222222" || cfg.CustomTheme.SelectedChip != "#ffccaa" || cfg.CustomTheme.SelectedChipBG != "#332211" || cfg.CustomTheme.Glyph != "#dddddd" || cfg.CustomTheme.MatchFG != "#ffeeaa" || cfg.CustomTheme.SelectedMatchFG != "#aaffee" || cfg.CustomTheme.SelectedBG != "#333333" {
		t.Fatalf("custom theme = %#v", cfg.CustomTheme)
	}
	if cfg.Commands[0].Action != "popup" || cfg.Commands[0].Command != "tail -f app.log" {
		t.Fatalf("action command = %q %q", cfg.Commands[0].Action, cfg.Commands[0].Command)
	}
	if cfg.Commands[0].PopupWidth != "95%" || cfg.Commands[0].PopupHeight != "85%" {
		t.Fatalf("popup size = %q x %q", cfg.Commands[0].PopupWidth, cfg.Commands[0].PopupHeight)
	}
	if len(cfg.Commands) != 6 {
		t.Fatalf("command count = %d, want 6", len(cfg.Commands))
	}
	for _, internal := range []string{InternalThemePreview, InternalClearRecent, InternalConfigPath, InternalEditConfig, InternalReloadConfig} {
		found := false
		for _, cmd := range cfg.Commands {
			if cmd.Internal == internal {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing internal command %q", internal)
		}
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

func TestLoadFileRejectsDeprecatedActionFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	input := `
[[commands]]
title = "Old"
tmux = "display-panes"
`
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := LoadFile(path); err == nil {
		t.Fatal("LoadFile returned nil error")
	}
}

func TestLoadFileRejectsCommandWithoutActionCommandPair(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	input := `
[[commands]]
title = "Broken"
action = "popup"
`
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := LoadFile(path); err == nil {
		t.Fatal("LoadFile returned nil error")
	}
}

func TestLoadFileRejectsNegativeRecentLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	input := `
[ui]
recent_limit = -1

[[commands]]
title = "Pane"
action = "tmux"
command = "display-panes"
`
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := LoadFile(path); err == nil {
		t.Fatal("LoadFile returned nil error")
	}
}

func TestDefaultCommandsContainExpectedInitialSet(t *testing.T) {
	commands := DefaultCommands()
	want := map[string]bool{
		"Find Pane":             false,
		"Split Horizontal":      false,
		"Split Vertical":        false,
		"Close Pane":            false,
		"Zoom / Unzoom":         false,
		"New Window":            false,
		"Rename Window":         false,
		"Close Window":          false,
		"Choose Session":        false,
		"New Session":           false,
		"Rename Session":        false,
		"Detach":                false,
		"Preview Themes":        false,
		"Clear Recent Commands": false,
		"List Config Path":      false,
		"Open / Edit Config":    false,
		"Reload Config":         false,
		"Lazygit":               false,
		"Btop":                  false,
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
