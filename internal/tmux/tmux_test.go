package tmux

import "testing"

import "github.com/stefanschmerda/tmux-commander/internal/theme"

func TestPopupCommand(t *testing.T) {
	got := PopupCommand("tmux-commander", "75%", "70%")
	want := "tmux display-popup -E -w 75% -h 70% tmux-commander"
	if got != want {
		t.Fatalf("PopupCommand = %q, want %q", got, want)
	}
}

func TestPopupArgs(t *testing.T) {
	got := PopupArgs("tmux-commander", "75%", "70%", true, theme.Resolve("shades-of-purple"))
	want := []string{"display-popup", "-E", "-s", "fg=#ffffff,bg=#2d2b55", "-S", "fg=#fad000", "-w", "75%", "-h", "70%", "tmux-commander"}
	if len(got) != len(want) {
		t.Fatalf("arg count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestPopupArgsWithoutBorder(t *testing.T) {
	got := PopupArgs("tmux-commander", "75%", "70%", false, theme.Resolve("shades-of-purple"))
	want := []string{"display-popup", "-E", "-B", "-s", "fg=#ffffff,bg=#2d2b55", "-S", "fg=#fad000", "-w", "75%", "-h", "70%", "tmux-commander"}
	if len(got) != len(want) {
		t.Fatalf("arg count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestPopupStyle(t *testing.T) {
	got := PopupStyle(theme.Theme{Query: "#eeeeee", Background: "#111111"})
	want := "fg=#eeeeee,bg=#111111"
	if got != want {
		t.Fatalf("PopupStyle = %q, want %q", got, want)
	}
}
