package palette

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stefanschmerda/tmux-commander/internal/config"
	"github.com/stefanschmerda/tmux-commander/internal/fuzzy"
)

type Model struct {
	commands []config.Command
	query    string
	cursor   int
	selected *config.Command
	width    int
	height   int
}

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	headerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true).PaddingTop(1)
	descStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4"))
	chipStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	mutedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func New(commands []config.Command) Model {
	return Model{commands: commands}
}

func Run(commands []config.Command) (*config.Command, error) {
	program := tea.NewProgram(New(commands), tea.WithAltScreen())
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
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "enter":
			matches := fuzzy.Filter(m.commands, m.query)
			if len(matches) > 0 {
				cmd := matches[m.cursor].Command
				m.selected = &cmd
			}
			return m, tea.Quit
		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "ctrl+n":
			if m.cursor < len(fuzzy.Filter(m.commands, m.query))-1 {
				m.cursor++
			}
		case "backspace", "ctrl+h":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.cursor = 0
			}
		default:
			if len(msg.Runes) > 0 {
				m.query += string(msg.Runes)
				m.cursor = 0
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	matches := fuzzy.Filter(m.commands, m.query)
	var b strings.Builder
	b.WriteString(titleStyle.Render("tmux-commander"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Search: "))
	b.WriteString(m.query)
	if m.query == "" {
		b.WriteString(mutedStyle.Render(" Type to filter"))
	}
	b.WriteString("\n")

	if len(matches) == 0 {
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render("No commands found"))
		return b.String()
	}

	rowIndex := 0
	lastCategory := ""
	showHeaders := strings.TrimSpace(m.query) == ""
	maxRows := m.height - 4
	if maxRows < 4 {
		maxRows = 4
	}

	for _, match := range matches {
		cmd := match.Command
		if showHeaders && cmd.Category != lastCategory {
			b.WriteString(headerStyle.Render(cmd.Category))
			b.WriteString("\n")
			lastCategory = cmd.Category
		}

		row := renderRow(cmd)
		if rowIndex == m.cursor {
			row = selectedStyle.Width(max(1, m.width-1)).Render(row)
		}
		b.WriteString(row)
		b.WriteString("\n")
		rowIndex++
		if rowIndex >= maxRows {
			remaining := len(matches) - rowIndex
			if remaining > 0 {
				b.WriteString(mutedStyle.Render(fmt.Sprintf("%d more...", remaining)))
			}
			break
		}
	}
	return b.String()
}

func renderRow(cmd config.Command) string {
	icon := iconLabel(cmd.Icon)
	title := strings.TrimSpace(icon + " " + cmd.Title)
	meta := []string{}
	if cmd.Category != "" {
		meta = append(meta, cmd.Category)
	}
	for _, alias := range cmd.Aliases {
		meta = append(meta, "#"+alias)
	}
	line := "  " + titleStyle.Render(title)
	if len(meta) > 0 {
		line += " " + chipStyle.Render(strings.Join(meta, " "))
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
