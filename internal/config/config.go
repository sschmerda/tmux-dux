package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/stefanschmerda/tmux-commander/internal/theme"
)

type UI struct {
	Width           string `toml:"width"`
	Height          string `toml:"height"`
	PopupWidth      string `toml:"popup_width"`
	PopupHeight     string `toml:"popup_height"`
	Border          bool   `toml:"border"`
	Theme           string `toml:"theme"`
	Glyphs          bool   `toml:"glyphs"`
	ShowDescription bool   `toml:"show_description"`
	ShowToggleHint  bool   `toml:"show_toggle_hint"`
	TmuxDescription bool   `toml:"tmux_description"`
	RecentCommands  bool   `toml:"recent_commands"`
	RecentLimit     int    `toml:"recent_limit"`
	TmuxRecentLimit int    `toml:"tmux_recent_limit"`
	TmuxModeKey     string `toml:"tmux_mode_key"`
}

type Command struct {
	Title       string   `toml:"title"`
	Description string   `toml:"description"`
	Category    string   `toml:"category"`
	Aliases     []string `toml:"aliases"`
	Icon        string   `toml:"icon"`
	Action      string   `toml:"action"`
	Command     string   `toml:"command"`
	PopupWidth  string   `toml:"popup_width"`
	PopupHeight string   `toml:"popup_height"`
	Internal    string   `toml:"-"`
}

type Config struct {
	UI          UI          `toml:"ui"`
	CustomTheme theme.Theme `toml:"custom_theme"`
	Commands    []Command   `toml:"commands"`
}

func DefaultUI() UI {
	return UI{
		Width:           "40%",
		Height:          "80%",
		PopupWidth:      "80%",
		PopupHeight:     "80%",
		Border:          true,
		Theme:           "shades-of-purple",
		Glyphs:          true,
		ShowDescription: true,
		ShowToggleHint:  true,
		TmuxDescription: true,
		RecentCommands:  true,
		RecentLimit:     10,
		TmuxRecentLimit: 10,
		TmuxModeKey:     "ctrl+t",
	}
}

func DefaultConfig() Config {
	return Config{
		UI:       DefaultUI(),
		Commands: ensureInternalCommands(DefaultCommands()),
	}
}

func Path() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "tmux-commander", "config.toml"), nil
}

func Load() (Config, string, error) {
	path, err := Path()
	if err != nil {
		return Config{}, "", err
	}
	cfg, err := LoadFile(path)
	return cfg, path, err
}

func LoadFile(path string) (Config, error) {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		return Config{}, err
	}
	type rawUI struct {
		Width           string `toml:"width"`
		Height          string `toml:"height"`
		PopupWidth      string `toml:"popup_width"`
		PopupHeight     string `toml:"popup_height"`
		Border          *bool  `toml:"border"`
		Theme           string `toml:"theme"`
		Glyphs          *bool  `toml:"glyphs"`
		ShowDescription *bool  `toml:"show_description"`
		ShowToggleHint  *bool  `toml:"show_toggle_hint"`
		TmuxDescription *bool  `toml:"tmux_description"`
		RecentCommands  *bool  `toml:"recent_commands"`
		RecentLimit     *int   `toml:"recent_limit"`
		TmuxRecentLimit *int   `toml:"tmux_recent_limit"`
		TmuxModeKey     string `toml:"tmux_mode_key"`
	}
	type rawConfig struct {
		UI          rawUI       `toml:"ui"`
		CustomTheme theme.Theme `toml:"custom_theme"`
		Commands    []Command   `toml:"commands"`
	}
	var raw rawConfig
	meta, err := toml.DecodeFile(path, &raw)
	if err != nil {
		return Config{}, err
	}
	if err := rejectDeprecatedActionFields(meta); err != nil {
		return Config{}, err
	}
	cfg := Config{
		UI:          DefaultUI(),
		CustomTheme: raw.CustomTheme,
		Commands:    raw.Commands,
	}
	if raw.UI.Width != "" {
		cfg.UI.Width = raw.UI.Width
	}
	if raw.UI.Height != "" {
		cfg.UI.Height = raw.UI.Height
	}
	if raw.UI.PopupWidth != "" {
		cfg.UI.PopupWidth = raw.UI.PopupWidth
	}
	if raw.UI.PopupHeight != "" {
		cfg.UI.PopupHeight = raw.UI.PopupHeight
	}
	if raw.UI.Border != nil {
		cfg.UI.Border = *raw.UI.Border
	}
	if raw.UI.Theme != "" {
		cfg.UI.Theme = raw.UI.Theme
	}
	if raw.UI.Glyphs != nil {
		cfg.UI.Glyphs = *raw.UI.Glyphs
	}
	if raw.UI.ShowDescription != nil {
		cfg.UI.ShowDescription = *raw.UI.ShowDescription
	}
	if raw.UI.ShowToggleHint != nil {
		cfg.UI.ShowToggleHint = *raw.UI.ShowToggleHint
	}
	if raw.UI.TmuxDescription != nil {
		cfg.UI.TmuxDescription = *raw.UI.TmuxDescription
	}
	if raw.UI.RecentCommands != nil {
		cfg.UI.RecentCommands = *raw.UI.RecentCommands
	}
	if raw.UI.RecentLimit != nil {
		cfg.UI.RecentLimit = *raw.UI.RecentLimit
	}
	if raw.UI.TmuxRecentLimit != nil {
		cfg.UI.TmuxRecentLimit = *raw.UI.TmuxRecentLimit
	}
	if raw.UI.TmuxModeKey != "" {
		cfg.UI.TmuxModeKey = raw.UI.TmuxModeKey
	}
	cfg.UI.TmuxModeKey = NormalizeKey(cfg.UI.TmuxModeKey)
	if cfg.UI.RecentLimit < 0 {
		return Config{}, errors.New("ui.recent_limit must be >= 0")
	}
	if cfg.UI.TmuxRecentLimit < 0 {
		return Config{}, errors.New("ui.tmux_recent_limit must be >= 0")
	}
	if len(cfg.Commands) == 0 {
		cfg.Commands = DefaultCommands()
	}
	cfg.Commands = ensureInternalCommands(cfg.Commands)
	if err := validateCommands(cfg.Commands); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func NormalizeKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, "-", "+")
	return key
}

func CommandKey(cmd Command) string {
	return strings.ToLower(strings.TrimSpace(cmd.Action)) + ":" + strings.TrimSpace(cmd.Command)
}

func rejectDeprecatedActionFields(meta toml.MetaData) error {
	for _, key := range meta.Undecoded() {
		if len(key) == 2 && key[0] == "commands" {
			switch key[1] {
			case "tmux", "shell", "popup":
				return fmt.Errorf("commands.%s is deprecated; use action and command", key[1])
			}
		}
	}
	return nil
}

func validateCommands(commands []Command) error {
	for index := range commands {
		cmd := &commands[index]
		if cmd.Internal != "" {
			continue
		}
		cmd.Action = strings.ToLower(strings.TrimSpace(cmd.Action))
		if cmd.Action == "" {
			return fmt.Errorf("command %q must define action", cmd.Title)
		}
		if strings.TrimSpace(cmd.Command) == "" {
			return fmt.Errorf("command %q must define command", cmd.Title)
		}
		switch cmd.Action {
		case "tmux", "shell", "popup":
		default:
			return fmt.Errorf("command %q has unsupported action %q", cmd.Title, cmd.Action)
		}
	}
	return nil
}

func ensureInternalCommands(commands []Command) []Command {
	existing := map[string]bool{}
	for _, cmd := range commands {
		if cmd.Internal != "" {
			existing[cmd.Internal] = true
		}
	}
	for _, cmd := range SettingsCommands() {
		if !existing[cmd.Internal] {
			commands = append(commands, cmd)
		}
	}
	return commands
}
