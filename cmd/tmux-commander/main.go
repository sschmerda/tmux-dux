package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stefanschmerda/tmux-commander/internal/actions"
	"github.com/stefanschmerda/tmux-commander/internal/config"
	"github.com/stefanschmerda/tmux-commander/internal/history"
	"github.com/stefanschmerda/tmux-commander/internal/palette"
	"github.com/stefanschmerda/tmux-commander/internal/theme"
	"github.com/stefanschmerda/tmux-commander/internal/tmux"
	"github.com/stefanschmerda/tmux-commander/internal/tmuxcmd"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "themes" {
		fmt.Println(strings.Join(theme.ConfigNames(), "\n"))
		return
	}

	if len(os.Args) > 1 {
		cfg, _, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "tmux-commander: load config: %v\n", err)
			os.Exit(1)
		}
		activeTheme := theme.ResolveWithCustom(cfg.UI.Theme, cfg.CustomTheme)
		switch os.Args[1] {
		case "popup":
			openConfiguredPopup(cfg, activeTheme)
			return
		default:
			fmt.Fprintf(os.Stderr, "tmux-commander: unknown command %q\n", os.Args[1])
			os.Exit(1)
		}
	}

	if err := runPalette(); err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: %v\n", err)
		os.Exit(1)
	}
}

func runPalette() error {
	state := palette.State{}
	for {
		reload, nextState, err := runPaletteOnce(state)
		if err != nil {
			return err
		}
		if !reload {
			return nil
		}
		state = nextState
	}
}

func runPaletteOnce(state palette.State) (bool, palette.State, error) {
	cfg, cfgPath, err := config.Load()
	if err != nil {
		return false, state, fmt.Errorf("load config: %w", err)
	}

	activeTheme := theme.ResolveWithCustom(cfg.UI.Theme, cfg.CustomTheme)
	recentHistory, recentPath := loadRecentHistory(cfg)
	if recentPath == "" {
		if path, err := history.Path(); err == nil {
			recentPath = path
		}
	}

	result, err := palette.RunWithState(
		cfg.Commands,
		activeTheme,
		previewThemes(activeTheme),
		cfg.UI.Glyphs,
		cfg.UI.ShowDescription,
		cfg.UI.ShowToggleHint,
		cfg.UI.TmuxDescription,
		recentHistory.RecentKeys(cfg.UI.RecentLimit),
		cfg.UI.TmuxModeKey,
		tmuxcmd.Load(),
		recentHistory.RecentTmuxInvocations(cfg.UI.TmuxRecentLimit),
		cfgPath,
		recentPath,
		state,
	)
	if err != nil {
		return false, state, fmt.Errorf("run palette: %w", err)
	}
	if result.Tmux != nil {
		if cfg.UI.RecentCommands && cfg.UI.TmuxRecentLimit > 0 && recentPath != "" {
			_, err := history.RecordTmuxWithLimits(recentPath, recentHistory, *result.Tmux, cfg.UI.RecentLimit, cfg.UI.TmuxRecentLimit, time.Now().UTC())
			if err != nil {
				fmt.Fprintf(os.Stderr, "tmux-commander: update tmux history: %v\n", err)
			}
		}
		action := actions.BuildTmuxCommand(result.Tmux.CommandLine())
		if err := actions.Dispatch(action); err != nil {
			return false, result.State, fmt.Errorf("dispatch tmux command: %w", err)
		}
		return false, result.State, nil
	}
	if result.Command == nil {
		return false, result.State, nil
	}
	if result.Command.Internal != "" {
		if result.Command.Internal == config.InternalReloadConfig {
			return true, result.State, nil
		}
		if err := runInternalCommand(*result.Command, cfgPath, activeTheme); err != nil {
			return false, result.State, err
		}
		return false, result.State, nil
	}
	if cfg.UI.RecentCommands && cfg.UI.RecentLimit > 0 && recentPath != "" && result.Command.Internal == "" {
		_, err := history.RecordWithLimits(recentPath, recentHistory, *result.Command, cfg.UI.RecentLimit, cfg.UI.TmuxRecentLimit, time.Now().UTC())
		if err != nil {
			fmt.Fprintf(os.Stderr, "tmux-commander: update history: %v\n", err)
		}
	}

	action, err := actions.Build(*result.Command, cfg.UI, result.Theme)
	if err != nil {
		return false, result.State, fmt.Errorf("build action: %w", err)
	}
	if err := actions.Dispatch(action); err != nil {
		return false, result.State, fmt.Errorf("dispatch action: %w", err)
	}
	return false, result.State, nil
}

func runInternalCommand(cmd config.Command, cfgPath string, activeTheme theme.Theme) error {
	switch cmd.Internal {
	case config.InternalEditConfig:
		return editConfig(cfgPath, activeTheme)
	default:
		return fmt.Errorf("unknown internal command %q", cmd.Internal)
	}
}

func editConfig(path string, activeTheme theme.Theme) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("create config file: %w", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("close config file: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("stat config file: %w", err)
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	args := []string{"tmux", "display-popup", "-E", "-b", "rounded", "-w", "80%", "-h", "80%"}
	if style := tmux.PopupStyle(activeTheme); style != "" {
		args = append(args, "-s", style)
	}
	if borderStyle := tmux.PopupBorderStyle(activeTheme); borderStyle != "" {
		args = append(args, "-S", borderStyle)
	}
	args = append(args, editor, path)
	return actions.Dispatch(actions.Action{
		Kind:    actions.KindTmux,
		Command: "tmux",
		Args:    []string{"run-shell", "-b", "sleep 0.05; " + shellJoin(args...)},
	})
}

func shellJoin(args ...string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, shellQuote(arg))
	}
	return strings.Join(quoted, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func loadRecentHistory(cfg config.Config) (history.File, string) {
	if !cfg.UI.RecentCommands || (cfg.UI.RecentLimit <= 0 && cfg.UI.TmuxRecentLimit <= 0) {
		return history.File{}, ""
	}
	file, path, err := history.LoadDefault()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tmux-commander: load history: %v\n", err)
		return history.File{}, path
	}
	trimmed := history.TrimWithLimits(file, cfg.UI.RecentLimit, cfg.UI.TmuxRecentLimit)
	if len(trimmed.Entries) != len(file.Entries) || len(trimmed.TmuxEntries) != len(file.TmuxEntries) {
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
