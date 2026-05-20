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
		Height:          "70%",
		PopupWidth:      "80%",
		PopupHeight:     "80%",
		Border:          true,
		Theme:           "shades-of-purple",
		Glyphs:          true,
		ShowDescription: true,
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
	cfg := Config{UI: DefaultUI()}
	meta, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return Config{}, err
	}
	if err := rejectDeprecatedActionFields(meta); err != nil {
		return Config{}, err
	}
	if cfg.UI.Width == "" {
		cfg.UI.Width = DefaultUI().Width
	}
	if cfg.UI.Height == "" {
		cfg.UI.Height = DefaultUI().Height
	}
	if cfg.UI.PopupWidth == "" {
		cfg.UI.PopupWidth = DefaultUI().PopupWidth
	}
	if cfg.UI.PopupHeight == "" {
		cfg.UI.PopupHeight = DefaultUI().PopupHeight
	}
	if cfg.UI.Theme == "" {
		cfg.UI.Theme = DefaultUI().Theme
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
	for _, cmd := range commands {
		if cmd.Internal == InternalThemePreview {
			return commands
		}
	}
	return append(commands, ThemePreviewCommand())
}
