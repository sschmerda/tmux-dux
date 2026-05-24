package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/stefanschmerda/tmux-commander/internal/actions"
	"github.com/stefanschmerda/tmux-commander/internal/config"
	"github.com/stefanschmerda/tmux-commander/internal/history"
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
	recentHistory, recentPath := loadRecentHistory(cfg)

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

	result, err := palette.Run(cfg.Commands, activeTheme, previewThemes(activeTheme), cfg.UI.Glyphs, cfg.UI.ShowDescription, recentHistory.RecentKeys(cfg.UI.RecentLimit))
	if err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: run palette: %v\n", err)
		os.Exit(1)
	}
	if result.Command == nil {
		return
	}
	if cfg.UI.RecentCommands && cfg.UI.RecentLimit > 0 && recentPath != "" && result.Command.Internal == "" {
		_, err := history.Record(recentPath, recentHistory, *result.Command, cfg.UI.RecentLimit, time.Now().UTC())
		if err != nil {
			fmt.Fprintf(os.Stderr, "tmux-commander: update history: %v\n", err)
		}
	}

	action, err := actions.Build(*result.Command, cfg.UI, result.Theme)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: build action: %v\n", err)
		os.Exit(1)
	}
	if err := actions.Dispatch(action); err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: dispatch action: %v\n", err)
		os.Exit(1)
	}
}

func loadRecentHistory(cfg config.Config) (history.File, string) {
	if !cfg.UI.RecentCommands || cfg.UI.RecentLimit <= 0 {
		return history.File{}, ""
	}
	file, path, err := history.LoadDefault()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: load history: %v\n", err)
		return history.File{}, path
	}
	trimmed := history.Trim(file, cfg.UI.RecentLimit)
	if len(trimmed.Entries) != len(file.Entries) {
		if err := history.Save(path, trimmed); err != nil {
			fmt.Fprintf(os.Stderr, "tmux-commander: trim history: %v\n", err)
		}
	}
	file = trimmed
	return file, path
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
