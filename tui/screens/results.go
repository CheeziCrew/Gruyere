package screens

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ResultsModel displays the generated changelog with copy/write actions.
type ResultsModel struct {
	viewport    viewport.Model
	markdown    string
	openAPIPath string
	hasChanges  bool
	err         error
	flash       string
	width       int
	height      int
	ready       bool
}

func NewResults(markdown, openAPIPath string, hasChanges bool, err error, w, h int) ResultsModel {
	m := ResultsModel{
		markdown:    markdown,
		openAPIPath: openAPIPath,
		hasChanges:  hasChanges,
		err:         err,
		width:       w,
		height:      h,
	}
	m.initViewport(w, h)
	return m
}

func (m *ResultsModel) initViewport(w, h int) {
	vw := w - 6
	if vw < 40 {
		vw = 40
	}
	vh := h - 8
	if vh < 10 {
		vh = 10
	}
	m.viewport = viewport.New(viewport.WithWidth(vw), viewport.WithHeight(vh))

	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(colorRed).Bold(true)
		m.viewport.SetContent(errStyle.Render("Error: "+m.err.Error()) + "\n\n" + dimStyle.Render("esc to go back"))
	} else if !m.hasChanges {
		m.viewport.SetContent(accentStyle.Render("No API changes found.") + "\n\n" + dimStyle.Render("esc to go back"))
	} else {
		m.viewport.SetContent(m.markdown)
	}
	m.ready = true
}

func (m ResultsModel) Init() tea.Cmd {
	return nil
}

func (m ResultsModel) Update(msg tea.Msg) (ResultsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.initViewport(msg.Width, msg.Height)
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg { return BackToMenuMsg{} }

		case "w":
			if m.hasChanges {
				if err := os.WriteFile("changelog.md", []byte(m.markdown), 0644); err != nil {
					m.flash = fmt.Sprintf("Error: %v", err)
				} else {
					m.flash = "Written to changelog.md"
				}
				return m, nil
			}

		case "y":
			if m.hasChanges {
				m.flash = "Copied to clipboard"
				return m, copyToClipboard(m.markdown)
			}
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m ResultsModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	title := accentStyle.Render("API Changelog")
	if m.openAPIPath != "" {
		title += "  " + dimStyle.Render(m.openAPIPath)
	}

	var footer string
	if m.flash != "" {
		footer = "\n" + lipgloss.NewStyle().Foreground(colorGreen).Render(m.flash)
	}

	help := helpStyle.Render(dimStyle.Render("w write  y copy  esc back  ↑↓ scroll"))

	return title + "\n\n" + m.viewport.View() + footer + "\n" + help
}

// clipboardMsg is sent after a clipboard copy attempt.
type clipboardMsg struct{ err error }

func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
		err := cmd.Run()
		return clipboardMsg{err: err}
	}
}
