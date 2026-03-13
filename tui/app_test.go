package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/CheeziCrew/curd"
	"github.com/CheeziCrew/gruyere/tui/screens"
)

func keyPress(ch rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: ch, Text: string(ch)}
}

func specialKey(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}

func TestNew(t *testing.T) {
	m := New()
	if m.current != screenMenu {
		t.Errorf("initial screen = %d, want screenMenu (%d)", m.current, screenMenu)
	}
	if m.width != 80 || m.height != 24 {
		t.Errorf("dimensions = %dx%d, want 80x24", m.width, m.height)
	}
}

func TestInit(t *testing.T) {
	m := New()
	cmd := m.Init()
	if cmd != nil {
		t.Error("expected Init() to return nil")
	}
}

func TestViewMenu(t *testing.T) {
	m := New()
	v := m.View()
	if v.WindowTitle != "gruyere" {
		t.Errorf("WindowTitle = %q, want %q", v.WindowTitle, "gruyere")
	}
	if !v.AltScreen {
		t.Error("expected AltScreen = true")
	}
}

func TestUpdateWindowSize(t *testing.T) {
	m := New()
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}

	result, cmd := m.Update(msg)
	m2 := result.(Model)

	if m2.width != 120 || m2.height != 40 {
		t.Errorf("dimensions = %dx%d, want 120x40", m2.width, m2.height)
	}
	if cmd != nil {
		t.Error("expected nil cmd from window size update")
	}
}

func TestUpdateWindowSizeOnBranchInput(t *testing.T) {
	m := New()
	m.current = screenBranchInput
	m.branchInput = screens.NewBranchInput("/tmp", "repo")

	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	result, _ := m.Update(msg)
	m2 := result.(Model)

	if m2.width != 100 {
		t.Errorf("width = %d, want 100", m2.width)
	}
}

func TestUpdateWindowSizeOnProgress(t *testing.T) {
	m := New()
	m.current = screenProgress
	m.progress = screens.NewProgress("main", "feat")

	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	result, _ := m.Update(msg)
	m2 := result.(Model)

	if m2.width != 100 {
		t.Errorf("width = %d, want 100", m2.width)
	}
}

func TestUpdateWindowSizeOnResults(t *testing.T) {
	m := New()
	m.current = screenResults
	m.results = screens.NewResults("md", "", true, nil, 80, 24)

	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	result, _ := m.Update(msg)
	m2 := result.(Model)

	if m2.width != 100 {
		t.Errorf("width = %d, want 100", m2.width)
	}
}

func TestCtrlCQuits(t *testing.T) {
	m := New()
	msg := tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}

	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	result := cmd()
	if _, ok := result.(tea.QuitMsg); !ok {
		t.Errorf("expected QuitMsg, got %T", result)
	}
}

func TestQuitOnMenu(t *testing.T) {
	m := New()
	m.current = screenMenu
	msg := keyPress('q')

	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestQuitNotOnOtherScreens(t *testing.T) {
	m := New()
	m.current = screenBranchInput
	m.branchInput = screens.NewBranchInput("/tmp", "repo")
	msg := keyPress('q')

	result, _ := m.Update(msg)
	m2 := result.(Model)
	if m2.current != screenBranchInput {
		t.Error("q should not quit on non-menu screens")
	}
}

func TestMenuSelectionChangelog(t *testing.T) {
	m := New()
	msg := curd.MenuSelectionMsg{Command: "changelog"}

	result, cmd := m.Update(msg)
	m2 := result.(Model)

	if m2.current != screenRepoSelect {
		t.Errorf("current = %d, want screenRepoSelect (%d)", m2.current, screenRepoSelect)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd for repo select init")
	}
}

func TestMenuSelectionUnknown(t *testing.T) {
	m := New()
	msg := curd.MenuSelectionMsg{Command: "unknown"}

	result, cmd := m.Update(msg)
	m2 := result.(Model)

	if m2.current != screenMenu {
		t.Errorf("current = %d, want screenMenu (%d)", m2.current, screenMenu)
	}
	if cmd != nil {
		t.Error("expected nil cmd for unknown command")
	}
}

func TestStartChangelogMsg(t *testing.T) {
	m := New()
	msg := screens.StartChangelogMsg{
		BaseBranch:    "main",
		FeatureBranch: "feat",
		RepoPath:      "/tmp",
	}

	result, cmd := m.Update(msg)
	m2 := result.(Model)

	if m2.current != screenProgress {
		t.Errorf("current = %d, want screenProgress (%d)", m2.current, screenProgress)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd (batch)")
	}
}

func TestChangelogDoneMsg(t *testing.T) {
	m := New()
	msg := screens.ChangelogDoneMsg{
		Markdown:    "# Changelog",
		HasChanges:  true,
		OpenAPIPath: "api.yaml",
	}

	result, _ := m.Update(msg)
	m2 := result.(Model)

	if m2.current != screenResults {
		t.Errorf("current = %d, want screenResults (%d)", m2.current, screenResults)
	}
}

func TestChangelogDoneMsgWithError(t *testing.T) {
	m := New()
	msg := screens.ChangelogDoneMsg{
		Err: fmt.Errorf("something failed"),
	}

	result, _ := m.Update(msg)
	m2 := result.(Model)

	if m2.current != screenResults {
		t.Errorf("current = %d, want screenResults (%d)", m2.current, screenResults)
	}
}

func TestBackToMenuMsg(t *testing.T) {
	m := New()
	m.current = screenResults
	m.results = screens.NewResults("md", "", true, nil, 80, 24)

	msg := screens.BackToMenuMsg{}
	result, cmd := m.Update(msg)
	m2 := result.(Model)

	if m2.current != screenMenu {
		t.Errorf("current = %d, want screenMenu (%d)", m2.current, screenMenu)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd (window size)")
	}
}

func TestRepoSelectDoneMsgEmpty(t *testing.T) {
	m := New()
	msg := screens.RepoSelectDoneMsg{Paths: nil}

	result, cmd := m.Update(msg)
	m2 := result.(Model)

	// Should not change screen if no paths selected
	if m2.current != screenMenu {
		t.Errorf("current = %d, want screenMenu", m2.current)
	}
	if cmd != nil {
		t.Error("expected nil cmd for empty repo selection")
	}
}

func TestRepoSelectDoneMsg(t *testing.T) {
	m := New()
	msg := screens.RepoSelectDoneMsg{Paths: []string{"/tmp/my-repo"}}

	result, cmd := m.Update(msg)
	m2 := result.(Model)

	if m2.current != screenBranchInput {
		t.Errorf("current = %d, want screenBranchInput (%d)", m2.current, screenBranchInput)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd for branch input init")
	}
}

func TestViewAllScreens(t *testing.T) {
	tests := []struct {
		name   string
		screen screen
		setup  func(m *Model)
	}{
		{"menu", screenMenu, nil},
		{"branchInput", screenBranchInput, func(m *Model) {
			m.branchInput = screens.NewBranchInput("/tmp", "repo")
		}},
		{"progress", screenProgress, func(m *Model) {
			m.progress = screens.NewProgress("main", "feat")
		}},
		{"results", screenResults, func(m *Model) {
			m.results = screens.NewResults("md", "", true, nil, 80, 24)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.current = tt.screen
			if tt.setup != nil {
				tt.setup(&m)
			}

			v := m.View()
			if v.WindowTitle != "gruyere" {
				t.Errorf("WindowTitle = %q, want %q", v.WindowTitle, "gruyere")
			}
		})
	}
}

func TestDelegateToCurrentScreen(t *testing.T) {
	tests := []struct {
		name   string
		screen screen
		setup  func(m *Model)
	}{
		{"menu", screenMenu, nil},
		{"branchInput", screenBranchInput, func(m *Model) {
			m.branchInput = screens.NewBranchInput("/tmp", "repo")
		}},
		{"progress", screenProgress, func(m *Model) {
			m.progress = screens.NewProgress("main", "feat")
		}},
		{"results", screenResults, func(m *Model) {
			m.results = screens.NewResults("md", "", true, nil, 80, 24)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.current = tt.screen
			if tt.setup != nil {
				tt.setup(&m)
			}
			result, _ := m.Update(nil)
			if result == nil {
				t.Error("expected non-nil result from Update")
			}
		})
	}
}

func TestUpdateWindowSizeOnRepoSelect(t *testing.T) {
	m := New()
	m.current = screenRepoSelect
	m.repoSelect = screens.NewRepoSelect("changelog", ".", 0, 24)

	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	result, _ := m.Update(msg)
	m2 := result.(Model)

	if m2.width != 100 {
		t.Errorf("width = %d, want 100", m2.width)
	}
}

func TestGenerateChangelogCmd_Error(t *testing.T) {
	// Test that generateChangelog returns a valid tea.Cmd
	cmd := generateChangelog("main", "feature", "/nonexistent")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}

	// Execute the cmd — it will fail since the path doesn't exist
	result := cmd()
	msg, ok := result.(screens.ChangelogDoneMsg)
	if !ok {
		t.Fatalf("expected ChangelogDoneMsg, got %T", result)
	}
	if msg.Err == nil {
		t.Error("expected error for nonexistent repo path")
	}
}

func TestGenerateChangelogCmd_Success(t *testing.T) {
	dir := t.TempDir()

	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, out)
		}
	}

	runGit("init", "-b", "main")
	runGit("config", "user.email", "test@test.com")
	runGit("config", "user.name", "Test")

	apiDir := filepath.Join(dir, "api")
	os.MkdirAll(apiDir, 0o755)
	os.WriteFile(filepath.Join(apiDir, "openapi.yaml"), []byte(`openapi: "3.0.0"
paths:
  /users:
    get:
      tags: [Users]
components:
  schemas:
    User:
      properties:
        id:
          type: integer
`), 0o644)
	runGit("add", ".")
	runGit("commit", "-m", "init")

	runGit("checkout", "-b", "feature")
	os.WriteFile(filepath.Join(apiDir, "openapi.yaml"), []byte(`openapi: "3.0.0"
paths:
  /users:
    get:
      tags: [Users]
    post:
      tags: [Users]
components:
  schemas:
    User:
      properties:
        id:
          type: integer
        email:
          type: string
`), 0o644)
	runGit("add", ".")
	runGit("commit", "-m", "add endpoint")

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(origDir) })

	cmd := generateChangelog("main", "feature", dir)
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}

	result := cmd()
	msg, ok := result.(screens.ChangelogDoneMsg)
	if !ok {
		t.Fatalf("expected ChangelogDoneMsg, got %T", result)
	}
	if msg.Err != nil {
		t.Fatalf("unexpected error: %v", msg.Err)
	}
	if !msg.HasChanges {
		t.Error("expected HasChanges = true")
	}
	if msg.Markdown == "" {
		t.Error("expected non-empty markdown")
	}
}

func TestViewRepoSelect(t *testing.T) {
	m := New()
	m.current = screenRepoSelect
	m.repoSelect = screens.NewRepoSelect("changelog", ".", 0, 24)

	v := m.View()
	if v.WindowTitle != "gruyere" {
		t.Errorf("WindowTitle = %q, want %q", v.WindowTitle, "gruyere")
	}
}

func TestDelegateToRepoSelect(t *testing.T) {
	m := New()
	m.current = screenRepoSelect
	m.repoSelect = screens.NewRepoSelect("changelog", ".", 0, 24)
	result, _ := m.Update(nil)
	if result == nil {
		t.Error("expected non-nil result from Update")
	}
}

func TestBackToMenuMsgCmdProducesWindowSize(t *testing.T) {
	m := New()
	m.current = screenResults
	m.results = screens.NewResults("md", "", true, nil, 80, 24)
	m.width = 120
	m.height = 40

	msg := screens.BackToMenuMsg{}
	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	result := cmd()
	wsm, ok := result.(tea.WindowSizeMsg)
	if !ok {
		t.Fatalf("expected WindowSizeMsg, got %T", result)
	}
	if wsm.Width != 120 || wsm.Height != 40 {
		t.Errorf("size = %dx%d, want 120x40", wsm.Width, wsm.Height)
	}
}

func TestScreenConstants(t *testing.T) {
	if screenMenu != 0 {
		t.Errorf("screenMenu = %d, want 0", screenMenu)
	}
	if screenRepoSelect != 1 {
		t.Errorf("screenRepoSelect = %d, want 1", screenRepoSelect)
	}
	if screenBranchInput != 2 {
		t.Errorf("screenBranchInput = %d, want 2", screenBranchInput)
	}
	if screenProgress != 3 {
		t.Errorf("screenProgress = %d, want 3", screenProgress)
	}
	if screenResults != 4 {
		t.Errorf("screenResults = %d, want 4", screenResults)
	}
}
