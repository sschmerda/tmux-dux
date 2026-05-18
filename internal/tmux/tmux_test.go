package tmux

import "testing"

func TestPopupCommand(t *testing.T) {
	got := PopupCommand("tmux-commander", "75%", "70%")
	want := "tmux display-popup -E -w 75% -h 70% tmux-commander"
	if got != want {
		t.Fatalf("PopupCommand = %q, want %q", got, want)
	}
}

func TestPopupArgs(t *testing.T) {
	got := PopupArgs("tmux-commander", "75%", "70%", true)
	want := []string{"display-popup", "-E", "-w", "75%", "-h", "70%", "tmux-commander"}
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
	got := PopupArgs("tmux-commander", "75%", "70%", false)
	want := []string{"display-popup", "-E", "-B", "-w", "75%", "-h", "70%", "tmux-commander"}
	if len(got) != len(want) {
		t.Fatalf("arg count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d = %q, want %q", i, got[i], want[i])
		}
	}
}
