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
	repoPath     string
	repoName     string
	width        int
	height       int
}

func NewBranchInput(repoPath, repoName string) BranchInputModel {
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
		repoPath:     repoPath,
		repoName:     repoName,
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
		if updated, cmd, handled := m.handleKeyPress(msg); handled {
			return updated, cmd
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

func (m BranchInputModel) handleKeyPress(msg tea.KeyPressMsg) (BranchInputModel, tea.Cmd, bool) {
	switch msg.String() {
	case "esc":
		return m, func() tea.Msg { return BackToMenuMsg{} }, true

	case "tab", "down":
		if m.focusBase {
			return m.focusFeature(), textinput.Blink, true
		}

	case "shift+tab", "up":
		if !m.focusBase {
			return m.focusBaseField(), textinput.Blink, true
		}

	case "enter":
		return m.handleEnter()
	}
	return m, nil, false
}

func (m BranchInputModel) focusFeature() BranchInputModel {
	m.focusBase = false
	m.baseInput.Blur()
	m.featureInput.Focus()
	return m
}

func (m BranchInputModel) focusBaseField() BranchInputModel {
	m.focusBase = true
	m.featureInput.Blur()
	m.baseInput.Focus()
	return m
}

func (m BranchInputModel) handleEnter() (BranchInputModel, tea.Cmd, bool) {
	if !m.focusBase && m.featureInput.Value() != "" {
		repoPath := m.repoPath
		return m, func() tea.Msg {
			return StartChangelogMsg{
				BaseBranch:    m.baseInput.Value(),
				FeatureBranch: m.featureInput.Value(),
				RepoPath:      repoPath,
			}
		}, true
	}
	if m.focusBase {
		return m.focusFeature(), textinput.Blink, true
	}
	return m, nil, false
}

func (m BranchInputModel) View() string {
	title := accentStyle.Render("Generate Changelog")
	desc := dimStyle.Render("Enter the branches to diff for ") + accentStyle.Render(m.repoName)

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
