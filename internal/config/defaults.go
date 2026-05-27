package config

const (
	InternalThemePreview = "theme-preview"
	InternalClearRecent  = "clear-recent"
	InternalConfigPath   = "config-path"
	InternalControls     = "controls"
	InternalEditConfig   = "edit-config"
	InternalReloadConfig = "reload-config"

	settingsCategory = "Settings tmux-commander"
)

func DefaultCommands() []Command {
	commands := []Command{
		{
			Title:       "Find Pane",
			Description: "Select a pane interactively",
			Category:    "Panes",
			Aliases:     []string{"fp"},
			Icon:        "󰓩",
			Action:      "tmux",
			Command:     "display-panes",
		},
		{
			Title:       "Split Horizontal",
			Description: "Split pane side by side",
			Category:    "Panes",
			Aliases:     []string{"sh"},
			Icon:        "",
			Action:      "tmux",
			Command:     "split-window -h -c '#{pane_current_path}'",
		},
		{
			Title:       "Split Vertical",
			Description: "Split pane top and bottom",
			Category:    "Panes",
			Aliases:     []string{"sv"},
			Icon:        "",
			Action:      "tmux",
			Command:     "split-window -v -c '#{pane_current_path}'",
		},
		{
			Title:       "Close Pane",
			Description: "Kill the current pane",
			Category:    "Panes",
			Aliases:     []string{"kp"},
			Icon:        "󰅖",
			Action:      "tmux",
			Command:     "kill-pane",
		},
		{
			Title:       "Zoom / Unzoom",
			Description: "Toggle pane zoom",
			Category:    "Panes",
			Aliases:     []string{"z"},
			Icon:        "󰍉",
			Action:      "tmux",
			Command:     "resize-pane -Z",
		},
		{
			Title:       "New Window",
			Description: "Create a new window in the current path",
			Category:    "Windows",
			Aliases:     []string{"nw"},
			Icon:        "󰖯",
			Action:      "tmux",
			Command:     "new-window -c '#{pane_current_path}'",
		},
		{
			Title:       "Rename Window",
			Description: "Prompt for a new window name",
			Category:    "Windows",
			Aliases:     []string{"rw"},
			Icon:        "󰑕",
			Action:      "tmux",
			Command:     "command-prompt -I '#W' 'rename-window -- %1'",
		},
		{
			Title:       "Close Window",
			Description: "Kill the current window",
			Category:    "Windows",
			Aliases:     []string{"kw"},
			Icon:        "󰖭",
			Action:      "tmux",
			Command:     "kill-window",
		},
		{
			Title:       "Choose Session",
			Description: "Open tmux session chooser",
			Category:    "Sessions",
			Aliases:     []string{"cs"},
			Icon:        "󱂬",
			Action:      "tmux",
			Command:     "choose-tree -s",
		},
		{
			Title:       "New Session",
			Description: "Prompt for a new detached session",
			Category:    "Sessions",
			Aliases:     []string{"ns"},
			Icon:        "󰆧",
			Action:      "tmux",
			Command:     "command-prompt -p 'session name' 'new-session -d -s %1'",
		},
		{
			Title:       "Rename Session",
			Description: "Prompt for a new session name",
			Category:    "Sessions",
			Aliases:     []string{"rs"},
			Icon:        "󰑕",
			Action:      "tmux",
			Command:     "command-prompt -I '#S' 'rename-session -- %1'",
		},
		{
			Title:       "Detach",
			Description: "Detach the current tmux client",
			Category:    "Sessions",
			Aliases:     []string{"d"},
			Icon:        "󰍃",
			Action:      "tmux",
			Command:     "detach-client",
		},
	}
	commands = append(commands, SettingsCommands()...)
	commands = append(commands,
		Command{
			Title:       "Lazygit",
			Description: "Open lazygit in a popup",
			Category:    "Tools",
			Aliases:     []string{"lg"},
			Icon:        "󰊢",
			Action:      "popup",
			Command:     "lazygit",
		},
		Command{
			Title:       "Btop",
			Description: "Open btop in a popup",
			Category:    "Tools",
			Aliases:     []string{"bt"},
			Icon:        "󰘚",
			Action:      "popup",
			Command:     "btop",
		},
	)
	return commands
}

func SettingsCommands() []Command {
	return []Command{
		ThemePreviewCommand(),
		{
			Title:       "Clear Recent Commands",
			Description: "Delete the recent command history file",
			Category:    settingsCategory,
			Aliases:     []string{"cr"},
			Icon:        "󰆴",
			Internal:    InternalClearRecent,
		},
		{
			Title:       "List Config Path",
			Description: "Show the active TOML config path",
			Category:    settingsCategory,
			Aliases:     []string{"cp"},
			Icon:        "󰈙",
			Internal:    InternalConfigPath,
		},
		{
			Title:       "Show Controls",
			Description: "Show tmux-commander controls and hotkeys",
			Category:    settingsCategory,
			Aliases:     []string{"help"},
			Icon:        "󰋖",
			Internal:    InternalControls,
		},
		{
			Title:       "Open / Edit Config",
			Description: "Open the TOML config in $EDITOR",
			Category:    settingsCategory,
			Aliases:     []string{"ec"},
			Icon:        "󰏫",
			Internal:    InternalEditConfig,
		},
		{
			Title:       "Reload Config",
			Description: "Reload tmux-commander config by reopening the palette",
			Category:    settingsCategory,
			Aliases:     []string{"rc"},
			Icon:        "󰑐",
			Internal:    InternalReloadConfig,
		},
	}
}

func ThemePreviewCommand() Command {
	return Command{
		Title:       "Preview Themes",
		Description: "Preview built-in palette themes",
		Category:    settingsCategory,
		Aliases:     []string{"themes"},
		Icon:        "󰔎",
		Internal:    InternalThemePreview,
	}
}
