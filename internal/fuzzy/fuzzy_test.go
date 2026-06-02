package fuzzy

import (
	"testing"

	"github.com/sschmerda/tmux-dux/internal/config"
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

func TestFilterReturnsTitleMatchIndexes(t *testing.T) {
	commands := []config.Command{
		{Title: "Lazygit", Category: "Tools", Aliases: []string{"lg"}},
	}
	matches := Filter(commands, "laz")
	if len(matches) != 1 {
		t.Fatalf("match count = %d, want 1", len(matches))
	}
	want := []int{0, 1, 2}
	if !equalIndexes(matches[0].TitleIndexes, want) {
		t.Fatalf("title indexes = %#v, want %#v", matches[0].TitleIndexes, want)
	}
}

func TestFilterReturnsAliasMatchIndexes(t *testing.T) {
	commands := []config.Command{
		{Title: "Open Git", Category: "Tools", Aliases: []string{"lg"}},
	}
	matches := Filter(commands, "lg")
	if len(matches) != 1 {
		t.Fatalf("match count = %d, want 1", len(matches))
	}
	want := []int{0, 1}
	if !equalIndexes(matches[0].AliasIndexes["lg"], want) {
		t.Fatalf("alias indexes = %#v, want %#v", matches[0].AliasIndexes["lg"], want)
	}
}

func TestFilterDoesNotMatchAcrossFields(t *testing.T) {
	commands := []config.Command{
		{Title: "Split Horizontal", Category: "Panes", Aliases: []string{"sh"}},
		{Title: "Btop", Category: "Tools", Aliases: []string{"bt"}},
	}
	matches := Filter(commands, "top")
	if len(matches) != 1 {
		t.Fatalf("match count = %d, want 1", len(matches))
	}
	if matches[0].Command.Title != "Btop" {
		t.Fatalf("match = %q, want Btop", matches[0].Command.Title)
	}
}

func TestFilterDoesNotMatchDescriptions(t *testing.T) {
	commands := []config.Command{
		{Title: "Lazygit", Description: "Open lazygit in a popup", Category: "Tools", Aliases: []string{"lg"}},
		{Title: "Btop", Description: "Open btop in a popup", Category: "Tools", Aliases: []string{"bt"}},
	}
	matches := Filter(commands, "popup")
	if len(matches) != 0 {
		t.Fatalf("match count = %d, want 0", len(matches))
	}
}

func TestFilterRanksTitleMatchAboveCategoryMatch(t *testing.T) {
	commands := []config.Command{
		{Title: "Tools", Category: "Misc"},
		{Title: "Btop", Category: "Tools", Aliases: []string{"bt"}},
	}
	matches := Filter(commands, "tools")
	if len(matches) != 2 {
		t.Fatalf("match count = %d, want 2", len(matches))
	}
	if matches[0].Command.Title != "Tools" {
		t.Fatalf("first match = %q, want Tools", matches[0].Command.Title)
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

func equalIndexes(got []int, want []int) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
