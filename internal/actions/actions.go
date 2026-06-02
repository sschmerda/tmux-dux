package actions

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sschmerda/tmux-dux/internal/config"
	"github.com/sschmerda/tmux-dux/internal/theme"
	"github.com/sschmerda/tmux-dux/internal/tmux"
)

type Kind string

const (
	KindTmux         Kind = "tmux"
	KindShell        Kind = "shell"
	KindPopup        Kind = "popup"
	KindCurrentShell Kind = "current_shell"
)

type Action struct {
	Kind    Kind
	Command string
	Args    []string
}

func Build(cmd config.Command, ui config.UI, t theme.Theme) (Action, error) {
	action := Kind(strings.ToLower(strings.TrimSpace(cmd.Action)))
	command := strings.TrimSpace(cmd.Command)
	if action == "" || command == "" {
		return Action{}, errors.New("command must define action and command")
	}

	switch action {
	case KindTmux:
		return deferredTmuxAction(KindTmux, "tmux "+command), nil
	case KindShell:
		return Action{Kind: KindShell, Command: shellPath(), Args: []string{"-lc", command}}, nil
	case KindCurrentShell:
		return currentShellAction(command), nil
	case KindPopup:
		args := []string{"tmux", "display-popup", "-E", "-b", "rounded", "-d", "#{pane_current_path}"}
		if style := tmux.PopupStyle(t); style != "" {
			args = append(args, "-s", style)
		}
		if borderStyle := tmux.PopupBorderStyle(t); borderStyle != "" {
			args = append(args, "-S", borderStyle)
		}
		if width := popupWidth(cmd, ui); width != "" {
			args = append(args, "-w", width)
		}
		if height := popupHeight(cmd, ui); height != "" {
			args = append(args, "-h", height)
		}
		args = append(args, command)
		return deferredTmuxAction(KindPopup, shellJoin(args...)), nil
	default:
		return Action{}, fmt.Errorf("unsupported action %q", cmd.Action)
	}
}

func BuildTmuxCommand(command string) Action {
	return deferredTmuxAction(KindTmux, "tmux "+strings.TrimSpace(command))
}

func Dispatch(action Action) error {
	cmd := exec.Command(action.Command, action.Args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func shellPath() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/sh"
}

func popupWidth(cmd config.Command, ui config.UI) string {
	if cmd.PopupWidth != "" {
		return cmd.PopupWidth
	}
	return ui.PopupWidth
}

func popupHeight(cmd config.Command, ui config.UI) string {
	if cmd.PopupHeight != "" {
		return cmd.PopupHeight
	}
	return ui.PopupHeight
}

func deferredTmuxAction(kind Kind, command string) Action {
	return Action{
		Kind:    kind,
		Command: "tmux",
		Args:    []string{"run-shell", "-b", "sleep 0.05; " + command},
	}
}

func currentShellAction(command string) Action {
	return deferredTmuxAction(
		KindCurrentShell,
		"tmux send-keys -l -- "+shellQuote(command)+" \\; send-keys Enter",
	)
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
