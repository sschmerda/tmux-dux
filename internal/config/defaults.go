package config

const InternalThemePreview = "theme-preview"

func DefaultCommands() []Command {
	return []Command{
		{
			Title:       "Find Pane",
			Description: "Select a pane interactively",
			Category:    "Panes",
			Aliases:     []string{"fp"},
			Icon:        "󰓩",
			Tmux:        "display-panes",
		},
		{
			Title:       "Split Horizontal",
			Description: "Split pane side by side",
			Category:    "Panes",
			Aliases:     []string{"sh"},
			Icon:        "",
			Tmux:        "split-window -h -c '#{pane_current_path}'",
		},
		{
			Title:       "Split Vertical",
			Description: "Split pane top and bottom",
			Category:    "Panes",
			Aliases:     []string{"sv"},
			Icon:        "",
			Tmux:        "split-window -v -c '#{pane_current_path}'",
		},
		{
			Title:       "Close Pane",
			Description: "Kill the current pane",
			Category:    "Panes",
			Aliases:     []string{"kp"},
			Icon:        "󰅖",
			Tmux:        "kill-pane",
		},
		{
			Title:       "Zoom / Unzoom",
			Description: "Toggle pane zoom",
			Category:    "Panes",
			Aliases:     []string{"z"},
			Icon:        "󰍉",
			Tmux:        "resize-pane -Z",
		},
		{
			Title:       "New Window",
			Description: "Create a new window in the current path",
			Category:    "Windows",
			Aliases:     []string{"nw"},
			Icon:        "󰖯",
			Tmux:        "new-window -c '#{pane_current_path}'",
		},
		{
			Title:       "Rename Window",
			Description: "Prompt for a new window name",
			Category:    "Windows",
			Aliases:     []string{"rw"},
			Icon:        "󰑕",
			Tmux:        "command-prompt -I '#W' 'rename-window -- %1'",
		},
		{
			Title:       "Close Window",
			Description: "Kill the current window",
			Category:    "Windows",
			Aliases:     []string{"kw"},
			Icon:        "󰖭",
			Tmux:        "kill-window",
		},
		{
			Title:       "Choose Session",
			Description: "Open tmux session chooser",
			Category:    "Sessions",
			Aliases:     []string{"cs"},
			Icon:        "󱂬",
			Tmux:        "choose-tree -s",
		},
		{
			Title:       "New Session",
			Description: "Prompt for a new detached session",
			Category:    "Sessions",
			Aliases:     []string{"ns"},
			Icon:        "󰆧",
			Tmux:        "command-prompt -p 'session name' 'new-session -d -s %1'",
		},
		{
			Title:       "Rename Session",
			Description: "Prompt for a new session name",
			Category:    "Sessions",
			Aliases:     []string{"rs"},
			Icon:        "󰑕",
			Tmux:        "command-prompt -I '#S' 'rename-session -- %1'",
		},
		{
			Title:       "Detach",
			Description: "Detach the current tmux client",
			Category:    "Sessions",
			Aliases:     []string{"d"},
			Icon:        "󰍃",
			Tmux:        "detach-client",
		},
		{
			Title:       "Reload Config",
			Description: "Reload ~/.tmux.conf",
			Category:    "Tmux",
			Aliases:     []string{"rc"},
			Icon:        "󰑐",
			Tmux:        "source-file ~/.tmux.conf \\; display-message 'tmux config reloaded'",
		},
		ThemePreviewCommand(),
		{
			Title:       "Lazygit",
			Description: "Open lazygit in a popup",
			Category:    "Tools",
			Aliases:     []string{"lg"},
			Icon:        "󰊢",
			Popup:       "lazygit",
		},
		{
			Title:       "Btop",
			Description: "Open btop in a popup",
			Category:    "Tools",
			Aliases:     []string{"bt"},
			Icon:        "󰘚",
			Popup:       "btop",
		},
	}
}

func ThemePreviewCommand() Command {
	return Command{
		Title:       "Preview Themes",
		Description: "Preview built-in palette themes",
		Category:    "Settings",
		Aliases:     []string{"themes"},
		Icon:        "󰔎",
		Internal:    InternalThemePreview,
	}
}
