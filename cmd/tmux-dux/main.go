package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sschmerda/tmux-dux/internal/actions"
	"github.com/sschmerda/tmux-dux/internal/config"
	"github.com/sschmerda/tmux-dux/internal/history"
	"github.com/sschmerda/tmux-dux/internal/palette"
	"github.com/sschmerda/tmux-dux/internal/theme"
	"github.com/sschmerda/tmux-dux/internal/tmux"
	"github.com/sschmerda/tmux-dux/internal/tmuxcmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "themes" {
			fmt.Println(strings.Join(theme.ConfigNames(), "\n"))
			return
		}
		if os.Args[1] == "version" || os.Args[1] == "--version" || os.Args[1] == "-v" {
			fmt.Printf("tmux-dux %s (%s, %s)\n", version, commit, date)
			return
		}
		if os.Args[1] == "config" && len(os.Args) > 2 && os.Args[2] == "init" {
			if err := configInit(os.Stdout, os.Args[3:]); err != nil {
				fmt.Fprintf(os.Stderr, "tmux-dux: config init: %v\n", err)
				os.Exit(1)
			}
			return
		}
		cfg, _, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "tmux-dux: load config: %v\n", err)
			os.Exit(1)
		}
		activeTheme := theme.ResolveWithCustom(cfg.UI.Theme, cfg.CustomTheme)
		switch os.Args[1] {
		case "popup":
			openConfiguredPopup(cfg, activeTheme)
			return
		default:
			fmt.Fprintf(os.Stderr, "tmux-dux: unknown command %q\n", os.Args[1])
			os.Exit(1)
		}
	}

	if err := runPalette(); err != nil {
		fmt.Fprintf(os.Stderr, "tmux-dux: %v\n", err)
		os.Exit(1)
	}
}

func configInit(out io.Writer, args []string) error {
	targets, err := configInitTargets(args)
	if err != nil {
		return err
	}
	configPath, err := config.Path()
	if err != nil {
		return err
	}
	commandsPath := config.CommandsPath(configPath)
	scriptsPath := filepath.Join(filepath.Dir(configPath), "scripts")
	paths := make([]string, 0, 3)
	if targets.config {
		paths = append(paths, configPath)
	}
	if targets.commands {
		paths = append(paths, commandsPath)
	}
	if targets.scriptDir {
		paths = append(paths, scriptsPath)
	}
	existing, err := existingConfigFiles(paths...)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return fmt.Errorf("refusing to create config files because %s already exists", strings.Join(existing, ", "))
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if targets.config {
		if err := writeEmptyConfigFile(out, configPath); err != nil {
			return err
		}
	}
	if targets.commands {
		if err := writeEmptyConfigFile(out, commandsPath); err != nil {
			return err
		}
	}
	if targets.scriptDir {
		if err := writeEmptyConfigDirectory(out, scriptsPath); err != nil {
			return err
		}
	}
	return nil
}

type configInitSelection struct {
	config    bool
	commands  bool
	scriptDir bool
}

func configInitTargets(args []string) (configInitSelection, error) {
	targets := configInitSelection{}
	for _, arg := range args {
		switch arg {
		case "--config":
			targets.config = true
		case "--commands":
			targets.commands = true
		case "--script_dir":
			targets.scriptDir = true
		default:
			return targets, fmt.Errorf("unknown config init flag %q", arg)
		}
	}
	if !targets.config && !targets.commands && !targets.scriptDir {
		targets.config = true
		targets.commands = true
		targets.scriptDir = true
	}
	return targets, nil
}

func existingConfigFiles(paths ...string) ([]string, error) {
	var existing []string
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existing = append(existing, path)
			continue
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("stat %s: %w", path, err)
		}
	}
	return existing, nil
}

func writeEmptyConfigFile(out io.Writer, path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	fmt.Fprintf(out, "created: %s\n", path)
	return nil
}

func writeEmptyConfigDirectory(out io.Writer, path string) error {
	if err := os.Mkdir(path, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	fmt.Fprintf(out, "created: %s\n", path)
	return nil
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
		cfg.Keys,
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
				fmt.Fprintf(os.Stderr, "tmux-dux: update tmux history: %v\n", err)
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
	historyCommand := result.Command
	if result.HistoryCommand != nil {
		historyCommand = result.HistoryCommand
	}
	if cfg.UI.RecentCommands && cfg.UI.RecentLimit > 0 && recentPath != "" && historyCommand.Internal == "" {
		_, err := history.RecordWithLimits(recentPath, recentHistory, *historyCommand, cfg.UI.RecentLimit, cfg.UI.TmuxRecentLimit, time.Now().UTC())
		if err != nil {
			fmt.Fprintf(os.Stderr, "tmux-dux: update history: %v\n", err)
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
		fmt.Fprintf(os.Stderr, "tmux-dux: load history: %v\n", err)
		return history.File{}, path
	}
	trimmed := history.TrimWithLimits(file, cfg.UI.RecentLimit, cfg.UI.TmuxRecentLimit)
	if len(trimmed.Entries) != len(file.Entries) || len(trimmed.TmuxEntries) != len(file.TmuxEntries) {
		if err := history.Save(path, trimmed); err != nil {
			fmt.Fprintf(os.Stderr, "tmux-dux: trim history: %v\n", err)
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
		fmt.Fprintf(os.Stderr, "tmux-dux: find executable: %v\n", err)
		os.Exit(1)
	}
	if err := tmux.OpenPopup(binary, cfg.UI.Width, cfg.UI.Height, false, activeTheme); err != nil {
		fmt.Fprintf(os.Stderr, "tmux-dux: open popup: %v\n", err)
		os.Exit(1)
	}
}
