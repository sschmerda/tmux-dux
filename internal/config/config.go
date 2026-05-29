package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/sschmerda/tmux-commander/internal/theme"
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
}

type Keys struct {
	TmuxMode         string `toml:"tmux_mode"`
	MoveUp           string `toml:"move_up"`
	MoveDown         string `toml:"move_down"`
	ScrollUp         string `toml:"scroll_up"`
	ScrollDown       string `toml:"scroll_down"`
	HalfPageUp       string `toml:"half_page_up"`
	HalfPageDown     string `toml:"half_page_down"`
	NextCategory     string `toml:"next_category"`
	PreviousCategory string `toml:"previous_category"`
}

type Command struct {
	Title       string   `toml:"title"`
	Description string   `toml:"description"`
	Category    string   `toml:"category"`
	Aliases     []string `toml:"aliases"`
	Icon        string   `toml:"icon"`
	Action      string   `toml:"action"`
	Command     string   `toml:"command"`
	Prompt      string   `toml:"prompt"`
	PopupWidth  string   `toml:"popup_width"`
	PopupHeight string   `toml:"popup_height"`
	Internal    string   `toml:"-"`
}

type Config struct {
	UI          UI          `toml:"ui"`
	Keys        Keys        `toml:"keys"`
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
	}
}

func DefaultKeys() Keys {
	return Keys{
		TmuxMode:         "ctrl+t",
		MoveUp:           "ctrl+p",
		MoveDown:         "ctrl+n",
		ScrollUp:         "ctrl+y",
		ScrollDown:       "ctrl+e",
		HalfPageUp:       "ctrl+u",
		HalfPageDown:     "ctrl+d",
		NextCategory:     "tab",
		PreviousCategory: "shift+tab",
	}
}

func DefaultConfig() Config {
	return Config{
		UI:       DefaultUI(),
		Keys:     DefaultKeys(),
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
	}
	type rawConfig struct {
		UI          rawUI       `toml:"ui"`
		Keys        Keys        `toml:"keys"`
		CustomTheme theme.Theme `toml:"custom_theme"`
		Commands    []Command   `toml:"commands"`
	}
	var raw rawConfig
	meta, err := toml.DecodeFile(path, &raw)
	if err != nil {
		return Config{}, err
	}
	if err := rejectDeprecatedFields(meta); err != nil {
		return Config{}, err
	}
	cfg := Config{
		UI:          DefaultUI(),
		Keys:        DefaultKeys(),
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
	applyKeys(&cfg.Keys, raw.Keys)
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

func applyKeys(keys *Keys, raw Keys) {
	if raw.TmuxMode != "" {
		keys.TmuxMode = raw.TmuxMode
	}
	if raw.MoveUp != "" {
		keys.MoveUp = raw.MoveUp
	}
	if raw.MoveDown != "" {
		keys.MoveDown = raw.MoveDown
	}
	if raw.ScrollUp != "" {
		keys.ScrollUp = raw.ScrollUp
	}
	if raw.ScrollDown != "" {
		keys.ScrollDown = raw.ScrollDown
	}
	if raw.HalfPageUp != "" {
		keys.HalfPageUp = raw.HalfPageUp
	}
	if raw.HalfPageDown != "" {
		keys.HalfPageDown = raw.HalfPageDown
	}
	if raw.NextCategory != "" {
		keys.NextCategory = raw.NextCategory
	}
	if raw.PreviousCategory != "" {
		keys.PreviousCategory = raw.PreviousCategory
	}
	keys.TmuxMode = NormalizeKey(keys.TmuxMode)
	keys.MoveUp = NormalizeKey(keys.MoveUp)
	keys.MoveDown = NormalizeKey(keys.MoveDown)
	keys.ScrollUp = NormalizeKey(keys.ScrollUp)
	keys.ScrollDown = NormalizeKey(keys.ScrollDown)
	keys.HalfPageUp = NormalizeKey(keys.HalfPageUp)
	keys.HalfPageDown = NormalizeKey(keys.HalfPageDown)
	keys.NextCategory = NormalizeKey(keys.NextCategory)
	keys.PreviousCategory = NormalizeKey(keys.PreviousCategory)
}

func CommandKey(cmd Command) string {
	return strings.ToLower(strings.TrimSpace(cmd.Action)) + ":" + strings.TrimSpace(cmd.Command)
}

func CommandTitleKey(cmd Command) string {
	return strings.ToLower(strings.TrimSpace(cmd.Action)) + ":title:" + strings.ToLower(strings.TrimSpace(cmd.Title))
}

func rejectDeprecatedFields(meta toml.MetaData) error {
	for _, key := range meta.Undecoded() {
		if len(key) == 2 && key[0] == "ui" && key[1] == "tmux_mode_key" {
			return errors.New("ui.tmux_mode_key is no longer supported; use keys.tmux_mode")
		}
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
		if strings.EqualFold(strings.TrimSpace(cmd.Category), settingsCategory) {
			return fmt.Errorf("command %q uses reserved category %q", cmd.Title, settingsCategory)
		}
		cmd.Action = strings.ToLower(strings.TrimSpace(cmd.Action))
		if cmd.Action == "" {
			return fmt.Errorf("command %q must define action", cmd.Title)
		}
		if strings.TrimSpace(cmd.Command) == "" {
			return fmt.Errorf("command %q must define command", cmd.Title)
		}
		switch cmd.Action {
		case "tmux", "shell", "popup", "current_shell":
		default:
			return fmt.Errorf("command %q has unsupported action %q", cmd.Title, cmd.Action)
		}
		cmd.Prompt = strings.ToLower(strings.TrimSpace(cmd.Prompt))
		if cmd.Prompt != "" && !IsPromptName(cmd.Prompt) {
			return fmt.Errorf("command %q has unsupported prompt %q", cmd.Title, cmd.Prompt)
		}
		if cmd.Prompt != "" && !strings.Contains(cmd.Command, "{{input}}") && !strings.Contains(cmd.Command, "{{raw_input}}") {
			return fmt.Errorf("command %q defines prompt %q but command does not contain {{input}} or {{raw_input}}", cmd.Title, cmd.Prompt)
		}
	}
	return nil
}

func IsPromptName(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "session_name", "window_name", "target_index", "file_path", "command", "search_query":
		return true
	default:
		return false
	}
}

func ensureInternalCommands(commands []Command) []Command {
	userCommands := make([]Command, 0, len(commands))
	for _, cmd := range commands {
		if isSettingsInternal(cmd.Internal) {
			continue
		}
		userCommands = append(userCommands, cmd)
	}
	commands = userCommands
	commands = append(commands, SettingsCommands()...)
	return commands
}

func isSettingsInternal(internal string) bool {
	if internal == "" {
		return false
	}
	for _, cmd := range SettingsCommands() {
		if cmd.Internal == internal {
			return true
		}
	}
	return false
}
