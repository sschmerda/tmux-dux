package palette

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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

func TestMatchesShowsRecentGroupAndKeepsNormalCategoryEntry(t *testing.T) {
	recent := config.Command{Title: "Lazygit", Category: "Tools", Action: "popup", Command: "lazygit"}
	other := config.Command{Title: "Split Horizontal", Category: "Panes", Action: "tmux", Command: "split-window -h"}
	model := New(
		[]config.Command{other, recent},
		theme.Resolve("shades-of-purple"),
		nil,
		true,
		true,
		[]string{config.CommandKey(recent)},
		"",
		"",
	)

	matches := model.matches()
	if len(matches) != 3 {
		t.Fatalf("match count = %d, want 3", len(matches))
	}
	if matches[0].Command.Title != "Lazygit" || matches[0].Command.Category != recentCategory {
		t.Fatalf("first match = %#v", matches[0].Command)
	}
	if matches[1].Command.Title != "Split Horizontal" {
		t.Fatalf("second match = %#v", matches[1].Command)
	}
	if matches[2].Command.Title != "Lazygit" || matches[2].Command.Category != "Tools" {
		t.Fatalf("third match = %#v", matches[2].Command)
	}
}

func TestMatchesAppliesRecentBoostWhenFiltering(t *testing.T) {
	recent := config.Command{Title: "Git Status", Category: "Tools", Action: "tmux", Command: "git-status"}
	other := config.Command{Title: "Git Stash", Category: "Tools", Action: "tmux", Command: "git-stash"}
	model := New(
		[]config.Command{other, recent},
		theme.Resolve("shades-of-purple"),
		nil,
		true,
		true,
		[]string{config.CommandKey(recent)},
		"",
		"",
	)
	model.query = "git st"

	matches := model.matches()
	if len(matches) != 2 {
		t.Fatalf("match count = %d, want 2", len(matches))
	}
	if matches[0].Command.Title != "Git Status" {
		t.Fatalf("first match = %#v", matches[0].Command)
	}
}

func TestConfigPathMessageShowsPath(t *testing.T) {
	model := New(
		nil,
		theme.Resolve("shades-of-purple"),
		nil,
		true,
		true,
		nil,
		"/tmp/tmux-commander/config.toml",
		"",
	)
	model.width = 80
	model.height = 24
	model.openMessage(config.InternalConfigPath)

	view := model.viewMessage()
	if !strings.Contains(view, "/tmp/tmux-commander/config.toml") {
		t.Fatalf("view did not include config path: %q", view)
	}
}

func TestMessageLineHighlightsPaths(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.width = 80
	model.height = 24

	path := model.renderMessageLine("/tmp/tmux-commander/config.toml")
	plain := model.renderMessageLine("Recent command history cleared:")
	if path == plain {
		t.Fatal("path and plain message rendered identically")
	}
}

func TestOpenMessagePreservesCursorAndQuery(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "/tmp/config.toml", "/tmp/history.toml")
	model.cursor = 3
	model.offset = 2
	model.query = "config"

	model.openMessage(config.InternalConfigPath)
	if model.cursor != 3 || model.offset != 2 || model.query != "config" {
		t.Fatalf("state = cursor %d offset %d query %q", model.cursor, model.offset, model.query)
	}
}

func TestOpenThemePreviewPreservesCursorAndQuery(t *testing.T) {
	model := New(
		[]config.Command{{Title: "Preview Themes", Internal: config.InternalThemePreview}},
		theme.Resolve("shades-of-purple"),
		[]theme.Theme{theme.Resolve("shades-of-purple")},
		true,
		true,
		nil,
		"",
		"",
	)
	model.cursor = 0
	model.offset = 2
	model.query = "themes"

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if updated.cursor != 0 || updated.offset != 2 || updated.query != "themes" {
		t.Fatalf("state = cursor %d offset %d query %q", updated.cursor, updated.offset, updated.query)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated = next.(Model)
	if updated.mode != modeCommands {
		t.Fatalf("mode = %v, want commands", updated.mode)
	}
	if updated.cursor != 0 || updated.offset != 2 || updated.query != "themes" {
		t.Fatalf("state after return = cursor %d offset %d query %q", updated.cursor, updated.offset, updated.query)
	}
}

func TestApplyStatePreservesReloadPosition(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.applyState(State{Query: "reload", Cursor: 4, Offset: 3})

	if got := model.state(); got.Query != "reload" || got.Cursor != 4 || got.Offset != 3 {
		t.Fatalf("state = %#v", got)
	}
}

func TestPopupInternalCommandExitsPalette(t *testing.T) {
	model := New(
		[]config.Command{{Title: "Open Config", Internal: config.InternalEditConfig}},
		theme.Resolve("shades-of-purple"),
		[]theme.Theme{theme.Resolve("shades-of-purple")},
		true,
		true,
		nil,
		"",
		"",
	)

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if updated.selected == nil {
		t.Fatal("expected popup-style internal command to be selected")
	}
	if updated.selected.Internal != config.InternalEditConfig {
		t.Fatalf("selected internal = %q, want %q", updated.selected.Internal, config.InternalEditConfig)
	}
	if cmd == nil {
		t.Fatal("expected popup-style internal command to quit the palette")
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
