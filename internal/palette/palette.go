package palette

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stefanschmerda/tmux-commander/internal/config"
	"github.com/stefanschmerda/tmux-commander/internal/fuzzy"
	"github.com/stefanschmerda/tmux-commander/internal/theme"
)

type Model struct {
	commands        []config.Command
	activeTheme     theme.Theme
	previewThemes   []theme.Theme
	previewIndex    int
	showGlyphs      bool
	showDescription bool
	recentKeys      []string
	configPath      string
	historyPath     string
	styles          styles
	mode            mode
	query           string
	messageTitle    string
	messageBody     string
	caretVisible    bool
	cursor          int
	offset          int
	selected        *config.Command
	width           int
	height          int
}

type Result struct {
	Command *config.Command
	Theme   theme.Theme
	State   State
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
)

const horizontalPadding = 3

const cursorBlinkInterval = 500 * time.Millisecond

const recentCategory = "Recent"

const recentScoreBoost = 50

type cursorBlinkMsg struct{}

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
		recentKeys:      append([]string{}, recentKeys...),
		configPath:      configPath,
		historyPath:     historyPath,
		caretVisible:    true,
		styles:          newStyles(active),
	}
}

func Run(commands []config.Command, active theme.Theme, previewThemes []theme.Theme, showGlyphs bool, showDescription bool, recentKeys []string, configPath string, historyPath string) (Result, error) {
	return RunWithState(commands, active, previewThemes, showGlyphs, showDescription, recentKeys, configPath, historyPath, State{})
}

func RunWithState(commands []config.Command, active theme.Theme, previewThemes []theme.Theme, showGlyphs bool, showDescription bool, recentKeys []string, configPath string, historyPath string, state State) (Result, error) {
	model := New(commands, active, previewThemes, showGlyphs, showDescription, recentKeys, configPath, historyPath)
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
	return Result{Command: model.selected, Theme: model.activeTheme, State: model.state()}, nil
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
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "enter":
			matches := m.matches()
			if len(matches) > 0 {
				cmd := matches[m.cursor].Command
				if cmd.Internal == config.InternalThemePreview {
					m.mode = modeThemePreview
					m.previewIndex = previewIndex(m.previewThemes, m.activeTheme.Name)
					m.styles = newStyles(m.previewThemes[m.previewIndex])
					return m, nil
				}
				if cmd.Internal == config.InternalClearRecent || cmd.Internal == config.InternalConfigPath {
					m.openMessage(cmd.Internal)
					return m, nil
				}
				m.selected = &cmd
			}
			return m, tea.Quit
		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}
			m.ensureCursorVisible()
		case "down", "ctrl+n":
			if m.cursor < len(m.matches())-1 {
				m.cursor++
			}
			m.ensureCursorVisible()
		case "tab":
			m.moveToNextCategory()
			m.ensureCursorVisible()
		case "shift+tab":
			m.moveToPreviousCategory()
			m.ensureCursorVisible()
		case "backspace", "ctrl+h":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.cursor = 0
				m.offset = 0
				m.caretVisible = true
			}
		default:
			if len(msg.Runes) > 0 {
				m.query += string(msg.Runes)
				m.cursor = 0
				m.offset = 0
				m.caretVisible = true
			}
		}
	}
	return m, nil
}

func blinkCursor() tea.Cmd {
	return tea.Tick(cursorBlinkInterval, func(time.Time) tea.Msg {
		return cursorBlinkMsg{}
	})
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

func (m Model) updateMessage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.mode = modeCommands
		m.messageTitle = ""
		m.messageBody = ""
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	matches := m.matches()
	s := m.styles
	if m.mode == modeThemePreview {
		return m.viewThemePreview()
	}
	if m.mode == modeMessage {
		return m.viewMessage()
	}
	var b strings.Builder
	b.WriteString(m.renderSearchBox())
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
	b.WriteString(s.muted.Render("Esc/q returns to commands"))
	return m.renderFrame(b.String())
}

func (m Model) renderMessageLine(line string) string {
	if isPathLine(line) {
		return m.renderPathLine(line)
	}
	return m.styles.title.Render(line)
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
	default:
		m.messageTitle = "Message"
		m.messageBody = ""
	}
	m.mode = modeMessage
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
			b.WriteString(m.styles.selected.Width(m.innerWidth()).Render("  " + previewTheme.Name))
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
		matches[i].Score += m.recentBoost(config.CommandKey(matches[i].Command))
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Score == matches[j].Score {
			return matches[i].Command.Title < matches[j].Command.Title
		}
		return matches[i].Score > matches[j].Score
	})
	return matches
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
	}
	commands := make([]config.Command, 0, len(m.commands))
	for _, key := range m.recentKeys {
		cmd, ok := byKey[key]
		if !ok {
			continue
		}
		cmd.Category = recentCategory
		commands = append(commands, cmd)
	}
	for _, cmd := range m.commands {
		commands = append(commands, cmd)
	}
	return commands
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

func (m *Model) moveToNextCategory() {
	matches := m.matches()
	m.cursor = nextCategoryIndex(matches, m.cursor)
}

func (m *Model) moveToPreviousCategory() {
	matches := m.matches()
	m.cursor = previousCategoryIndex(matches, m.cursor)
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

func (m *Model) ensureCursorVisible() {
	matches := m.matches()
	if len(matches) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}
	if m.cursor >= len(matches) {
		m.cursor = len(matches) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	for !m.commandCursorVisible(matches) && m.offset < m.cursor {
		m.offset++
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m Model) commandLineBudget() int {
	rows := m.contentHeight() - 6
	if rows < 1 {
		return 1
	}
	return rows
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
		separator := " - "
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
	indent := "  "
	if selected {
		return s.selected.Render(indent)
	}
	return indent
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
