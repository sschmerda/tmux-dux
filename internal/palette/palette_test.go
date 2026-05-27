package palette

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stefanschmerda/tmux-commander/internal/config"
	"github.com/stefanschmerda/tmux-commander/internal/fuzzy"
	"github.com/stefanschmerda/tmux-commander/internal/theme"
	"github.com/stefanschmerda/tmux-commander/internal/tmuxcmd"
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

func TestControlsMessageShowsConfiguredToggleKey(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.tmuxModeKey = "ctrl+y"
	model.openMessage(config.InternalControls)

	if model.messageTitle != "Controls" {
		t.Fatalf("message title = %q", model.messageTitle)
	}
	if !strings.Contains(model.messageBody, "Ctrl-Y") {
		t.Fatalf("controls did not include configured key: %q", model.messageBody)
	}
	if !strings.Contains(model.messageBody, "tmux Command Arguments") {
		t.Fatalf("controls did not include tmux argument section: %q", model.messageBody)
	}
	if !strings.Contains(model.messageBody, "Global\n────────\n") {
		t.Fatalf("controls did not include section separator: %q", model.messageBody)
	}
}

func TestControlsCategoriesAreDetected(t *testing.T) {
	if !isControlsCategory("Global") || !isControlsCategory("tmux Command Arguments") {
		t.Fatal("expected controls category")
	}
	if isControlsCategory("Enter              Select focused entry") {
		t.Fatal("hotkey row detected as category")
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

func TestCtrlTTogglesTmuxCommandMode(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.tmuxCommands = []tmuxcmd.Command{{Name: "split-window", Usage: "[-h]", TakesArgs: true}}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	updated := next.(Model)
	if updated.listMode != listModeTmux {
		t.Fatalf("listMode = %v, want tmux", updated.listMode)
	}
	if len(updated.tmuxMatches()) != 1 {
		t.Fatalf("tmux match count = %d, want 1", len(updated.tmuxMatches()))
	}
}

func TestConfiguredKeyTogglesTmuxCommandMode(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.tmuxModeKey = "ctrl+y"
	model.tmuxCommands = []tmuxcmd.Command{{Name: "split-window", Usage: "[-h]", TakesArgs: true}}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlY})
	updated := next.(Model)
	if updated.listMode != listModeTmux {
		t.Fatalf("listMode = %v, want tmux", updated.listMode)
	}
}

func TestRenderModeHintShowsConfiguredKey(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.tmuxModeKey = "ctrl+y"

	hint := model.renderModeHint()
	if !strings.Contains(hint, "Mode: Commands") || !strings.Contains(hint, "Ctrl-Y to toggle") {
		t.Fatalf("hint = %q", hint)
	}
}

func TestViewCanHideToggleHint(t *testing.T) {
	model := New(
		[]config.Command{{Title: "Lazygit", Category: "Tools", Action: "popup", Command: "lazygit"}},
		theme.Resolve("shades-of-purple"),
		nil,
		true,
		true,
		nil,
		"",
		"",
	)
	model.width = 80
	model.height = 24
	model.showToggleHint = false

	view := model.View()
	if strings.Contains(view, "Mode:") {
		t.Fatalf("view included toggle hint: %q", view)
	}
}

func TestSelectingTmuxCommandWithArgsOpensArgumentInput(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.listMode = listModeTmux
	model.tmuxCommands = []tmuxcmd.Command{{Name: "split-window", Usage: "[-h]", TakesArgs: true}}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if updated.mode != modeTmuxArgs {
		t.Fatalf("mode = %v, want tmux args", updated.mode)
	}
	if updated.argCommand.Name != "split-window" {
		t.Fatalf("arg command = %#v", updated.argCommand)
	}
}

func TestTmuxArgumentInputReturnsInvocation(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.mode = modeTmuxArgs
	model.argCommand = tmuxcmd.Command{Name: "split-window", TakesArgs: true}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-h")})
	updated := next.(Model)
	next, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated = next.(Model)
	if updated.selectedTmux == nil {
		t.Fatal("expected tmux invocation")
	}
	if updated.selectedTmux.CommandLine() != "split-window -h" {
		t.Fatalf("invocation = %#v", updated.selectedTmux)
	}
	if cmd == nil {
		t.Fatal("expected argument input to quit the palette")
	}
}

func TestTmuxArgumentViewShowsArgumentHelp(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.mode = modeTmuxArgs
	model.width = 100
	model.height = 30
	model.argCommand = tmuxcmd.Command{
		Name:        "split-window",
		Usage:       "[-h] [-v]",
		Description: "Split a pane",
		ArgHelp:     []string{"-h: split left/right", "-v: split top/bottom"},
		TakesArgs:   true,
	}

	view := model.viewTmuxArgs()
	if !strings.Contains(view, "-h: split left/right") || !strings.Contains(view, "args") {
		t.Fatalf("view did not include argument help and prompt: %q", view)
	}
	descriptionIndex := strings.Index(view, "Split a pane")
	usageIndex := strings.Index(view, "[-h] [-v]")
	helpIndex := strings.Index(view, "-h: split left/right")
	promptIndex := strings.Index(view, "args")
	if descriptionIndex < 0 || usageIndex < 0 || helpIndex < 0 || promptIndex < 0 {
		t.Fatalf("view missing expected sections: %q", view)
	}
	if !(descriptionIndex < usageIndex && usageIndex < helpIndex && helpIndex < promptIndex) {
		t.Fatalf("unexpected section order: description=%d usage=%d help=%d prompt=%d", descriptionIndex, usageIndex, helpIndex, promptIndex)
	}
}

func TestRecentTmuxCommandsAppearBeforeCatalog(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.listMode = listModeTmux
	model.tmuxCommands = []tmuxcmd.Command{{Name: "split-window", Usage: "[-h]", TakesArgs: true}}
	model.recentTmux = []tmuxcmd.Invocation{{Name: "split-window", Args: "-h"}}

	matches := model.tmuxMatches()
	if len(matches) != 2 {
		t.Fatalf("match count = %d, want 2", len(matches))
	}
	if !matches[0].Recent || matches[0].Invocation.CommandLine() != "split-window -h" {
		t.Fatalf("first match = %#v", matches[0])
	}
	if matches[1].Recent {
		t.Fatalf("second match unexpectedly recent: %#v", matches[1])
	}
}

func TestTmuxCommandRowsUseDescriptionAndArgsChip(t *testing.T) {
	cmd := tmuxCommandToConfig(tmuxcmd.Command{
		Name:        "split-window",
		Usage:       "[-h] [-v]",
		Description: "Split a pane",
		TakesArgs:   true,
	}, tmuxCommandCategory)

	if cmd.Description != "Split a pane" {
		t.Fatalf("description = %q, want short description", cmd.Description)
	}
	if strings.Contains(cmd.Description, "[-h]") {
		t.Fatalf("description included usage: %q", cmd.Description)
	}
	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "args" {
		t.Fatalf("aliases = %#v, want args chip", cmd.Aliases)
	}
}

func TestTmuxCommandViewCanHideDescriptions(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.width = 100
	model.height = 24
	model.listMode = listModeTmux
	model.tmuxDescription = false
	model.tmuxCommands = []tmuxcmd.Command{{
		Name:        "split-window",
		Description: "Split a pane",
		TakesArgs:   true,
	}}

	view := model.viewTmuxCommands()
	if strings.Contains(view, "Split a pane") {
		t.Fatalf("view included tmux description: %q", view)
	}
	if !strings.Contains(view, "split-window") {
		t.Fatalf("view did not include command name: %q", view)
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
