package palette

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sschmerda/tmux-commander/internal/config"
	"github.com/sschmerda/tmux-commander/internal/fuzzy"
	"github.com/sschmerda/tmux-commander/internal/theme"
	"github.com/sschmerda/tmux-commander/internal/tmuxcmd"
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

func TestScrollListMovesCursorAndOffset(t *testing.T) {
	model := New(commandList(20), theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.width = 80
	model.height = 12
	model.cursor = 4
	model.offset = 2

	model.scrollList(1)
	if model.cursor != 5 || model.offset <= 2 {
		t.Fatalf("after scroll down cursor=%d offset=%d, want cursor 5 and offset advanced", model.cursor, model.offset)
	}

	model.scrollList(-1)
	if model.cursor != 4 || model.offset >= 4 {
		t.Fatalf("after scroll up cursor=%d offset=%d, want cursor 4 and offset reduced", model.cursor, model.offset)
	}
}

func TestScrollViewportDoesNotMoveCursorAtEdges(t *testing.T) {
	model := New(commandList(20), theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.width = 80
	model.height = 12
	model.cursor = 0
	model.offset = 0

	model.scrollViewport(-1)
	if model.cursor != 0 || model.offset != 0 {
		t.Fatalf("top scroll cursor=%d offset=%d, want 0 0", model.cursor, model.offset)
	}

	model.cursor = 5
	model.offset = 2
	model.scrollViewport(1)
	if model.cursor != 5 || model.offset <= 2 {
		t.Fatalf("scroll down cursor=%d offset=%d, want cursor unchanged and offset advanced", model.cursor, model.offset)
	}
}

func TestScrollHalfPageHasMinimumOneLine(t *testing.T) {
	model := New(commandList(3), theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.width = 80
	model.height = 1

	if got := model.scrollHalfPage(); got != 1 {
		t.Fatalf("scrollHalfPage = %d, want 1", got)
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

func TestMatchesUsesTitleFallbackForRecentCommands(t *testing.T) {
	current := config.Command{Title: "New Session", Category: "Sessions", Action: "tmux", Command: "new-session -d -s {{input}}"}
	model := New(
		[]config.Command{current},
		theme.Resolve("shades-of-purple"),
		nil,
		true,
		true,
		[]string{config.CommandTitleKey(config.Command{Title: "New Session", Action: "tmux"})},
		"",
		"",
	)

	matches := model.matches()
	if len(matches) != 2 {
		t.Fatalf("match count = %d, want 2", len(matches))
	}
	if matches[0].Command.Title != "New Session" || matches[0].Command.Category != recentCategory {
		t.Fatalf("first match = %#v", matches[0].Command)
	}
	if matches[1].Command.Category != "Sessions" {
		t.Fatalf("second match = %#v", matches[1].Command)
	}
}

func TestMatchesDeduplicatesRecentFallbackKeys(t *testing.T) {
	current := config.Command{Title: "New Session", Category: "Sessions", Action: "tmux", Command: "new-session -d -s {{input}}"}
	model := New(
		[]config.Command{current},
		theme.Resolve("shades-of-purple"),
		nil,
		true,
		true,
		[]string{config.CommandKey(current), config.CommandTitleKey(current)},
		"",
		"",
	)

	matches := model.matches()
	if len(matches) != 2 {
		t.Fatalf("match count = %d, want recent plus normal command", len(matches))
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
	model.keys.TmuxMode = "ctrl+y"
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

func TestSelectCommandClampsCursorPastEnd(t *testing.T) {
	model := New(
		[]config.Command{
			{Title: "One", Action: "shell", Command: "one"},
			{Title: "Two", Action: "shell", Command: "two"},
		},
		theme.Resolve("shades-of-purple"),
		nil,
		true,
		true,
		nil,
		"",
		"",
	)
	model.cursor = 10
	model.offset = 10

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if updated.selected == nil {
		t.Fatal("expected selected command")
	}
	if updated.selected.Title != "Two" {
		t.Fatalf("selected = %q, want Two", updated.selected.Title)
	}
	if updated.cursor != 1 || updated.offset != 10 {
		t.Fatalf("cursor/offset = %d/%d, want 1/10", updated.cursor, updated.offset)
	}
	if cmd == nil {
		t.Fatal("expected palette quit command")
	}
}

func TestShowOutputCommandStaysInPalette(t *testing.T) {
	model := New(
		[]config.Command{{Title: "Show Date", Action: "shell", Command: "printf output", ShowOutput: true}},
		theme.Resolve("shades-of-purple"),
		nil,
		true,
		true,
		nil,
		"",
		"",
	)

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if updated.mode != modeMessage {
		t.Fatalf("mode = %v, want message", updated.mode)
	}
	if updated.selected != nil {
		t.Fatal("show output command should not select external action")
	}
	if cmd == nil {
		t.Fatal("expected output command")
	}
	msg := cmd()
	next, _ = updated.Update(msg)
	updated = next.(Model)
	if updated.messageTitle != "Show Date" || updated.messageBody != "output" {
		t.Fatalf("message = %q %q, want Show Date/output", updated.messageTitle, updated.messageBody)
	}
}

func TestSelectCommandClampsNegativeCursor(t *testing.T) {
	model := New(
		[]config.Command{
			{Title: "One", Action: "shell", Command: "one"},
			{Title: "Two", Action: "shell", Command: "two"},
		},
		theme.Resolve("shades-of-purple"),
		nil,
		true,
		true,
		nil,
		"",
		"",
	)
	model.cursor = -3
	model.offset = -2

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if updated.selected == nil {
		t.Fatal("expected selected command")
	}
	if updated.selected.Title != "One" {
		t.Fatalf("selected = %q, want One", updated.selected.Title)
	}
	if updated.cursor != 0 || updated.offset != 0 {
		t.Fatalf("cursor/offset = %d/%d, want 0/0", updated.cursor, updated.offset)
	}
}

func TestSelectTmuxCommandClampsCursorPastEnd(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.listMode = listModeTmux
	model.tmuxCommands = []tmuxcmd.Command{
		{Name: "display-message"},
		{Name: "list-sessions"},
	}
	model.cursor = 8
	model.offset = 8

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if updated.selectedTmux == nil {
		t.Fatal("expected selected tmux command")
	}
	if updated.selectedTmux.Name != "list-sessions" {
		t.Fatalf("selected tmux = %q, want list-sessions", updated.selectedTmux.Name)
	}
	if updated.cursor != 1 || updated.offset != 8 {
		t.Fatalf("cursor/offset = %d/%d, want 1/8", updated.cursor, updated.offset)
	}
	if cmd == nil {
		t.Fatal("expected palette quit command")
	}
}

func TestClearingSearchRestoresPreviousSelection(t *testing.T) {
	model := New(commandList(8), theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.width = 80
	model.height = 24
	model.cursor = 5
	model.offset = 3

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	updated := next.(Model)
	if updated.cursor != 0 || updated.offset != 0 {
		t.Fatalf("search cursor/offset = %d/%d, want 0/0", updated.cursor, updated.offset)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	updated = next.(Model)
	if updated.query != "" {
		t.Fatalf("query = %q, want empty", updated.query)
	}
	if updated.cursor != 5 || updated.offset != 3 {
		t.Fatalf("restored cursor/offset = %d/%d, want 5/3", updated.cursor, updated.offset)
	}
}

func TestEditingNonEmptySearchKeepsFilteredSelectionAtTop(t *testing.T) {
	model := New(commandList(8), theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.width = 80
	model.height = 24
	model.cursor = 5
	model.offset = 3

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("xy")})
	updated := next.(Model)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	updated = next.(Model)

	if updated.query != "x" {
		t.Fatalf("query = %q, want x", updated.query)
	}
	if updated.cursor != 0 || updated.offset != 0 {
		t.Fatalf("cursor/offset = %d/%d, want 0/0", updated.cursor, updated.offset)
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
	model.keys.TmuxMode = "ctrl+y"
	model.tmuxCommands = []tmuxcmd.Command{{Name: "split-window", Usage: "[-h]", TakesArgs: true}}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlY})
	updated := next.(Model)
	if updated.listMode != listModeTmux {
		t.Fatalf("listMode = %v, want tmux", updated.listMode)
	}
}

func TestRenderModeHintShowsConfiguredKey(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.keys.TmuxMode = "ctrl+y"

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

func TestSelectingPromptedCommandOpensCommandInput(t *testing.T) {
	model := New(
		[]config.Command{{Title: "New Session", Action: "tmux", Command: "new-session -d -s {{input}}", Prompt: "session_name"}},
		theme.Resolve("shades-of-purple"),
		nil,
		true,
		true,
		nil,
		"",
		"",
	)

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if updated.mode != modeCommandInput {
		t.Fatalf("mode = %v, want command input", updated.mode)
	}
	if updated.inputPrompt.Label != "session name" {
		t.Fatalf("prompt label = %q, want session name", updated.inputPrompt.Label)
	}
}

func TestCommandInputReturnsRenderedCommand(t *testing.T) {
	model := New(nil, theme.Resolve("shades-of-purple"), nil, true, true, nil, "", "")
	model.mode = modeCommandInput
	model.inputCommand = config.Command{Title: "New Session", Action: "tmux", Command: "new-session -d -s {{input}}", Prompt: "session_name"}
	model.inputPrompt, _ = commandPromptSpec("session_name")

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("work repo")})
	updated := next.(Model)
	next, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated = next.(Model)
	if updated.selected == nil {
		t.Fatal("expected selected command")
	}
	if updated.selected.Command != "new-session -d -s 'work repo'" {
		t.Fatalf("command = %q", updated.selected.Command)
	}
	if updated.selectedHistory == nil {
		t.Fatal("expected selected history command")
	}
	if updated.selectedHistory.Command != "new-session -d -s {{input}}" {
		t.Fatalf("history command = %q", updated.selectedHistory.Command)
	}
	if cmd == nil {
		t.Fatal("expected command input to quit the palette")
	}
}

func TestCommandInputCanRenderRawPlaceholder(t *testing.T) {
	got := renderCommandInput("find-window {{raw_input}}", "foo bar")
	if got != "find-window foo bar" {
		t.Fatalf("command = %q", got)
	}
}

func TestCountPromptRequiresDigits(t *testing.T) {
	spec, _ := commandPromptSpec("count")
	if !validPromptValue(spec, "3") {
		t.Fatal("count prompt rejected numeric input")
	}
	if validPromptValue(spec, "3; display-message bad") {
		t.Fatal("count prompt accepted non-numeric input")
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

func commandList(count int) []config.Command {
	commands := make([]config.Command, 0, count)
	for i := 0; i < count; i++ {
		commands = append(commands, config.Command{
			Title:    fmt.Sprintf("Command %02d", i),
			Category: "Test",
			Action:   "tmux",
			Command:  fmt.Sprintf("display-message %d", i),
		})
	}
	return commands
}
