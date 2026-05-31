package palette

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sschmerda/tmux-commander/internal/config"
	"github.com/sschmerda/tmux-commander/internal/fuzzy"
	"github.com/sschmerda/tmux-commander/internal/theme"
	"github.com/sschmerda/tmux-commander/internal/tmuxcmd"
)

type Model struct {
	commands        []config.Command
	activeTheme     theme.Theme
	previewThemes   []theme.Theme
	previewIndex    int
	showGlyphs      bool
	showDescription bool
	showToggleHint  bool
	tmuxDescription bool
	recentKeys      []string
	keys            config.Keys
	tmuxCommands    []tmuxcmd.Command
	recentTmux      []tmuxcmd.Invocation
	configPath      string
	historyPath     string
	styles          styles
	mode            mode
	listMode        listMode
	query           string
	messageTitle    string
	messageBody     string
	messageStatus   string
	argCommand      tmuxcmd.Command
	argValue        string
	inputCommand    config.Command
	inputPrompt     promptSpec
	inputValue      string
	caretVisible    bool
	cursor          int
	offset          int
	preSearchActive bool
	preSearchCursor int
	preSearchOffset int
	selected        *config.Command
	selectedHistory *config.Command
	selectedTmux    *tmuxcmd.Invocation
	width           int
	height          int
}

type Result struct {
	Command        *config.Command
	HistoryCommand *config.Command
	Tmux           *tmuxcmd.Invocation
	Theme          theme.Theme
	State          State
}

type State struct {
	Query  string
	Cursor int
	Offset int
}

type mode int

const (
	modeCommands mode = iota
	modeThemePreview
	modeMessage
	modeTmuxArgs
	modeCommandInput
)

type listMode int

const (
	listModeCommands listMode = iota
	listModeTmux
)

const horizontalPadding = 3

const cursorBlinkInterval = 500 * time.Millisecond

const recentCategory = "Recent"

const tmuxCommandCategory = "tmux Commands"

const recentScoreBoost = 50

type cursorBlinkMsg struct{}

type commandOutputMsg struct {
	title string
	body  string
}

type messageCopiedMsg struct {
	err error
}

type tmuxMatch struct {
	Match      fuzzy.Match
	Command    tmuxcmd.Command
	Invocation tmuxcmd.Invocation
	Recent     bool
}

type promptSpec struct {
	Name        string
	Label       string
	Description string
	Placeholder string
}

type styles struct {
	root          lipgloss.Style
	frame         lipgloss.Style
	title         lipgloss.Style
	header        lipgloss.Style
	desc          lipgloss.Style
	prompt        lipgloss.Style
	query         lipgloss.Style
	empty         lipgloss.Style
	selected      lipgloss.Style
	selectedTitle lipgloss.Style
	selectedDesc  lipgloss.Style
	selectedChip  lipgloss.Style
	match         lipgloss.Style
	selectedMatch lipgloss.Style
	chipMatch     lipgloss.Style
	glyph         lipgloss.Style
	selectedGlyph lipgloss.Style
	chip          lipgloss.Style
	muted         lipgloss.Style
}

func newStyles(t theme.Theme) styles {
	return styles{
		root:          lipgloss.NewStyle().Background(lipgloss.Color(t.Background)),
		frame:         lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(t.CommanderBorder)).BorderBackground(lipgloss.Color(t.Background)).Background(lipgloss.Color(t.Background)),
		title:         lipgloss.NewStyle().Foreground(lipgloss.Color(t.Title)).Background(lipgloss.Color(t.Background)),
		header:        lipgloss.NewStyle().Foreground(lipgloss.Color(t.Header)).Background(lipgloss.Color(t.Background)).Bold(true),
		desc:          lipgloss.NewStyle().Foreground(lipgloss.Color(t.Description)).Background(lipgloss.Color(t.Background)),
		prompt:        lipgloss.NewStyle().Foreground(lipgloss.Color(t.Prompt)).Background(lipgloss.Color(t.Background)),
		query:         lipgloss.NewStyle().Foreground(lipgloss.Color(t.Query)).Background(lipgloss.Color(t.Background)),
		empty:         lipgloss.NewStyle().Foreground(lipgloss.Color(t.Empty)).Background(lipgloss.Color(t.Background)),
		selected:      lipgloss.NewStyle().Foreground(lipgloss.Color(t.SelectedFG)).Background(lipgloss.Color(t.SelectedBG)),
		selectedTitle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.SelectedFG)).Background(lipgloss.Color(t.SelectedBG)),
		selectedDesc:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.SelectedFG)).Background(lipgloss.Color(t.SelectedBG)),
		selectedChip:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.SelectedChip)).Background(lipgloss.Color(t.SelectedChipBG)),
		match:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.MatchFG)).Background(lipgloss.Color(t.Background)),
		selectedMatch: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.SelectedMatchFG)).Background(lipgloss.Color(t.SelectedBG)),
		chipMatch:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.MatchFG)).Background(lipgloss.Color(t.ChipBG)),
		glyph:         lipgloss.NewStyle().Foreground(lipgloss.Color(t.Glyph)).Background(lipgloss.Color(t.Background)),
		selectedGlyph: lipgloss.NewStyle().Foreground(lipgloss.Color(t.SelectedFG)).Background(lipgloss.Color(t.SelectedBG)),
		chip:          lipgloss.NewStyle().Foreground(lipgloss.Color(t.Chip)).Background(lipgloss.Color(t.ChipBG)),
		muted:         lipgloss.NewStyle().Foreground(lipgloss.Color(t.Muted)).Background(lipgloss.Color(t.Background)),
	}
}

func New(commands []config.Command, active theme.Theme, previewThemes []theme.Theme, showGlyphs bool, showDescription bool, recentKeys []string, configPath string, historyPath string) Model {
	return Model{
		commands:        commands,
		activeTheme:     active,
		previewThemes:   previewThemes,
		previewIndex:    previewIndex(previewThemes, active.Name),
		showGlyphs:      showGlyphs,
		showDescription: showDescription,
		showToggleHint:  true,
		tmuxDescription: true,
		recentKeys:      append([]string{}, recentKeys...),
		configPath:      configPath,
		historyPath:     historyPath,
		keys:            config.DefaultKeys(),
		caretVisible:    true,
		styles:          newStyles(active),
	}
}

func Run(commands []config.Command, active theme.Theme, previewThemes []theme.Theme, showGlyphs bool, showDescription bool, recentKeys []string, configPath string, historyPath string) (Result, error) {
	return RunWithState(commands, active, previewThemes, showGlyphs, showDescription, true, true, recentKeys, config.DefaultKeys(), nil, nil, configPath, historyPath, State{})
}

func RunWithState(commands []config.Command, active theme.Theme, previewThemes []theme.Theme, showGlyphs bool, showDescription bool, showToggleHint bool, tmuxDescription bool, recentKeys []string, keys config.Keys, tmuxCommands []tmuxcmd.Command, recentTmux []tmuxcmd.Invocation, configPath string, historyPath string, state State) (Result, error) {
	model := New(commands, active, previewThemes, showGlyphs, showDescription, recentKeys, configPath, historyPath)
	model.showToggleHint = showToggleHint
	model.tmuxDescription = tmuxDescription
	model.keys = normalizeKeys(keys)
	model.tmuxCommands = append([]tmuxcmd.Command{}, tmuxCommands...)
	model.recentTmux = append([]tmuxcmd.Invocation{}, recentTmux...)
	model.applyState(state)
	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return Result{}, err
	}
	model, ok := finalModel.(Model)
	if !ok {
		return Result{Theme: active, State: state}, nil
	}
	return Result{Command: model.selected, HistoryCommand: model.selectedHistory, Tmux: model.selectedTmux, Theme: model.activeTheme, State: model.state()}, nil
}

func (m Model) Init() tea.Cmd {
	return blinkCursor()
}

func (m *Model) applyState(state State) {
	m.query = state.Query
	m.cursor = state.Cursor
	m.offset = state.Offset
}

func (m Model) state() State {
	return State{
		Query:  m.query,
		Cursor: m.cursor,
		Offset: m.offset,
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case commandOutputMsg:
		m.openCommandOutput(msg.title, msg.body)
		return m, nil
	case messageCopiedMsg:
		if msg.err != nil {
			m.messageStatus = "Copy failed: " + msg.err.Error()
		} else {
			m.messageStatus = "Copied to tmux clipboard"
		}
		return m, nil
	case cursorBlinkMsg:
		m.caretVisible = !m.caretVisible
		return m, blinkCursor()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ensureCursorVisible()
	case tea.KeyMsg:
		if m.mode == modeThemePreview {
			return m.updateThemePreview(msg)
		}
		if m.mode == modeMessage {
			return m.updateMessage(msg)
		}
		if m.mode == modeTmuxArgs {
			return m.updateTmuxArgs(msg)
		}
		if m.mode == modeCommandInput {
			return m.updateCommandInput(msg)
		}
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.listMode == listModeTmux {
				return m.selectTmuxCommand()
			}
			return m.selectCommand()
		case m.keys.TmuxMode:
			m.toggleListMode()
		case "up", m.keys.MoveUp:
			if m.cursor > 0 {
				m.cursor--
			}
			m.ensureCursorVisible()
		case "down", m.keys.MoveDown:
			if m.cursor < m.matchCount()-1 {
				m.cursor++
			}
			m.ensureCursorVisible()
		case m.keys.ScrollUp:
			m.scrollViewport(-1)
		case m.keys.ScrollDown:
			m.scrollViewport(1)
		case m.keys.HalfPageUp:
			m.scrollList(-m.scrollHalfPage())
		case m.keys.HalfPageDown:
			m.scrollList(m.scrollHalfPage())
		case m.keys.NextCategory:
			m.moveToNextCategory()
			m.ensureCursorVisible()
		case m.keys.PreviousCategory:
			m.moveToPreviousCategory()
			m.ensureCursorVisible()
		case "backspace", "ctrl+h":
			if len(m.query) > 0 {
				m.removeQueryRune()
			}
		default:
			if len(msg.Runes) > 0 {
				m.appendQuery(string(msg.Runes))
			}
		}
	}
	return m, nil
}

func (m *Model) appendQuery(value string) {
	if strings.TrimSpace(m.query) == "" && value != "" && !m.preSearchActive {
		m.preSearchActive = true
		m.preSearchCursor = m.cursor
		m.preSearchOffset = m.offset
	}
	m.query += value
	m.cursor = 0
	m.offset = 0
	m.caretVisible = true
}

func (m *Model) removeQueryRune() {
	if len(m.query) == 0 {
		return
	}
	runes := []rune(m.query)
	m.query = string(runes[:len(runes)-1])
	if m.query == "" && m.preSearchActive {
		m.cursor = m.preSearchCursor
		m.offset = m.preSearchOffset
		m.preSearchActive = false
		m.ensureCursorVisible()
	} else {
		m.cursor = 0
		m.offset = 0
	}
	m.caretVisible = true
}

func blinkCursor() tea.Cmd {
	return tea.Tick(cursorBlinkInterval, func(time.Time) tea.Msg {
		return cursorBlinkMsg{}
	})
}

func (m Model) selectCommand() (tea.Model, tea.Cmd) {
	matches := m.matches()
	if len(matches) == 0 {
		return m, tea.Quit
	}
	m.clampSelection(len(matches))
	cmd := matches[m.cursor].Command
	if cmd.Internal == config.InternalThemePreview {
		m.mode = modeThemePreview
		m.previewIndex = previewIndex(m.previewThemes, m.activeTheme.Name)
		m.styles = newStyles(m.previewThemes[m.previewIndex])
		return m, nil
	}
	if cmd.Internal == config.InternalClearRecent || cmd.Internal == config.InternalConfigPath || cmd.Internal == config.InternalControls {
		m.openMessage(cmd.Internal)
		return m, nil
	}
	if cmd.Prompt != "" {
		spec, ok := commandPromptSpec(cmd.Prompt)
		if ok {
			m.mode = modeCommandInput
			m.inputCommand = cmd
			m.inputPrompt = spec
			m.inputValue = ""
			m.caretVisible = true
			return m, nil
		}
	}
	if cmd.ShowOutput {
		m.openCommandOutput(cmd.Title, "Running...")
		return m, runOutputCommand(cmd)
	}
	m.selected = &cmd
	return m, tea.Quit
}

func (m Model) selectTmuxCommand() (tea.Model, tea.Cmd) {
	matches := m.tmuxMatches()
	if len(matches) == 0 {
		return m, nil
	}
	m.clampSelection(len(matches))
	selected := matches[m.cursor]
	if selected.Recent {
		invocation := selected.Invocation
		m.selectedTmux = &invocation
		return m, tea.Quit
	}
	m.argCommand = selected.Command
	m.argValue = ""
	if !selected.Command.TakesArgs {
		invocation := tmuxcmd.Invocation{Name: selected.Command.Name}
		m.selectedTmux = &invocation
		return m, tea.Quit
	}
	m.mode = modeTmuxArgs
	m.caretVisible = true
	return m, nil
}

func (m *Model) toggleListMode() {
	if m.listMode == listModeTmux {
		m.listMode = listModeCommands
	} else {
		m.listMode = listModeTmux
	}
	m.cursor = 0
	m.offset = 0
	m.preSearchActive = false
	m.caretVisible = true
}

func (m Model) updateThemePreview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "enter":
		m.mode = modeCommands
		m.activeTheme = m.previewThemes[m.previewIndex]
		m.styles = newStyles(m.activeTheme)
	case "up", "left", "ctrl+p":
		m.previousTheme()
	case "down", "right", "ctrl+n":
		m.nextTheme()
	}
	return m, nil
}

func (m Model) updateTmuxArgs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.mode = modeCommands
		m.argValue = ""
	case "enter":
		invocation := tmuxcmd.Invocation{Name: m.argCommand.Name, Args: m.argValue}
		m.selectedTmux = &invocation
		return m, tea.Quit
	case "backspace", "ctrl+h":
		if len(m.argValue) > 0 {
			m.argValue = m.argValue[:len(m.argValue)-1]
			m.caretVisible = true
		}
	default:
		if len(msg.Runes) > 0 {
			m.argValue += string(msg.Runes)
			m.caretVisible = true
		}
	}
	return m, nil
}

func (m Model) updateCommandInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.mode = modeCommands
		m.inputValue = ""
	case "enter":
		if !validPromptValue(m.inputPrompt, m.inputValue) {
			return m, nil
		}
		cmd := m.inputCommand
		historyCommand := m.inputCommand
		cmd.Command = renderCommandInput(cmd.Command, m.inputValue)
		if cmd.ShowOutput {
			m.selectedHistory = &historyCommand
			m.openCommandOutput(cmd.Title, "Running...")
			return m, runOutputCommand(cmd)
		}
		m.selected = &cmd
		m.selectedHistory = &historyCommand
		return m, tea.Quit
	case "backspace", "ctrl+h":
		if len(m.inputValue) > 0 {
			m.inputValue = m.inputValue[:len(m.inputValue)-1]
			m.caretVisible = true
		}
	default:
		if len(msg.Runes) > 0 {
			m.inputValue += string(msg.Runes)
			m.caretVisible = true
		}
	}
	return m, nil
}

func (m Model) updateMessage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.mode = modeCommands
		m.messageTitle = ""
		m.messageBody = ""
		m.messageStatus = ""
	case "c", "y":
		m.messageStatus = "Copying..."
		return m, copyMessageBody(m.messageBody)
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	s := m.styles
	if m.mode == modeThemePreview {
		return m.viewThemePreview()
	}
	if m.mode == modeMessage {
		return m.viewMessage()
	}
	if m.mode == modeTmuxArgs {
		return m.viewTmuxArgs()
	}
	if m.mode == modeCommandInput {
		return m.viewCommandInput()
	}
	if m.listMode == listModeTmux {
		return m.viewTmuxCommands()
	}
	matches := m.matches()
	var b strings.Builder
	b.WriteString(m.renderSearchBox())
	if m.showToggleHint {
		b.WriteString("\n")
		b.WriteString(m.renderModeHint())
	}
	b.WriteString("\n\n")

	if len(matches) == 0 {
		b.WriteString("\n")
		b.WriteString(s.empty.Render("No commands found"))
		return m.renderFrame(b.String())
	}

	lastCategory := ""
	showHeaders := strings.TrimSpace(m.query) == ""
	lineBudget := m.commandLineBudget()
	linesUsed := 0
	start := m.offset
	if start >= len(matches) {
		start = max(0, len(matches)-1)
	}

	for rowIndex := start; rowIndex < len(matches); rowIndex++ {
		match := matches[rowIndex]
		cmd := match.Command
		if showHeaders && cmd.Category != lastCategory {
			if lastCategory == recentCategory {
				if linesUsed+1 > lineBudget {
					break
				}
				b.WriteString(m.renderRecentDivider())
				b.WriteString("\n")
				linesUsed++
			}
			if linesUsed+1 > lineBudget {
				break
			}
			b.WriteString(s.header.Render(cmd.Category))
			b.WriteString("\n")
			lastCategory = cmd.Category
			linesUsed++
		}

		selected := rowIndex == m.cursor
		contentWidth := m.innerWidth()
		row := renderRow(match, s, selected, m.showGlyphs, m.showDescription, contentWidth)
		rowLines := lipgloss.Height(row)
		if linesUsed+rowLines > lineBudget {
			break
		}
		b.WriteString(row)
		b.WriteString("\n")
		linesUsed += rowLines
		if linesUsed >= lineBudget {
			remaining := len(matches) - rowIndex - 1
			if remaining > 0 && linesUsed < lineBudget {
				b.WriteString(s.muted.Render(fmt.Sprintf("%d more...", remaining)))
			}
			break
		}
	}
	return m.renderFrame(b.String())
}

func (m Model) viewTmuxCommands() string {
	matches := m.tmuxMatches()
	s := m.styles
	var b strings.Builder
	b.WriteString(m.renderSearchBox())
	if m.showToggleHint {
		b.WriteString("\n")
		b.WriteString(m.renderModeHint())
	}
	b.WriteString("\n\n")

	if len(matches) == 0 {
		b.WriteString("\n")
		b.WriteString(s.empty.Render("No tmux commands found"))
		return m.renderFrame(b.String())
	}

	lastCategory := ""
	showHeaders := strings.TrimSpace(m.query) == ""
	lineBudget := m.commandLineBudget()
	linesUsed := 0
	start := m.offset
	if start >= len(matches) {
		start = max(0, len(matches)-1)
	}

	for rowIndex := start; rowIndex < len(matches); rowIndex++ {
		match := matches[rowIndex].Match
		cmd := match.Command
		if showHeaders && cmd.Category != lastCategory {
			if lastCategory == recentCategory {
				if linesUsed+1 > lineBudget {
					break
				}
				b.WriteString(m.renderRecentDivider())
				b.WriteString("\n")
				linesUsed++
			}
			if linesUsed+1 > lineBudget {
				break
			}
			b.WriteString(s.header.Render(cmd.Category))
			b.WriteString("\n")
			lastCategory = cmd.Category
			linesUsed++
		}

		selected := rowIndex == m.cursor
		row := renderRow(match, s, selected, m.showGlyphs, m.tmuxDescription, m.innerWidth())
		if linesUsed+1 > lineBudget {
			break
		}
		b.WriteString(row)
		b.WriteString("\n")
		linesUsed++
	}
	return m.renderFrame(b.String())
}

func (m Model) viewThemePreview() string {
	s := m.styles
	var b strings.Builder
	b.WriteString(m.renderSubviewHeader("Theme Preview"))
	b.WriteString("\n\n")
	b.WriteString(s.header.Render("Included Themes"))
	b.WriteString("\n")
	b.WriteString(m.renderThemeList())
	b.WriteString("\n\n")
	b.WriteString(s.header.Render("Sample Commands"))
	b.WriteString("\n")
	b.WriteString(renderCommandRow(config.Command{
		Title:       "Split Horizontal",
		Description: "Split pane side by side",
		Aliases:     []string{"sh"},
		Icon:        "",
	}, s, false, m.showGlyphs, m.showDescription, m.innerWidth()))
	b.WriteString("\n")
	b.WriteString(renderCommandRow(config.Command{
		Title:       "Lazygit",
		Description: "Open lazygit in a popup",
		Aliases:     []string{"lg"},
		Icon:        "󰊢",
	}, s, false, m.showGlyphs, m.showDescription, m.innerWidth()))
	b.WriteString("\n\n")
	b.WriteString(s.selected.Width(m.innerWidth()).Render("  Selected row preview"))
	b.WriteString("\n\n")
	b.WriteString(s.muted.Render("Up/Down or Left/Right previews themes, Enter/Esc returns"))
	return m.renderFrame(b.String())
}

func (m Model) viewMessage() string {
	s := m.styles
	var b strings.Builder
	b.WriteString(m.renderSubviewHeader(m.messageTitle))
	b.WriteString("\n\n")
	for _, line := range strings.Split(m.messageBody, "\n") {
		b.WriteString(m.renderMessageLine(line))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	if m.messageStatus != "" {
		b.WriteString(s.desc.Render(m.messageStatus))
		b.WriteString("\n")
	}
	b.WriteString(s.muted.Render("c/y copies, Esc/q returns to commands"))
	return m.renderFrame(b.String())
}

func (m Model) viewTmuxArgs() string {
	s := m.styles
	var b strings.Builder
	b.WriteString(m.renderSubviewHeader("tmux Command Arguments"))
	b.WriteString("\n\n")
	b.WriteString(s.header.Render(m.argCommand.Name))
	if m.argCommand.Description != "" {
		b.WriteString("\n")
		b.WriteString(s.muted.Render(m.argCommand.Description))
	}
	if m.argCommand.Usage != "" {
		b.WriteString("\n\n")
		b.WriteString(s.desc.Render(m.argCommand.Usage))
	}
	if len(m.argCommand.ArgHelp) > 0 {
		b.WriteString("\n\n")
		for _, line := range m.argCommand.ArgHelp {
			b.WriteString(s.muted.Render(line))
			b.WriteString("\n")
		}
	}
	b.WriteString("\n\n")
	b.WriteString(m.renderArgumentBox())
	b.WriteString("\n\n")
	b.WriteString(s.muted.Render("Enter runs tmux command, Esc returns to command list"))
	return m.renderFrame(b.String())
}

func (m Model) viewCommandInput() string {
	s := m.styles
	var b strings.Builder
	b.WriteString(m.renderSubviewHeader("Command Input"))
	b.WriteString("\n\n")
	b.WriteString(s.header.Render(m.inputCommand.Title))
	if m.inputCommand.Description != "" {
		b.WriteString("\n")
		b.WriteString(s.muted.Render(m.inputCommand.Description))
	}
	if m.inputPrompt.Description != "" {
		b.WriteString("\n\n")
		b.WriteString(s.desc.Render(m.inputPrompt.Description))
	}
	b.WriteString("\n\n")
	b.WriteString(m.renderInputBox(m.inputPrompt.Label, m.inputValue))
	b.WriteString("\n\n")
	b.WriteString(s.muted.Render("Enter runs command, Esc returns to command list"))
	return m.renderFrame(b.String())
}

func (m Model) renderArgumentBox() string {
	return m.renderInputBox("args", m.argValue)
}

func (m Model) renderInputBox(label string, value string) string {
	width := m.innerWidth()
	fill := m.activeTheme.SearchBG
	contentWidth := width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}

	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.activeTheme.SearchFG)).Background(lipgloss.Color(fill))
	cursor := " "
	if m.caretVisible {
		cursor = "█"
	}
	promptStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.activeTheme.Glyph)).Background(lipgloss.Color(fill))
	prompt := label + " ❯ "
	prefix := promptStyle.Render(prompt)
	inputBudget := contentWidth - lipgloss.Width(prompt) - lipgloss.Width(cursor)
	if inputBudget < 1 {
		inputBudget = 1
	}
	content := prefix + valueStyle.Render(truncate(value, inputBudget)) + promptStyle.Render(cursor)

	return lipgloss.NewStyle().
		Width(width-2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.activeTheme.PromptBorder)).
		BorderBackground(lipgloss.Color(m.activeTheme.Background)).
		Background(lipgloss.Color(fill)).
		Padding(0, 1).
		Render(content)
}

func (m Model) renderMessageLine(line string) string {
	if isPathLine(line) {
		return m.renderPathLine(line)
	}
	if isSeparatorLine(line) {
		return m.styles.muted.Render(line)
	}
	if isControlsCategory(line) {
		return m.styles.header.Render(line)
	}
	return m.styles.title.Render(line)
}

func isSeparatorLine(line string) bool {
	line = strings.TrimSpace(line)
	return line != "" && strings.Trim(line, "─") == ""
}

func isControlsCategory(line string) bool {
	switch strings.TrimSpace(line) {
	case "Global", "tmux Command Arguments", "Theme Preview", "Internal Messages":
		return true
	default:
		return false
	}
}

func (m Model) renderPathLine(line string) string {
	bar := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.activeTheme.Glyph)).
		Background(lipgloss.Color(m.activeTheme.Background)).
		Render("▌")
	path := m.styles.selected.Render(line)
	return m.styles.root.Width(m.innerWidth()).Render(bar + path)
}

func isPathLine(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "/") || strings.HasPrefix(line, "~") || strings.Contains(line, ":\\")
}

func (m *Model) openMessage(internal string) {
	m.messageStatus = ""
	switch internal {
	case config.InternalClearRecent:
		m.messageTitle = "Clear Recent Commands"
		if m.historyPath == "" {
			m.messageBody = "Recent command history is disabled."
			break
		}
		if err := os.Remove(m.historyPath); err != nil && !os.IsNotExist(err) {
			m.messageBody = "Could not clear recent command history:\n" + err.Error()
			break
		}
		m.recentKeys = nil
		m.messageBody = "Recent command history cleared:\n\n" + m.historyPath
	case config.InternalConfigPath:
		m.messageTitle = "Config Path"
		m.messageBody = m.configPath
	case config.InternalControls:
		m.messageTitle = "Controls"
		m.messageBody = m.controlsMessage()
	default:
		m.messageTitle = "Message"
		m.messageBody = ""
	}
	m.mode = modeMessage
}

func (m *Model) openCommandOutput(title string, body string) {
	m.mode = modeMessage
	m.messageTitle = title
	m.messageBody = strings.TrimRight(body, "\n")
	m.messageStatus = ""
	if m.messageBody == "" {
		m.messageBody = "(no output)"
	}
}

func runOutputCommand(cmd config.Command) tea.Cmd {
	return func() tea.Msg {
		return commandOutputMsg{title: cmd.Title, body: commandOutput(cmd)}
	}
}

func copyMessageBody(body string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("tmux", "load-buffer", "-w", "-")
		cmd.Stdin = strings.NewReader(body)
		return messageCopiedMsg{err: cmd.Run()}
	}
}

func commandOutput(cmd config.Command) string {
	action := strings.ToLower(strings.TrimSpace(cmd.Action))
	command := strings.TrimSpace(cmd.Command)

	var execCmd *exec.Cmd
	switch action {
	case "tmux":
		execCmd = exec.Command(shellPath(), "-lc", "tmux "+command)
	case "shell":
		execCmd = exec.Command(shellPath(), "-lc", command)
	default:
		return fmt.Sprintf("show_output is only supported for tmux and shell actions, not %q", cmd.Action)
	}

	output, err := execCmd.CombinedOutput()
	body := string(output)
	if err != nil {
		if body != "" {
			body += "\n"
		}
		body += err.Error()
	}
	return body
}

func shellPath() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/sh"
}

func (m Model) controlsMessage() string {
	return strings.Join([]string{
		"Global",
		"────────",
		controlLine("Up / "+displayKey(m.keys.MoveUp), "Move selection up"),
		controlLine("Down / "+displayKey(m.keys.MoveDown), "Move selection down"),
		controlLine(displayKey(m.keys.ScrollUp), "Scroll up one row"),
		controlLine(displayKey(m.keys.ScrollDown), "Scroll down one row"),
		controlLine(displayKey(m.keys.HalfPageUp), "Scroll up half page"),
		controlLine(displayKey(m.keys.HalfPageDown), "Scroll down half page"),
		controlLine(displayKey(m.keys.NextCategory), "Next category"),
		controlLine(displayKey(m.keys.PreviousCategory), "Previous category"),
		"Enter              Select focused entry",
		"Esc / Ctrl-C       Close commander",
		controlLine(displayKey(m.keys.TmuxMode), "Toggle command mode"),
		"",
		"tmux Command Arguments",
		"────────",
		"Enter              Run tmux command",
		"Esc                Return to tmux command list",
		"",
		"Theme Preview",
		"────────",
		"Up / Left          Previous theme",
		"Down / Right       Next theme",
		"Enter / Esc        Return to command list",
		"",
		"Internal Messages",
		"────────",
		"Esc / q            Return to command list",
	}, "\n")
}

func controlLine(key string, description string) string {
	const descriptionColumn = 19
	padding := descriptionColumn - lipgloss.Width(key)
	if padding < 1 {
		padding = 1
	}
	return key + strings.Repeat(" ", padding) + description
}

func (m Model) renderSubviewHeader(title string) string {
	label := " " + title + " "
	width := m.innerWidth()
	if lipgloss.Width(label) >= width {
		return m.styles.header.Width(width).Align(lipgloss.Center).Render(title)
	}

	left := (width - lipgloss.Width(label)) / 2
	right := width - lipgloss.Width(label) - left
	return m.styles.muted.Render(strings.Repeat("─", left)) +
		m.styles.header.Render(label) +
		m.styles.muted.Render(strings.Repeat("─", right))
}

func (m Model) renderThemeList() string {
	if len(m.previewThemes) == 0 {
		return m.styles.empty.Render("No themes available")
	}

	var b strings.Builder
	for i, previewTheme := range m.previewThemes {
		if i > 0 {
			b.WriteString("\n")
		}
		label := "  " + previewTheme.Name
		if i == m.previewIndex {
			selectedLabel := m.styles.glyph.Render("▌") + m.styles.selected.Render(" "+previewTheme.Name)
			b.WriteString(padSelectedRow(selectedLabel, m.innerWidth(), m.styles))
			continue
		}
		b.WriteString(m.styles.title.Render(label))
	}
	return b.String()
}

func (m *Model) previousTheme() {
	if len(m.previewThemes) == 0 {
		return
	}
	m.previewIndex--
	if m.previewIndex < 0 {
		m.previewIndex = len(m.previewThemes) - 1
	}
	m.styles = newStyles(m.previewThemes[m.previewIndex])
}

func (m *Model) nextTheme() {
	if len(m.previewThemes) == 0 {
		return
	}
	m.previewIndex++
	if m.previewIndex >= len(m.previewThemes) {
		m.previewIndex = 0
	}
	m.styles = newStyles(m.previewThemes[m.previewIndex])
}

func previewIndex(themes []theme.Theme, name string) int {
	for i, t := range themes {
		if t.Name == name {
			return i
		}
	}
	return 0
}

func (m Model) matches() []fuzzy.Match {
	if strings.TrimSpace(m.query) == "" {
		return fuzzy.Filter(m.commandsForEmptyQuery(), "")
	}
	matches := fuzzy.Filter(m.commands, m.query)
	for i := range matches {
		matches[i].Score += m.recentBoostForCommand(matches[i].Command)
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Score == matches[j].Score {
			return matches[i].Command.Title < matches[j].Command.Title
		}
		return matches[i].Score > matches[j].Score
	})
	return matches
}

func (m Model) tmuxMatches() []tmuxMatch {
	commands := m.tmuxCommandsForQuery()
	matches := fuzzy.Filter(commands, m.query)
	result := make([]tmuxMatch, 0, len(matches))
	for _, match := range matches {
		tmuxCommand, invocation, recent := m.tmuxMatchData(match.Command)
		if strings.TrimSpace(m.query) != "" {
			match.Score += m.recentTmuxBoost(invocation.Key())
		}
		result = append(result, tmuxMatch{
			Match:      match,
			Command:    tmuxCommand,
			Invocation: invocation,
			Recent:     recent,
		})
	}
	if strings.TrimSpace(m.query) != "" {
		sort.SliceStable(result, func(i, j int) bool {
			if result[i].Match.Score == result[j].Match.Score {
				return result[i].Match.Command.Title < result[j].Match.Command.Title
			}
			return result[i].Match.Score > result[j].Match.Score
		})
	}
	return result
}

func (m Model) tmuxCommandsForQuery() []config.Command {
	if strings.TrimSpace(m.query) == "" {
		return m.tmuxCommandsForEmptyQuery()
	}
	commands := make([]config.Command, 0, len(m.tmuxCommands))
	for _, cmd := range m.tmuxCommands {
		commands = append(commands, tmuxCommandToConfig(cmd, tmuxCommandCategory))
	}
	return commands
}

func (m Model) tmuxCommandsForEmptyQuery() []config.Command {
	commands := make([]config.Command, 0, len(m.recentTmux)+len(m.tmuxCommands))
	for _, invocation := range m.recentTmux {
		title := invocation.Name
		description := invocation.Args
		if description == "" {
			description = "Run without arguments"
		}
		commands = append(commands, config.Command{
			Title:       title,
			Description: description,
			Category:    recentCategory,
			Action:      "tmux",
			Command:     invocation.CommandLine(),
			Icon:        "▸",
		})
	}
	for _, cmd := range m.tmuxCommands {
		commands = append(commands, tmuxCommandToConfig(cmd, tmuxCommandCategory))
	}
	return commands
}

func tmuxCommandToConfig(cmd tmuxcmd.Command, category string) config.Command {
	description := cmd.Description
	if description == "" {
		description = "tmux command"
	}
	aliases := []string{}
	if cmd.TakesArgs {
		aliases = append(aliases, "args")
	}
	return config.Command{
		Title:       cmd.Name,
		Description: description,
		Category:    category,
		Aliases:     aliases,
		Action:      "tmux",
		Command:     cmd.Name,
		Icon:        "▸",
	}
}

func (m Model) tmuxMatchData(cmd config.Command) (tmuxcmd.Command, tmuxcmd.Invocation, bool) {
	if cmd.Category == recentCategory {
		invocation := tmuxcmd.Invocation{Name: cmd.Title, Args: strings.TrimSpace(strings.TrimPrefix(cmd.Command, cmd.Title))}
		return m.findTmuxCommand(invocation.Name), invocation, true
	}
	return m.findTmuxCommand(cmd.Title), tmuxcmd.Invocation{Name: cmd.Title}, false
}

func (m Model) findTmuxCommand(name string) tmuxcmd.Command {
	for _, cmd := range m.tmuxCommands {
		if cmd.Name == name {
			return cmd
		}
	}
	return tmuxcmd.Command{Name: name, TakesArgs: true}
}

func commandPromptSpec(name string) (promptSpec, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "session_name":
		return promptSpec{Name: "session_name", Label: "session name", Description: "Enter the tmux session name to use for this command."}, true
	case "window_name":
		return promptSpec{Name: "window_name", Label: "window name", Description: "Enter the tmux window name to use for this command."}, true
	case "target_index":
		return promptSpec{Name: "target_index", Label: "target index", Description: "Enter the tmux target index to use for this command."}, true
	case "count":
		return promptSpec{Name: "count", Label: "count", Description: "Enter the number of times to repeat this command."}, true
	case "file_path":
		return promptSpec{Name: "file_path", Label: "file path", Description: "Enter the file path to use for this command."}, true
	case "command":
		return promptSpec{Name: "command", Label: "command", Description: "Enter the command string to use for this command."}, true
	case "search_query":
		return promptSpec{Name: "search_query", Label: "search query", Description: "Enter the search query to use for this command."}, true
	default:
		return promptSpec{}, false
	}
}

func validPromptValue(spec promptSpec, value string) bool {
	if spec.Name != "count" {
		return true
	}
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func renderCommandInput(command string, value string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(command, "{{raw_input}}", value),
		"{{input}}",
		shellQuote(value),
	)
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func (m Model) commandsForEmptyQuery() []config.Command {
	if len(m.recentKeys) == 0 {
		return m.commands
	}
	byKey := map[string]config.Command{}
	for _, cmd := range m.commands {
		if cmd.Internal != "" {
			continue
		}
		byKey[config.CommandKey(cmd)] = cmd
		byKey[config.CommandTitleKey(cmd)] = cmd
	}
	commands := make([]config.Command, 0, len(m.commands))
	seen := map[string]bool{}
	for _, key := range m.recentKeys {
		cmd, ok := byKey[key]
		if !ok {
			continue
		}
		commandKey := config.CommandKey(cmd)
		if seen[commandKey] {
			continue
		}
		seen[commandKey] = true
		cmd.Category = recentCategory
		commands = append(commands, cmd)
	}
	for _, cmd := range m.commands {
		commands = append(commands, cmd)
	}
	return commands
}

func (m Model) recentBoostForCommand(cmd config.Command) int {
	if boost := m.recentBoost(config.CommandKey(cmd)); boost > 0 {
		return boost
	}
	return m.recentBoost(config.CommandTitleKey(cmd))
}

func (m Model) recentBoost(key string) int {
	for i, recentKey := range m.recentKeys {
		if recentKey == key {
			boost := recentScoreBoost - i
			if boost < 1 {
				return 1
			}
			return boost
		}
	}
	return 0
}

func (m Model) recentTmuxBoost(key string) int {
	for i, invocation := range m.recentTmux {
		if invocation.Key() == key {
			boost := recentScoreBoost - i
			if boost < 1 {
				return 1
			}
			return boost
		}
	}
	return 0
}

func (m Model) matchCount() int {
	if m.listMode == listModeTmux {
		return len(m.tmuxMatches())
	}
	return len(m.matches())
}

func (m *Model) moveToNextCategory() {
	if m.listMode == listModeTmux {
		matches := m.tmuxMatches()
		categories := make([]fuzzy.Match, 0, len(matches))
		for _, match := range matches {
			categories = append(categories, match.Match)
		}
		m.cursor = nextCategoryIndex(categories, m.cursor)
		return
	}
	matches := m.matches()
	m.cursor = nextCategoryIndex(matches, m.cursor)
}

func (m *Model) moveToPreviousCategory() {
	if m.listMode == listModeTmux {
		matches := m.tmuxMatches()
		categories := make([]fuzzy.Match, 0, len(matches))
		for _, match := range matches {
			categories = append(categories, match.Match)
		}
		m.cursor = previousCategoryIndex(categories, m.cursor)
		return
	}
	matches := m.matches()
	m.cursor = previousCategoryIndex(matches, m.cursor)
}

func (m *Model) scrollList(delta int) {
	count := m.matchCount()
	if count == 0 || delta == 0 {
		return
	}
	m.cursor = clampCursor(m.cursor+delta, count)
	m.offset = clampCursor(m.offset+delta, count)
	m.ensureCursorVisible()
}

func (m *Model) scrollViewport(delta int) {
	count := m.matchCount()
	if count == 0 || delta == 0 {
		return
	}
	nextOffset := m.offset + delta
	if nextOffset < 0 {
		return
	}
	if delta > 0 && m.isLastMatchVisible() {
		return
	}
	m.offset = clampCursor(nextOffset, count)
}

func (m Model) scrollHalfPage() int {
	halfPage := m.commandLineBudget() / 2
	if halfPage < 1 {
		return 1
	}
	return halfPage
}

func (m Model) isLastMatchVisible() bool {
	count := m.matchCount()
	if count == 0 {
		return true
	}
	saved := m
	saved.cursor = count - 1
	return saved.cursorVisible()
}

func nextCategoryIndex(matches []fuzzy.Match, cursor int) int {
	if len(matches) == 0 {
		return 0
	}
	cursor = clampCursor(cursor, len(matches))
	current := matches[cursor].Command.Category
	for i := cursor + 1; i < len(matches); i++ {
		if matches[i].Command.Category != current {
			return i
		}
	}
	if matches[0].Command.Category == current {
		return cursor
	}
	return 0
}

func previousCategoryIndex(matches []fuzzy.Match, cursor int) int {
	if len(matches) == 0 {
		return 0
	}
	cursor = clampCursor(cursor, len(matches))
	currentStart := categoryStart(matches, cursor)
	if currentStart == 0 {
		if matches[len(matches)-1].Command.Category == matches[cursor].Command.Category {
			return cursor
		}
		return categoryStart(matches, len(matches)-1)
	}
	return categoryStart(matches, currentStart-1)
}

func categoryStart(matches []fuzzy.Match, index int) int {
	index = clampCursor(index, len(matches))
	category := matches[index].Command.Category
	for index > 0 && matches[index-1].Command.Category == category {
		index--
	}
	return index
}

func clampCursor(cursor int, length int) int {
	if cursor < 0 {
		return 0
	}
	if cursor >= length {
		return length - 1
	}
	return cursor
}

func (m *Model) clampSelection(count int) {
	if count == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}
	m.cursor = clampCursor(m.cursor, count)
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m *Model) ensureCursorVisible() {
	count := m.matchCount()
	if count == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}
	if m.cursor >= count {
		m.cursor = count - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	for !m.cursorVisible() && m.offset < m.cursor {
		m.offset++
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m Model) commandLineBudget() int {
	rows := m.contentHeight() - 7
	if !m.showToggleHint {
		rows++
	}
	if rows < 1 {
		return 1
	}
	return rows
}

func (m Model) renderModeHint() string {
	current := "Commands"
	if m.listMode == listModeTmux {
		current = "tmux Commands"
	}
	return m.styles.muted.Render(fmt.Sprintf("Mode: %s · %s to toggle", current, displayKey(m.keys.TmuxMode)))
}

func normalizeKeys(keys config.Keys) config.Keys {
	defaults := config.DefaultKeys()
	if strings.TrimSpace(keys.TmuxMode) == "" {
		keys.TmuxMode = defaults.TmuxMode
	}
	if strings.TrimSpace(keys.MoveUp) == "" {
		keys.MoveUp = defaults.MoveUp
	}
	if strings.TrimSpace(keys.MoveDown) == "" {
		keys.MoveDown = defaults.MoveDown
	}
	if strings.TrimSpace(keys.ScrollUp) == "" {
		keys.ScrollUp = defaults.ScrollUp
	}
	if strings.TrimSpace(keys.ScrollDown) == "" {
		keys.ScrollDown = defaults.ScrollDown
	}
	if strings.TrimSpace(keys.HalfPageUp) == "" {
		keys.HalfPageUp = defaults.HalfPageUp
	}
	if strings.TrimSpace(keys.HalfPageDown) == "" {
		keys.HalfPageDown = defaults.HalfPageDown
	}
	if strings.TrimSpace(keys.NextCategory) == "" {
		keys.NextCategory = defaults.NextCategory
	}
	if strings.TrimSpace(keys.PreviousCategory) == "" {
		keys.PreviousCategory = defaults.PreviousCategory
	}
	keys.TmuxMode = config.NormalizeKey(keys.TmuxMode)
	keys.MoveUp = config.NormalizeKey(keys.MoveUp)
	keys.MoveDown = config.NormalizeKey(keys.MoveDown)
	keys.ScrollUp = config.NormalizeKey(keys.ScrollUp)
	keys.ScrollDown = config.NormalizeKey(keys.ScrollDown)
	keys.HalfPageUp = config.NormalizeKey(keys.HalfPageUp)
	keys.HalfPageDown = config.NormalizeKey(keys.HalfPageDown)
	keys.NextCategory = config.NormalizeKey(keys.NextCategory)
	keys.PreviousCategory = config.NormalizeKey(keys.PreviousCategory)
	return keys
}

func displayKey(key string) string {
	parts := strings.Split(config.NormalizeKey(key), "+")
	for i, part := range parts {
		switch part {
		case "ctrl":
			parts[i] = "Ctrl"
		case "alt":
			parts[i] = "Alt"
		case "shift":
			parts[i] = "Shift"
		default:
			if len(part) == 1 {
				parts[i] = strings.ToUpper(part)
			} else if part != "" {
				parts[i] = strings.ToUpper(part[:1]) + part[1:]
			}
		}
	}
	return strings.Join(parts, "-")
}

func (m Model) renderSearchBox() string {
	width := m.innerWidth()
	fill := m.activeTheme.SearchBG
	contentWidth := width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}

	value := m.query
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.activeTheme.SearchFG)).Background(lipgloss.Color(fill))

	cursor := " "
	if m.caretVisible {
		cursor = "█"
	}

	promptStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.activeTheme.Glyph)).Background(lipgloss.Color(fill))
	prompt := "❯ "
	if m.listMode == listModeTmux {
		prompt = "tmux ❯ "
	}
	prefix := promptStyle.Render(prompt)
	inputBudget := contentWidth - lipgloss.Width(prompt) - lipgloss.Width(cursor)
	if inputBudget < 1 {
		inputBudget = 1
	}
	content := prefix + valueStyle.Render(truncate(value, inputBudget)) + promptStyle.Render(cursor)

	return lipgloss.NewStyle().
		Width(width-2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.activeTheme.PromptBorder)).
		BorderBackground(lipgloss.Color(m.activeTheme.Background)).
		Background(lipgloss.Color(fill)).
		Padding(0, 1).
		Render(content)
}

func (m Model) innerWidth() int {
	width := m.contentWidth() - horizontalPadding*2
	if width < 1 {
		return 1
	}
	return width
}

func (m Model) contentWidth() int {
	width := m.width - 2
	if width < 1 {
		return 1
	}
	return width
}

func (m Model) contentHeight() int {
	height := m.height - 2
	if height < 1 {
		return 1
	}
	return height
}

func (m Model) renderFrame(content string) string {
	s := m.styles
	inner := s.root.Padding(0, horizontalPadding).Width(m.contentWidth()).Height(m.contentHeight()).Render(content)
	return s.frame.Width(m.contentWidth()).Height(m.contentHeight()).Render(inner)
}

func (m Model) renderRecentDivider() string {
	width := m.innerWidth()
	if width < 1 {
		width = 1
	}
	return m.styles.muted.Render(strings.Repeat("─", width))
}

func (m Model) cursorVisible() bool {
	if m.listMode == listModeTmux {
		return m.tmuxCursorVisible(m.tmuxMatches())
	}
	return m.commandCursorVisible(m.matches())
}

func (m Model) commandCursorVisible(matches []fuzzy.Match) bool {
	if m.cursor < m.offset {
		return false
	}
	lineBudget := m.commandLineBudget()
	linesUsed := 0
	lastCategory := ""
	showHeaders := strings.TrimSpace(m.query) == ""
	for rowIndex := m.offset; rowIndex < len(matches); rowIndex++ {
		cmd := matches[rowIndex].Command
		if showHeaders && cmd.Category != lastCategory {
			if lastCategory == recentCategory {
				if linesUsed+1 > lineBudget {
					return false
				}
				linesUsed++
			}
			if linesUsed+1 > lineBudget {
				return false
			}
			linesUsed++
			lastCategory = cmd.Category
		}
		rowLines := commandRowLines(cmd)
		if linesUsed+rowLines > lineBudget {
			return false
		}
		if rowIndex == m.cursor {
			return true
		}
		linesUsed += rowLines
	}
	return false
}

func (m Model) tmuxCursorVisible(matches []tmuxMatch) bool {
	if m.cursor < m.offset {
		return false
	}
	lineBudget := m.commandLineBudget()
	linesUsed := 0
	lastCategory := ""
	showHeaders := strings.TrimSpace(m.query) == ""
	for rowIndex := m.offset; rowIndex < len(matches); rowIndex++ {
		cmd := matches[rowIndex].Match.Command
		if showHeaders && cmd.Category != lastCategory {
			if lastCategory == recentCategory {
				if linesUsed+1 > lineBudget {
					return false
				}
				linesUsed++
			}
			if linesUsed+1 > lineBudget {
				return false
			}
			linesUsed++
			lastCategory = cmd.Category
		}
		if linesUsed+1 > lineBudget {
			return false
		}
		if rowIndex == m.cursor {
			return true
		}
		linesUsed++
	}
	return false
}

func commandRowLines(config.Command) int {
	return 1
}

func renderCommandRow(cmd config.Command, s styles, selected bool, showGlyphs bool, showDescription bool, width int) string {
	return renderRow(fuzzy.Match{Command: cmd}, s, selected, showGlyphs, showDescription, width)
}

func renderRow(match fuzzy.Match, s styles, selected bool, showGlyphs bool, showDescription bool, width int) string {
	cmd := match.Command
	descStyle := s.desc
	if selected {
		descStyle = s.selectedDesc
	}
	line := renderRowPrefix(match, s, selected, showGlyphs)
	if showDescription && cmd.Description != "" {
		separator := " • "
		budget := width - lipgloss.Width(line) - lipgloss.Width(separator)
		if budget > 0 {
			line += descStyle.Render(separator + truncate(cmd.Description, budget))
		}
	}
	if selected {
		return padSelectedRow(line, width, s)
	}
	return line
}

func renderRowPrefix(match fuzzy.Match, s styles, selected bool, showGlyphs bool) string {
	cmd := match.Command
	titleStyle := s.title
	chipStyle := s.chip
	glyphStyle := s.glyph
	matchStyle := s.match
	chipMatchStyle := s.chipMatch
	spacerStyle := s.root
	if selected {
		titleStyle = s.selectedTitle
		chipStyle = s.selectedChip
		glyphStyle = s.selectedGlyph
		matchStyle = s.selectedMatch
		spacerStyle = s.selected
	}

	var b strings.Builder
	b.WriteString(rowIndent(selected, s))
	if showGlyphs && strings.TrimSpace(cmd.Icon) != "" {
		b.WriteString(renderRowPart(cmd.Icon, glyphStyle, true))
	}
	hasAliases := len(cmd.Aliases) > 0
	b.WriteString(renderMatchedRowPart(cmd.Title, titleStyle, matchStyle, match.TitleIndexes, hasAliases))
	for i, alias := range cmd.Aliases {
		b.WriteString(renderAliasChip(alias, chipStyle, chipMatchStyle, match.AliasIndexes[alias]))
		if i < len(cmd.Aliases)-1 {
			b.WriteString(spacerStyle.Render(" "))
		}
	}
	return b.String()
}

func renderRowPart(value string, style lipgloss.Style, withTrailingSpace bool) string {
	if withTrailingSpace {
		return style.Width(lipgloss.Width(value) + 1).Render(value)
	}
	return style.Render(value)
}

func renderMatchedRowPart(value string, style lipgloss.Style, matchStyle lipgloss.Style, indexes []int, withTrailingSpace bool) string {
	rendered := renderMatchedText(value, style, matchStyle, indexes)
	if withTrailingSpace {
		rendered += style.Render(" ")
	}
	return rendered
}

func renderAliasChip(alias string, style lipgloss.Style, matchStyle lipgloss.Style, indexes []int) string {
	return style.Render(" ") + renderMatchedText(alias, style, matchStyle, indexes) + style.Render(" ")
}

func renderMatchedText(value string, style lipgloss.Style, matchStyle lipgloss.Style, indexes []int) string {
	if len(indexes) == 0 {
		return style.Render(value)
	}

	matched := map[int]bool{}
	for _, index := range indexes {
		matched[index] = true
	}

	var b strings.Builder
	for index, r := range value {
		part := string(r)
		if matched[index] {
			b.WriteString(matchStyle.Render(part))
			continue
		}
		b.WriteString(style.Render(part))
	}
	return b.String()
}

func rowIndent(selected bool, s styles) string {
	if selected {
		return s.glyph.Render("▌") + s.selected.Render(" ")
	}
	return "  "
}

func padSelectedRow(line string, width int, s styles) string {
	fill := width - lipgloss.Width(line)
	if fill <= 0 {
		return line
	}
	return line + s.selected.Render(strings.Repeat(" ", fill))
}

func truncate(value string, maxWidth int) string {
	if maxWidth < 1 {
		return ""
	}
	if lipgloss.Width(value) <= maxWidth {
		return value
	}
	if maxWidth == 1 {
		return "…"
	}
	runes := []rune(value)
	for len(runes) > 0 && lipgloss.Width(string(runes)+"…") > maxWidth {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}
