package palette

import (
	"strings"
	"testing"

	"github.com/stefanschmerda/tmux-commander/internal/config"
	"github.com/stefanschmerda/tmux-commander/internal/fuzzy"
	"github.com/stefanschmerda/tmux-commander/internal/theme"
)

func TestNextCategoryIndexMovesToFirstCommandInNextCategory(t *testing.T) {
	matches := categoryMatches("Panes", "Panes", "Windows", "Windows", "Sessions")

	if got := nextCategoryIndex(matches, 0); got != 2 {
		t.Fatalf("nextCategoryIndex = %d, want 2", got)
	}
	if got := nextCategoryIndex(matches, 2); got != 4 {
		t.Fatalf("nextCategoryIndex = %d, want 4", got)
	}
}

func TestNextCategoryIndexWrapsToFirstCategory(t *testing.T) {
	matches := categoryMatches("Panes", "Windows", "Sessions")

	if got := nextCategoryIndex(matches, 2); got != 0 {
		t.Fatalf("nextCategoryIndex = %d, want 0", got)
	}
}

func TestNextCategoryIndexDoesNotMoveWhenOnlyOneCategoryIsVisible(t *testing.T) {
	matches := categoryMatches("Panes", "Panes", "Panes")

	if got := nextCategoryIndex(matches, 1); got != 1 {
		t.Fatalf("nextCategoryIndex = %d, want 1", got)
	}
}

func TestPreviousCategoryIndexMovesToFirstCommandInPreviousCategory(t *testing.T) {
	matches := categoryMatches("Panes", "Panes", "Windows", "Windows", "Sessions")

	if got := previousCategoryIndex(matches, 4); got != 2 {
		t.Fatalf("previousCategoryIndex = %d, want 2", got)
	}
	if got := previousCategoryIndex(matches, 2); got != 0 {
		t.Fatalf("previousCategoryIndex = %d, want 0", got)
	}
}

func TestPreviousCategoryIndexWrapsToLastCategoryStart(t *testing.T) {
	matches := categoryMatches("Panes", "Panes", "Windows", "Sessions", "Sessions")

	if got := previousCategoryIndex(matches, 0); got != 3 {
		t.Fatalf("previousCategoryIndex = %d, want 3", got)
	}
}

func TestPreviousCategoryIndexDoesNotMoveWhenOnlyOneCategoryIsVisible(t *testing.T) {
	matches := categoryMatches("Panes", "Panes", "Panes")

	if got := previousCategoryIndex(matches, 1); got != 1 {
		t.Fatalf("previousCategoryIndex = %d, want 1", got)
	}
}

func TestRenderRowCanHideDescription(t *testing.T) {
	styles := newStyles(theme.Resolve("shades-of-purple"))
	match := fuzzy.Match{Command: config.Command{
		Title:       "Lazygit",
		Description: "Open lazygit in a popup",
	}}

	withDescription := renderRow(match, styles, false, true, true, 80)
	if !strings.Contains(withDescription, "Open lazygit in a popup") {
		t.Fatalf("row did not include description: %q", withDescription)
	}

	withoutDescription := renderRow(match, styles, false, true, false, 80)
	if strings.Contains(withoutDescription, "Open lazygit in a popup") {
		t.Fatalf("row included description: %q", withoutDescription)
	}
}

func categoryMatches(categories ...string) []fuzzy.Match {
	matches := make([]fuzzy.Match, 0, len(categories))
	for _, category := range categories {
		matches = append(matches, fuzzy.Match{
			Command: config.Command{
				Title:    category,
				Category: category,
			},
		})
	}
	return matches
}
