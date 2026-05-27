package actions

import (
	"strings"
	"testing"

	"github.com/stefanschmerda/tmux-commander/internal/config"
	"github.com/stefanschmerda/tmux-commander/internal/theme"
)

var testTheme = theme.Resolve("shades-of-purple")

func TestBuildTmuxAction(t *testing.T) {
	action, err := Build(config.Command{Action: "tmux", Command: "split-window -h"}, config.DefaultUI(), testTheme)
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

func TestBuildTmuxCommand(t *testing.T) {
	action := BuildTmuxCommand("split-window -h")
	if action.Kind != KindTmux || action.Command != "tmux" {
		t.Fatalf("unexpected action: %#v", action)
	}
	if len(action.Args) != 3 || action.Args[2] != "sleep 0.05; tmux split-window -h" {
		t.Fatalf("unexpected args: %#v", action.Args)
	}
}

func TestBuildShellAction(t *testing.T) {
	action, err := Build(config.Command{Action: "shell", Command: "echo hi"}, config.DefaultUI(), testTheme)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if action.Kind != KindShell || strings.Join(action.Args, " ") != "-lc echo hi" {
		t.Fatalf("unexpected action: %#v", action)
	}
}

func TestBuildPopupAction(t *testing.T) {
	ui := config.UI{PopupWidth: "90%", PopupHeight: "50%"}
	action, err := Build(config.Command{Action: "popup", Command: "lazygit"}, ui, testTheme)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if action.Kind != KindPopup || action.Command != "tmux" {
		t.Fatalf("unexpected action: %#v", action)
	}
	want := []string{"run-shell", "-b", "sleep 0.05; 'tmux' 'display-popup' '-E' '-b' 'rounded' '-d' '#{pane_current_path}' '-s' 'fg=#ffffff,bg=#2d2b55' '-S' 'fg=#d7d3ff,bg=#2d2b55' '-w' '90%' '-h' '50%' 'lazygit'"}
	if strings.Join(action.Args, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("args = %#v, want %#v", action.Args, want)
	}
}

func TestBuildPopupActionUsesCommandSizeOverride(t *testing.T) {
	ui := config.UI{PopupWidth: "90%", PopupHeight: "50%"}
	cmd := config.Command{Action: "popup", Command: "lazygit", PopupWidth: "95%", PopupHeight: "85%"}
	action, err := Build(cmd, ui, testTheme)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	want := []string{"run-shell", "-b", "sleep 0.05; 'tmux' 'display-popup' '-E' '-b' 'rounded' '-d' '#{pane_current_path}' '-s' 'fg=#ffffff,bg=#2d2b55' '-S' 'fg=#d7d3ff,bg=#2d2b55' '-w' '95%' '-h' '85%' 'lazygit'"}
	if strings.Join(action.Args, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("args = %#v, want %#v", action.Args, want)
	}
}

func TestBuildPopupActionCanOverrideOnlyOneSize(t *testing.T) {
	ui := config.UI{PopupWidth: "90%", PopupHeight: "50%"}
	cmd := config.Command{Action: "popup", Command: "lazygit", PopupHeight: "85%"}
	action, err := Build(cmd, ui, testTheme)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	want := []string{"run-shell", "-b", "sleep 0.05; 'tmux' 'display-popup' '-E' '-b' 'rounded' '-d' '#{pane_current_path}' '-s' 'fg=#ffffff,bg=#2d2b55' '-S' 'fg=#d7d3ff,bg=#2d2b55' '-w' '90%' '-h' '85%' 'lazygit'"}
	if strings.Join(action.Args, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("args = %#v, want %#v", action.Args, want)
	}
}

func TestBuildPopupActionQuotesShellCommand(t *testing.T) {
	action, err := Build(config.Command{Action: "popup", Command: "echo 'hi'"}, config.DefaultUI(), testTheme)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if !strings.Contains(action.Args[2], "'echo '\\''hi'\\'''") {
		t.Fatalf("popup command was not shell quoted: %#v", action.Args)
	}
}

func TestBuildRejectsMissingAction(t *testing.T) {
	if _, err := Build(config.Command{}, config.DefaultUI(), testTheme); err == nil {
		t.Fatal("Build returned nil error")
	}
}

func TestBuildRejectsUnsupportedAction(t *testing.T) {
	if _, err := Build(config.Command{Action: "editor", Command: "vim"}, config.DefaultUI(), testTheme); err == nil {
		t.Fatal("Build returned nil error")
	}
}
