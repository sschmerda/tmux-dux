package fuzzy

import (
	"testing"

	"github.com/stefanschmerda/tmux-commander/internal/config"
)

func TestFilterMatchesMultipleTokens(t *testing.T) {
	commands := []config.Command{
		{Title: "Split Horizontal", Category: "Panes", Aliases: []string{"sh"}},
		{Title: "New Window", Category: "Windows", Aliases: []string{"nw"}},
	}
	matches := Filter(commands, "split pane")
	if len(matches) != 1 {
		t.Fatalf("match count = %d, want 1", len(matches))
	}
	if matches[0].Command.Title != "Split Horizontal" {
		t.Fatalf("match = %q", matches[0].Command.Title)
	}
}

func TestFilterMatchesAliasesAndInitials(t *testing.T) {
	commands := []config.Command{
		{Title: "Split Horizontal", Category: "Panes", Aliases: []string{"side"}},
	}
	if got := Filter(commands, "sh"); len(got) != 1 {
		t.Fatalf("initial match count = %d, want 1", len(got))
	}
	if got := Filter(commands, "side"); len(got) != 1 {
		t.Fatalf("alias match count = %d, want 1", len(got))
	}
}

func TestFilterReturnsAllWhenQueryEmpty(t *testing.T) {
	commands := []config.Command{
		{Title: "A"},
		{Title: "B"},
	}
	matches := Filter(commands, "")
	if len(matches) != len(commands) {
		t.Fatalf("match count = %d, want %d", len(matches), len(commands))
	}
}
