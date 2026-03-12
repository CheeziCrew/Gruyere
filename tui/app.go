package tui

import (
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/CheeziCrew/curd"
	"github.com/CheeziCrew/gruyere/internal/ops"
	"github.com/CheeziCrew/gruyere/tui/screens"
)

type screen int

const (
	screenMenu screen = iota
	screenRepoSelect
	screenBranchInput
	screenProgress
	screenResults
)

// Model is the root Bubble Tea model.
type Model struct {
	current    screen
	menu       screens.MenuModel
	repoSelect screens.RepoSelectModel
	branchInput screens.BranchInputModel
	progress   screens.ProgressModel
	results    screens.ResultsModel
	scanPath   string
	width      int
	height     int
}

// New creates a fresh root model.
func New() Model {
	return Model{
		current: screenMenu,
		menu:    screens.NewMenu(),
		width:   80,
		height:  24,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.menu, _ = m.menu.Update(msg)
		switch m.current {
		case screenRepoSelect:
			m.repoSelect, _ = m.repoSelect.Update(msg)
		case screenBranchInput:
			m.branchInput, _ = m.branchInput.Update(msg)
		case screenProgress:
			m.progress, _ = m.progress.Update(msg)
		case screenResults:
			m.results, _ = m.results.Update(msg)
		}
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "q" && m.current == screenMenu {
			return m, tea.Quit
		}

	case curd.MenuSelectionMsg:
		if msg.Command == "changelog" {
			m.current = screenRepoSelect
			m.repoSelect = screens.NewRepoSelect("changelog", ".", 0, m.height)
			return m, m.repoSelect.Init()
		}

	case screens.RepoSelectDoneMsg:
		if len(msg.Paths) > 0 {
			repoPath := msg.Paths[0]
			repoName := filepath.Base(repoPath)
			m.current = screenBranchInput
			m.branchInput = screens.NewBranchInput(repoPath, repoName)
			return m, m.branchInput.Init()
		}

	case screens.StartChangelogMsg:
		m.current = screenProgress
		m.progress = screens.NewProgress(msg.BaseBranch, msg.FeatureBranch)
		return m, tea.Batch(
			m.progress.Init(),
			generateChangelog(msg.BaseBranch, msg.FeatureBranch, msg.RepoPath),
		)

	case screens.ChangelogDoneMsg:
		m.current = screenResults
		m.results = screens.NewResults(msg.Markdown, msg.OpenAPIPath, msg.HasChanges, msg.Err, m.width, m.height)
		return m, nil

	case screens.BackToMenuMsg:
		m.current = screenMenu
		m.menu = screens.NewMenu()
		return m, func() tea.Msg {
			return tea.WindowSizeMsg{Width: m.width, Height: m.height}
		}
	}

	var cmd tea.Cmd
	switch m.current {
	case screenMenu:
		m.menu, cmd = m.menu.Update(msg)
	case screenRepoSelect:
		m.repoSelect, cmd = m.repoSelect.Update(msg)
	case screenBranchInput:
		m.branchInput, cmd = m.branchInput.Update(msg)
	case screenProgress:
		m.progress, cmd = m.progress.Update(msg)
	case screenResults:
		m.results, cmd = m.results.Update(msg)
	}
	return m, cmd
}

func (m Model) View() tea.View {
	var content string
	switch m.current {
	case screenMenu:
		content = m.menu.View()
	case screenRepoSelect:
		content = m.repoSelect.View()
	case screenBranchInput:
		content = m.branchInput.View()
	case screenProgress:
		content = m.progress.View()
	case screenResults:
		content = m.results.View()
	}
	v := tea.NewView(lipgloss.NewStyle().Padding(1, 2, 0, 2).Render(content))
	v.AltScreen = true
	v.WindowTitle = "gruyere"
	return v
}

func generateChangelog(base, feature, repoPath string) tea.Cmd {
	return func() tea.Msg {
		result, err := ops.GenerateChangelog(ops.Input{
			BaseBranch:    base,
			FeatureBranch: feature,
			RepoPath:      repoPath,
		})
		if err != nil {
			return screens.ChangelogDoneMsg{Err: err}
		}
		return screens.ChangelogDoneMsg{
			Markdown:    result.Markdown,
			HasChanges:  result.HasChanges,
			OpenAPIPath: result.OpenAPIPath,
		}
	}
}
