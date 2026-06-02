package config

const (
	InternalThemePreview = "theme-preview"
	InternalClearRecent  = "clear-recent"
	InternalConfigPath   = "config-path"
	InternalControls     = "controls"
	InternalEditConfig   = "edit-config"
	InternalReloadConfig = "reload-config"

	settingsCategory = "Settings tmux-dux"
)

func DefaultCommands() []Command {
	return SettingsCommands()
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
			Description: "Show tmux-dux controls and hotkeys",
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
			Description: "Reload tmux-dux config by reopening the palette",
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
