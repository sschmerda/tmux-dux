package actions

import (
	"strings"
	"testing"

	"github.com/stefanschmerda/tmux-commander/internal/config"
)

func TestBuildTmuxAction(t *testing.T) {
	action, err := Build(config.Command{Tmux: "split-window -h"}, config.DefaultUI())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if action.Kind != KindTmux {
		t.Fatalf("kind = %q, want %q", action.Kind, KindTmux)
	}
	if action.Command != "tmux" {
		t.Fatalf("command = %q, want tmux", action.Command)
	}
	if len(action.Args) != 3 || action.Args[0] != "run-shell" || action.Args[1] != "-b" || action.Args[2] != "sleep 0.05; tmux split-window -h" {
		t.Fatalf("unexpected args: %#v", action.Args)
	}
}

func TestBuildShellAction(t *testing.T) {
	action, err := Build(config.Command{Shell: "echo hi"}, config.DefaultUI())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if action.Kind != KindShell || strings.Join(action.Args, " ") != "-lc echo hi" {
		t.Fatalf("unexpected action: %#v", action)
	}
}

func TestBuildPopupAction(t *testing.T) {
	ui := config.UI{PopupWidth: "90%", PopupHeight: "50%"}
	action, err := Build(config.Command{Popup: "lazygit"}, ui)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if action.Kind != KindPopup || action.Command != "tmux" {
		t.Fatalf("unexpected action: %#v", action)
	}
	want := []string{"run-shell", "-b", "sleep 0.05; 'tmux' 'display-popup' '-E' '-w' '90%' '-h' '50%' 'lazygit'"}
	if strings.Join(action.Args, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("args = %#v, want %#v", action.Args, want)
	}
}

func TestBuildPopupActionQuotesShellCommand(t *testing.T) {
	action, err := Build(config.Command{Popup: "echo 'hi'"}, config.DefaultUI())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if !strings.Contains(action.Args[2], "'echo '\\''hi'\\'''") {
		t.Fatalf("popup command was not shell quoted: %#v", action.Args)
	}
}

func TestBuildRejectsMissingAction(t *testing.T) {
	if _, err := Build(config.Command{}, config.DefaultUI()); err == nil {
		t.Fatal("Build returned nil error")
	}
}
