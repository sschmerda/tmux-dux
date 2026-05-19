package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/stefanschmerda/tmux-commander/internal/actions"
	"github.com/stefanschmerda/tmux-commander/internal/config"
	"github.com/stefanschmerda/tmux-commander/internal/palette"
	"github.com/stefanschmerda/tmux-commander/internal/theme"
	"github.com/stefanschmerda/tmux-commander/internal/tmux"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "themes" {
		fmt.Println(strings.Join(theme.ConfigNames(), "\n"))
		return
	}

	cfg, _, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: load config: %v\n", err)
		os.Exit(1)
	}

	activeTheme := theme.ResolveWithCustom(cfg.UI.Theme, cfg.CustomTheme)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "popup":
			openConfiguredPopup(cfg, activeTheme)
			return
		default:
			fmt.Fprintf(os.Stderr, "tmux-commander: unknown command %q\n", os.Args[1])
			os.Exit(1)
		}
	}

	selected, err := palette.Run(cfg.Commands, activeTheme, previewThemes(activeTheme), cfg.UI.Glyphs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: run palette: %v\n", err)
		os.Exit(1)
	}
	if selected == nil {
		return
	}

	action, err := actions.Build(*selected, cfg.UI, activeTheme)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: build action: %v\n", err)
		os.Exit(1)
	}
	if err := actions.Dispatch(action); err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: dispatch action: %v\n", err)
		os.Exit(1)
	}
}

func previewThemes(active theme.Theme) []theme.Theme {
	themes := make([]theme.Theme, 0, len(theme.Names())+1)
	for _, name := range theme.Names() {
		themes = append(themes, theme.Resolve(name))
	}
	if active.Name == "custom" {
		themes = append(themes, active)
	}
	return themes
}

func openConfiguredPopup(cfg config.Config, activeTheme theme.Theme) {
	binary, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: find executable: %v\n", err)
		os.Exit(1)
	}
	if err := tmux.OpenPopup(binary, cfg.UI.Width, cfg.UI.Height, false, activeTheme); err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: open popup: %v\n", err)
		os.Exit(1)
	}
}
