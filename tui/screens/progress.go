package screens

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ProgressModel shows a spinner while changelog generation runs.
type ProgressModel struct {
	spinner spinner.Model
	base    string
	feature string
	width   int
	height  int
}

func NewProgress(base, feature string) ProgressModel {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(colorBrCyan)

	return ProgressModel{
		spinner: s,
		base:    base,
		feature: feature,
	}
}

func (m ProgressModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m ProgressModel) View() string {
	title := accentStyle.Render("Generating Changelog")
	info := dimStyle.Render(m.base + " → " + m.feature)
	spin := m.spinner.View() + " Scanning for OpenAPI specs and computing diff..."

	return inputBox.Render(title + "\n" + info + "\n\n" + spin)
}
