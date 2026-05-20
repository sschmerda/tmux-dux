package actions

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/stefanschmerda/tmux-commander/internal/config"
	"github.com/stefanschmerda/tmux-commander/internal/theme"
	"github.com/stefanschmerda/tmux-commander/internal/tmux"
)

type Kind string

const (
	KindTmux  Kind = "tmux"
	KindShell Kind = "shell"
	KindPopup Kind = "popup"
)

type Action struct {
	Kind    Kind
	Command string
	Args    []string
}

func Build(cmd config.Command, ui config.UI, t theme.Theme) (Action, error) {
	switch {
	case cmd.Tmux != "":
		return deferredTmuxAction(KindTmux, "tmux "+cmd.Tmux), nil
	case cmd.Shell != "":
		return Action{Kind: KindShell, Command: shellPath(), Args: []string{"-lc", cmd.Shell}}, nil
	case cmd.Popup != "":
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
		args = append(args, cmd.Popup)
		return deferredTmuxAction(KindPopup, shellJoin(args...)), nil
	default:
		return Action{}, errors.New("command has no action")
	}
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
