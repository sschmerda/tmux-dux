package actions

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/stefanschmerda/tmux-commander/internal/config"
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

func Build(cmd config.Command, ui config.UI) (Action, error) {
	switch {
	case cmd.Tmux != "":
		return deferredTmuxAction(KindTmux, "tmux "+cmd.Tmux), nil
	case cmd.Shell != "":
		return Action{Kind: KindShell, Command: shellPath(), Args: []string{"-lc", cmd.Shell}}, nil
	case cmd.Popup != "":
		args := []string{"tmux", "display-popup", "-E"}
		if ui.PopupWidth != "" {
			args = append(args, "-w", ui.PopupWidth)
		}
		if ui.PopupHeight != "" {
			args = append(args, "-h", ui.PopupHeight)
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
