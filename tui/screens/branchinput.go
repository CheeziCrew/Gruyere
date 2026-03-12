package screens

import (
	"fmt"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// BranchInputModel handles the two-field branch input.
type BranchInputModel struct {
	baseInput    textinput.Model
	featureInput textinput.Model
	focusBase    bool // true = base field focused, false = feature
	width        int
	height       int
}

// StartChangelogMsg is sent when the user confirms both branches.
type StartChangelogMsg struct {
	BaseBranch    string
	FeatureBranch string
}

func NewBranchInput() BranchInputModel {
	base := newStyledInput("base branch (e.g. main)")
	base.Focus()
	base.CharLimit = 100
	base.SetValue("main")

	feature := newStyledInput("feature branch")
	feature.CharLimit = 100

	return BranchInputModel{
		baseInput:    base,
		featureInput: feature,
		focusBase:    true,
	}
}

func newStyledInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetWidth(60)
	s := textinput.DefaultStyles(true)
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(colorBrCyan)
	s.Focused.Text = lipgloss.NewStyle().Foreground(colorFg)
	ti.SetStyles(s)
	return ti
}

func (m BranchInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m BranchInputModel) Update(msg tea.Msg) (BranchInputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }

		case "tab", "down":
			if m.focusBase {
				m.focusBase = false
				m.baseInput.Blur()
				m.featureInput.Focus()
				return m, textinput.Blink
			}

		case "shift+tab", "up":
			if !m.focusBase {
				m.focusBase = true
				m.featureInput.Blur()
				m.baseInput.Focus()
				return m, textinput.Blink
			}

		case "enter":
			if !m.focusBase && m.featureInput.Value() != "" {
				return m, func() tea.Msg {
					return StartChangelogMsg{
						BaseBranch:    m.baseInput.Value(),
						FeatureBranch: m.featureInput.Value(),
					}
				}
			}
			if m.focusBase {
				m.focusBase = false
				m.baseInput.Blur()
				m.featureInput.Focus()
				return m, textinput.Blink
			}
		}
	}

	var cmd tea.Cmd
	if m.focusBase {
		m.baseInput, cmd = m.baseInput.Update(msg)
	} else {
		m.featureInput, cmd = m.featureInput.Update(msg)
	}
	return m, cmd
}

func (m BranchInputModel) View() string {
	title := accentStyle.Render("Generate Changelog")
	desc := dimStyle.Render("Enter the branches to diff")

	baseLabel := dimStyle.Render("Base branch:")
	featureLabel := dimStyle.Render("Feature branch:")

	box := inputBox.Render(
		fmt.Sprintf("%s\n%s\n\n%s\n%s\n\n%s\n%s",
			title, desc,
			baseLabel, m.baseInput.View(),
			featureLabel, m.featureInput.View(),
		),
	)

	help := helpStyle.Render(dimStyle.Render("tab switch  enter confirm  esc back"))
	return box + "\n\n" + help
}
