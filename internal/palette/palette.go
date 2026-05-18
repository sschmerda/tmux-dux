package palette

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stefanschmerda/tmux-commander/internal/config"
	"github.com/stefanschmerda/tmux-commander/internal/fuzzy"
	"github.com/stefanschmerda/tmux-commander/internal/theme"
)

type Model struct {
	commands      []config.Command
	activeTheme   theme.Theme
	previewThemes []theme.Theme
	previewIndex  int
	styles        styles
	mode          mode
	query         string
	cursor        int
	offset        int
	selected      *config.Command
	width         int
	height        int
}

type mode int

const (
	modeCommands mode = iota
	modeThemePreview
)

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
	chip          lipgloss.Style
	muted         lipgloss.Style
}

func newStyles(t theme.Theme) styles {
	return styles{
		root:          lipgloss.NewStyle().Background(lipgloss.Color(t.Background)),
		frame:         lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(t.Header)).BorderBackground(lipgloss.Color(t.Background)).Background(lipgloss.Color(t.Background)),
		title:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.Title)).Background(lipgloss.Color(t.Background)),
		header:        lipgloss.NewStyle().Foreground(lipgloss.Color(t.Header)).Background(lipgloss.Color(t.Background)).Bold(true),
		desc:          lipgloss.NewStyle().Foreground(lipgloss.Color(t.Description)).Background(lipgloss.Color(t.Background)),
		prompt:        lipgloss.NewStyle().Foreground(lipgloss.Color(t.Prompt)).Background(lipgloss.Color(t.Background)),
		query:         lipgloss.NewStyle().Foreground(lipgloss.Color(t.Query)).Background(lipgloss.Color(t.Background)),
		empty:         lipgloss.NewStyle().Foreground(lipgloss.Color(t.Empty)).Background(lipgloss.Color(t.Background)),
		selected:      lipgloss.NewStyle().Foreground(lipgloss.Color(t.SelectedFG)).Background(lipgloss.Color(t.SelectedBG)),
		selectedTitle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.SelectedFG)).Background(lipgloss.Color(t.SelectedBG)),
		selectedDesc:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.SelectedFG)).Background(lipgloss.Color(t.SelectedBG)),
		selectedChip:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.SelectedFG)).Background(lipgloss.Color(t.SelectedBG)),
		chip:          lipgloss.NewStyle().Foreground(lipgloss.Color(t.Chip)).Background(lipgloss.Color(t.Background)),
		muted:         lipgloss.NewStyle().Foreground(lipgloss.Color(t.Muted)).Background(lipgloss.Color(t.Background)),
	}
}

func New(commands []config.Command, active theme.Theme, previewThemes []theme.Theme) Model {
	return Model{
		commands:      commands,
		activeTheme:   active,
		previewThemes: previewThemes,
		previewIndex:  previewIndex(previewThemes, active.Name),
		styles:        newStyles(active),
	}
}

func Run(commands []config.Command, active theme.Theme, previewThemes []theme.Theme) (*config.Command, error) {
	program := tea.NewProgram(New(commands, active, previewThemes), tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return nil, err
	}
	model, ok := finalModel.(Model)
	if !ok {
		return nil, nil
	}
	return model.selected, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ensureCursorVisible()
	case tea.KeyMsg:
		if m.mode == modeThemePreview {
			return m.updateThemePreview(msg)
		}
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "enter":
			matches := fuzzy.Filter(m.commands, m.query)
			if len(matches) > 0 {
				cmd := matches[m.cursor].Command
				if cmd.Internal == config.InternalThemePreview {
					m.mode = modeThemePreview
					m.query = ""
					m.cursor = 0
					m.previewIndex = previewIndex(m.previewThemes, m.activeTheme.Name)
					m.styles = newStyles(m.previewThemes[m.previewIndex])
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
			if m.cursor < len(fuzzy.Filter(m.commands, m.query))-1 {
				m.cursor++
			}
			m.ensureCursorVisible()
		case "backspace", "ctrl+h":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.cursor = 0
				m.offset = 0
			}
		default:
			if len(msg.Runes) > 0 {
				m.query += string(msg.Runes)
				m.cursor = 0
				m.offset = 0
			}
		}
	}
	return m, nil
}

func (m Model) updateThemePreview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "enter":
		m.mode = modeCommands
		m.styles = newStyles(m.activeTheme)
	case "up", "left", "ctrl+p":
		m.previousTheme()
	case "down", "right", "ctrl+n":
		m.nextTheme()
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	matches := fuzzy.Filter(m.commands, m.query)
	s := m.styles
	if m.mode == modeThemePreview {
		return m.viewThemePreview()
	}
	var b strings.Builder
	b.WriteString(s.title.Render("tmux-commander"))
	b.WriteString("\n")
	b.WriteString(s.prompt.Render("Search: "))
	b.WriteString(s.query.Render(m.query))
	if m.query == "" {
		b.WriteString(s.muted.Render(" Type to filter"))
	}
	b.WriteString("\n")

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
			if linesUsed+1 > lineBudget {
				break
			}
			b.WriteString(s.header.Render(cmd.Category))
			b.WriteString("\n")
			lastCategory = cmd.Category
			linesUsed++
		}

		selected := rowIndex == m.cursor
		contentWidth := m.contentWidth()
		row := renderRow(cmd, s, selected)
		rowLines := lipgloss.Height(row)
		if linesUsed+rowLines > lineBudget {
			break
		}
		if selected {
			row = s.selected.Width(contentWidth).Render(row)
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
	current := m.previewThemes[m.previewIndex]
	var b strings.Builder
	b.WriteString(s.title.Render("Theme Preview"))
	b.WriteString("\n")
	b.WriteString(s.prompt.Render("Theme: "))
	b.WriteString(s.query.Render(current.Name))
	b.WriteString("\n")
	b.WriteString(s.muted.Render(`Set with: theme = "` + current.Name + `"`))
	b.WriteString("\n\n")
	b.WriteString(s.header.Render("Sample Commands"))
	b.WriteString("\n")
	b.WriteString(renderRow(config.Command{
		Title:       "Split Horizontal",
		Description: "Split pane side by side",
		Aliases:     []string{"sh"},
	}, s, false))
	b.WriteString("\n")
	b.WriteString(renderRow(config.Command{
		Title:       "Lazygit",
		Description: "Open lazygit in a popup",
		Aliases:     []string{"lg"},
		Icon:        "git",
	}, s, false))
	b.WriteString("\n\n")
	b.WriteString(s.selected.Width(m.contentWidth()).Render("  Selected row preview"))
	b.WriteString("\n\n")
	b.WriteString(s.muted.Render("Up/Down or Left/Right previews themes, Enter/Esc returns"))
	return m.renderFrame(b.String())
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

func (m *Model) ensureCursorVisible() {
	matches := fuzzy.Filter(m.commands, m.query)
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
	for !m.cursorVisible(matches) && m.offset < m.cursor {
		m.offset++
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m Model) commandLineBudget() int {
	rows := m.contentHeight() - 4
	if rows < 1 {
		return 1
	}
	return rows
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
	return s.frame.Width(m.contentWidth()).Height(m.contentHeight()).Render(content)
}

func (m Model) cursorVisible(matches []fuzzy.Match) bool {
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

func commandRowLines(cmd config.Command) int {
	if cmd.Description != "" {
		return 2
	}
	return 1
}

func renderRow(cmd config.Command, s styles, selected bool) string {
	icon := iconLabel(cmd.Icon)
	title := strings.TrimSpace(icon + " " + cmd.Title)
	meta := []string{}
	for _, alias := range cmd.Aliases {
		meta = append(meta, "#"+alias)
	}
	titleStyle := s.title
	descStyle := s.desc
	chipStyle := s.chip
	if selected {
		titleStyle = s.selectedTitle
		descStyle = s.selectedDesc
		chipStyle = s.selectedChip
	}
	line := "  " + titleStyle.Render(title)
	if len(meta) > 0 {
		line += chipStyle.Render(" " + strings.Join(meta, " "))
	}
	if cmd.Description != "" {
		line += "\n    " + descStyle.Render(cmd.Description)
	}
	return line
}

func iconLabel(icon string) string {
	switch icon {
	case "git":
		return "[git]"
	case "cpu":
		return "[cpu]"
	case "":
		return ""
	default:
		return "[" + icon + "]"
	}
}
