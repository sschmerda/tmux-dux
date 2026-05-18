package tmux

import "testing"

func TestPopupCommand(t *testing.T) {
	got := PopupCommand("tmux-commander", "75%", "70%")
	want := "tmux display-popup -E -w 75% -h 70% tmux-commander"
	if got != want {
		t.Fatalf("PopupCommand = %q, want %q", got, want)
	}
}
