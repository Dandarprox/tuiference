package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Dandarprox/tuiference/internal/wordreference"
)

type lookupMsg struct {
	results []wordreference.Result
	err     error
}

type Model struct {
	input       textinput.Model
	viewport    viewport.Model
	client      wordreference.Client
	originIndex int
	targetIndex int
	loading     bool
	message     string
	results     []wordreference.Result
	width       int
	height      int
}

var (
	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	mutedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	pairStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("222"))
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	tableHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("111"))
	sectionStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("228"))
	wordStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255"))
	infoStyle        = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("150"))
	noteStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
)

func New() Model {
	input := textinput.New()
	input.Placeholder = "type a word or phrase"
	input.Focus()
	input.CharLimit = 120
	input.Prompt = "lookup: "

	vp := viewport.New(80, 20)

	return Model{
		input:       input,
		viewport:    vp,
		client:      wordreference.NewClient(),
		originIndex: 0,
		targetIndex: 1,
		message:     "Enter a lookup term.",
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = max(3, msg.Height-6)
		m.refreshViewport()
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab":
			m.originIndex = nextLanguage(m.originIndex, m.targetIndex)
			m.message = "Origin language changed."
			return m, nil
		case "shift+tab":
			m.targetIndex = nextLanguage(m.targetIndex, m.originIndex)
			m.message = "Target language changed."
			return m, nil
		case "ctrl+w":
			m.deleteCurrentWord()
			return m, nil
		case "enter":
			term := strings.TrimSpace(m.input.Value())
			m.loading = true
			m.message = "Looking up " + term + "..."
			m.results = nil
			m.refreshViewport()
			return m, lookupCmd(m.client, m.origin(), m.target(), term)
		case "up", "down", "pgup", "pgdown":
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	case lookupMsg:
		m.loading = false
		if msg.err != nil {
			m.message = msg.err.Error()
			m.results = nil
		} else {
			m.results = msg.results
			m.message = fmt.Sprintf("%d result rows", len(msg.results))
		}
		m.refreshViewport()
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("tuiference"))
	b.WriteString(mutedStyle.Render("                         Tab origin | Shift+Tab target | Ctrl+W delete word | Enter lookup | Esc quit"))
	b.WriteString("\n\n")
	b.WriteString(pairStyle.Render(fmt.Sprintf("[ %s ] -> [ %s ]", m.origin().Name, m.target().Name)))
	b.WriteString("\n\n")
	b.WriteString(m.input.View())
	b.WriteString("\n")

	status := m.message
	if m.loading {
		status = "Loading..."
	}
	if status != "" {
		style := mutedStyle
		if !m.loading && len(m.results) == 0 && status != "Enter a lookup term." {
			style = errorStyle
		}
		b.WriteString(style.Render(status))
		b.WriteString("\n")
	}

	b.WriteString(m.viewport.View())
	return b.String()
}

func (m Model) origin() wordreference.Language {
	return wordreference.Languages[m.originIndex]
}

func (m Model) target() wordreference.Language {
	return wordreference.Languages[m.targetIndex]
}

func (m *Model) deleteCurrentWord() {
	value := m.input.Value()
	pos := m.input.Position()
	if pos == 0 || value == "" {
		return
	}

	left := value[:pos]
	right := value[pos:]
	left = strings.TrimRight(left, " \t")
	idx := strings.LastIndexAny(left, " \t")
	if idx == -1 {
		left = ""
	} else {
		left = left[:idx+1]
	}

	m.input.SetValue(left + right)
	m.input.SetCursor(len(left))
}

func (m *Model) refreshViewport() {
	m.viewport.SetContent(renderTable(m.results, max(60, m.width)))
}

func lookupCmd(client wordreference.Client, origin, target wordreference.Language, term string) tea.Cmd {
	return func() tea.Msg {
		results, err := client.Lookup(context.Background(), origin, target, term)
		return lookupMsg{results: results, err: err}
	}
}

func nextLanguage(current, other int) int {
	next := (current + 1) % len(wordreference.Languages)
	if next == other {
		next = (next + 1) % len(wordreference.Languages)
	}
	return next
}

func renderTable(results []wordreference.Result, width int) string {
	if len(results) == 0 {
		return ""
	}

	leftW := max(18, width/3)
	rightW := max(18, width/3)
	notesW := max(16, width-leftW-rightW-10)

	var b strings.Builder
	currentSection := ""
	for _, result := range results {
		if result.Section != currentSection {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			currentSection = result.Section
			b.WriteString(sectionStyle.Render(currentSection))
			b.WriteString("\n")
			b.WriteString(tableHeaderStyle.Render(formatWRRow("Original", "Translation", "Notes", leftW, rightW, notesW)))
			b.WriteString("\n")
			b.WriteString(mutedStyle.Render(strings.Repeat("─", min(width, leftW+rightW+notesW+10))))
			b.WriteString("\n")
		}

		b.WriteString(formatResultRow(result, leftW, rightW, notesW))
		b.WriteString("\n")
	}
	return b.String()
}

func formatResultRow(result wordreference.Result, leftW, rightW, notesW int) string {
	source := wordStyle.Render(truncate(result.Source, leftW))
	if result.SourceInfo != "" {
		source = source + " " + infoStyle.Render(truncate(result.SourceInfo, max(4, leftW-len([]rune(result.Source))-1)))
	}

	translation := wordStyle.Render(truncate(result.Translation, rightW))
	if result.TranslationInfo != "" {
		translation = translation + " " + infoStyle.Render(truncate(result.TranslationInfo, max(4, rightW-len([]rune(result.Translation))-1)))
	}

	return fmt.Sprintf("  %-*s  %-*s  %s", leftW, lipgloss.NewStyle().Width(leftW).Render(source), rightW, lipgloss.NewStyle().Width(rightW).Render(translation), noteStyle.Render(truncate(result.Notes, notesW)))
}

func formatWRRow(source, translation, notes string, leftW, rightW, notesW int) string {
	return fmt.Sprintf("  %-*s  %-*s  %-*s", leftW, truncate(source, leftW), rightW, truncate(translation, rightW), notesW, truncate(notes, notesW))
}

func truncate(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
