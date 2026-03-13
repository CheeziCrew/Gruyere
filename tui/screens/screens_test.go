package screens

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

func keyPress(ch rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: ch, Text: string(ch)}
}

func specialKey(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}

// ── Menu tests ──────────────────────────────────────────────────────

func TestNewMenu(t *testing.T) {
	m := NewMenu()
	v := m.View()
	if v == "" {
		t.Error("NewMenu().View() returned empty string")
	}
}

// ── BranchInput tests ───────────────────────────────────────────────

func TestNewBranchInput(t *testing.T) {
	m := NewBranchInput("/tmp/repo", "my-repo")
	if m.repoPath != "/tmp/repo" {
		t.Errorf("repoPath = %q, want %q", m.repoPath, "/tmp/repo")
	}
	if m.repoName != "my-repo" {
		t.Errorf("repoName = %q, want %q", m.repoName, "my-repo")
	}
	if !m.focusBase {
		t.Error("expected focusBase to be true initially")
	}
	v := m.View()
	if v == "" {
		t.Error("NewBranchInput().View() returned empty string")
	}
}

func TestBranchInputInit(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	cmd := m.Init()
	if cmd == nil {
		t.Error("expected non-nil cmd from Init (textinput.Blink)")
	}
}

func TestBranchInputUpdateWindowSize(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	m2, _ := m.Update(msg)
	if m2.width != 120 || m2.height != 40 {
		t.Errorf("dimensions = %dx%d, want 120x40", m2.width, m2.height)
	}
}

func TestBranchInputEscReturnsBackToMenu(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	m2, cmd := m.Update(specialKey(tea.KeyEscape))
	_ = m2
	if cmd == nil {
		t.Fatal("expected cmd from esc")
	}
	result := cmd()
	if _, ok := result.(BackToMenuMsg); !ok {
		t.Errorf("expected BackToMenuMsg, got %T", result)
	}
}

func TestBranchInputTabSwitchesFocus(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	if !m.focusBase {
		t.Fatal("expected focusBase initially")
	}

	// Tab should move to feature
	m2, cmd := m.Update(specialKey(tea.KeyTab))
	if m2.focusBase {
		t.Error("expected focusBase = false after tab")
	}
	if cmd == nil {
		t.Error("expected blink cmd")
	}
}

func TestBranchInputDownSwitchesFocus(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	m2, _ := m.Update(specialKey(tea.KeyDown))
	if m2.focusBase {
		t.Error("expected focusBase = false after down")
	}
}

func TestBranchInputShiftTabSwitchesBack(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	// Move to feature first
	m2, _ := m.Update(specialKey(tea.KeyTab))
	if m2.focusBase {
		t.Fatal("should be on feature")
	}

	// Shift+tab should go back to base
	m3, cmd := m2.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if !m3.focusBase {
		t.Error("expected focusBase = true after shift+tab")
	}
	if cmd == nil {
		t.Error("expected blink cmd")
	}
}

func TestBranchInputUpSwitchesBack(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	// Move to feature
	m2, _ := m.Update(specialKey(tea.KeyTab))
	// Up should go back
	m3, _ := m2.Update(specialKey(tea.KeyUp))
	if !m3.focusBase {
		t.Error("expected focusBase = true after up")
	}
}

func TestBranchInputTabNoOpOnFeature(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	m.focusBase = false
	// Tab on feature should not switch (handled = false, falls through)
	m2, _ := m.Update(specialKey(tea.KeyTab))
	if m2.focusBase {
		t.Error("tab on feature should not switch to base")
	}
}

func TestBranchInputShiftTabNoOpOnBase(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	// shift+tab on base should not switch
	m2, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if !m2.focusBase {
		t.Error("shift+tab on base should stay on base")
	}
}

func TestBranchInputEnterOnBase(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	// Enter on base should switch to feature
	m2, cmd := m.Update(specialKey(tea.KeyEnter))
	if m2.focusBase {
		t.Error("expected focus to switch to feature on enter from base")
	}
	if cmd == nil {
		t.Error("expected blink cmd")
	}
}

func TestBranchInputEnterOnFeatureEmpty(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	m.focusBase = false
	// Feature is empty, enter should not trigger start
	m2, cmd := m.Update(specialKey(tea.KeyEnter))
	_ = m2
	// No start msg because feature is empty
	if cmd != nil {
		// cmd may be non-nil if it falls through to textinput update
		// but it should NOT be a StartChangelogMsg
		result := cmd()
		if _, ok := result.(StartChangelogMsg); ok {
			t.Error("should not start with empty feature branch")
		}
	}
}

func TestBranchInputEnterOnFeatureWithValue(t *testing.T) {
	m := NewBranchInput("/tmp/myrepo", "myrepo")
	// Switch to feature
	m.focusBase = false
	m.featureInput.SetValue("feature/x")

	m2, cmd := m.Update(specialKey(tea.KeyEnter))
	_ = m2
	if cmd == nil {
		t.Fatal("expected StartChangelogMsg cmd")
	}
	result := cmd()
	msg, ok := result.(StartChangelogMsg)
	if !ok {
		t.Fatalf("expected StartChangelogMsg, got %T", result)
	}
	if msg.FeatureBranch != "feature/x" {
		t.Errorf("FeatureBranch = %q, want %q", msg.FeatureBranch, "feature/x")
	}
	if msg.BaseBranch != "main" {
		t.Errorf("BaseBranch = %q, want %q", msg.BaseBranch, "main")
	}
	if msg.RepoPath != "/tmp/myrepo" {
		t.Errorf("RepoPath = %q, want %q", msg.RepoPath, "/tmp/myrepo")
	}
}

func TestBranchInputFocusFeature(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	m2 := m.focusFeature()
	if m2.focusBase {
		t.Error("expected focusBase = false")
	}
}

func TestBranchInputFocusBaseField(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	m.focusBase = false
	m2 := m.focusBaseField()
	if !m2.focusBase {
		t.Error("expected focusBase = true")
	}
}

func TestBranchInputDelegatesToFocusedInput(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	// Typing a character on base should delegate to base input
	m2, _ := m.Update(keyPress('x'))
	_ = m2 // Just verify no panic

	// Same for feature
	m.focusBase = false
	m3, _ := m.Update(keyPress('y'))
	_ = m3
}

// ── Progress tests ──────────────────────────────────────────────────

func TestNewProgress(t *testing.T) {
	m := NewProgress("main", "feature/x")
	v := m.View()
	if v == "" {
		t.Error("NewProgress().View() returned empty string")
	}
}

func TestProgressInit(t *testing.T) {
	m := NewProgress("main", "feat")
	cmd := m.Init()
	if cmd == nil {
		t.Error("expected non-nil cmd (spinner tick)")
	}
}

func TestProgressUpdateWindowSize(t *testing.T) {
	m := NewProgress("main", "feat")
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	m2, _ := m.Update(msg)
	if m2.width != 100 || m2.height != 30 {
		t.Errorf("dimensions = %dx%d, want 100x30", m2.width, m2.height)
	}
}

func TestProgressUpdateSpinnerTick(t *testing.T) {
	m := NewProgress("main", "feat")
	// Create a spinner tick msg
	tickMsg := spinner.TickMsg{ID: m.spinner.ID()}
	m2, cmd := m.Update(tickMsg)
	_ = m2
	// Spinner should produce a follow-up tick cmd
	if cmd == nil {
		t.Error("expected follow-up tick cmd")
	}
}

func TestProgressUpdateUnknownMsg(t *testing.T) {
	m := NewProgress("main", "feat")
	m2, cmd := m.Update(nil)
	_ = m2
	if cmd != nil {
		t.Error("expected nil cmd for unknown msg")
	}
}

func TestProgressView(t *testing.T) {
	m := NewProgress("main", "feature/test")
	v := m.View()
	if !strings.Contains(v, "main") {
		t.Error("expected view to contain base branch")
	}
}

// ── Results tests ───────────────────────────────────────────────────

func TestNewResults_NoChanges(t *testing.T) {
	m := NewResults("", "", false, nil, 80, 24)
	v := m.View()
	if v == "" {
		t.Error("NewResults(no changes).View() returned empty string")
	}
	if !m.ready {
		t.Error("expected ready = true")
	}
}

func TestNewResults_WithMarkdown(t *testing.T) {
	m := NewResults("# Changes\n- added endpoint", "/api.yaml", true, nil, 80, 24)
	v := m.View()
	if v == "" {
		t.Error("NewResults(with markdown).View() returned empty string")
	}
}

func TestNewResults_WithError(t *testing.T) {
	m := NewResults("", "", false, fmt.Errorf("something failed"), 80, 24)
	v := m.View()
	if v == "" {
		t.Error("NewResults(with error).View() returned empty string")
	}
}

func TestResultsInit(t *testing.T) {
	m := NewResults("md", "", true, nil, 80, 24)
	cmd := m.Init()
	if cmd != nil {
		t.Error("expected nil cmd from Init")
	}
}

func TestResultsUpdateWindowSize(t *testing.T) {
	m := NewResults("md", "", true, nil, 80, 24)
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	m2, cmd := m.Update(msg)
	if m2.width != 100 || m2.height != 30 {
		t.Errorf("dimensions = %dx%d, want 100x30", m2.width, m2.height)
	}
	if cmd != nil {
		t.Error("expected nil cmd from window size")
	}
}

func TestResultsEscReturnsBackToMenu(t *testing.T) {
	m := NewResults("md", "", true, nil, 80, 24)
	_, cmd := m.Update(specialKey(tea.KeyEscape))
	if cmd == nil {
		t.Fatal("expected cmd from esc")
	}
	result := cmd()
	if _, ok := result.(BackToMenuMsg); !ok {
		t.Errorf("expected BackToMenuMsg, got %T", result)
	}
}

func TestResultsQReturnsBackToMenu(t *testing.T) {
	m := NewResults("md", "", true, nil, 80, 24)
	_, cmd := m.Update(keyPress('q'))
	if cmd == nil {
		t.Fatal("expected cmd from q")
	}
	result := cmd()
	if _, ok := result.(BackToMenuMsg); !ok {
		t.Errorf("expected BackToMenuMsg, got %T", result)
	}
}

func TestResultsWriteChangelog(t *testing.T) {
	// Change to temp dir so changelog.md is written there
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	t.Cleanup(func() { os.Chdir(origDir) })

	m := NewResults("# My Changes", "/api.yaml", true, nil, 80, 24)
	m2, cmd := m.Update(keyPress('w'))
	if cmd != nil {
		t.Error("expected nil cmd from write (synchronous)")
	}
	if m2.flash != "Written to changelog.md" {
		t.Errorf("flash = %q", m2.flash)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "changelog.md"))
	if err != nil {
		t.Fatalf("failed to read changelog.md: %v", err)
	}
	if string(content) != "# My Changes" {
		t.Errorf("content = %q", string(content))
	}
}

func TestResultsWriteNoChanges(t *testing.T) {
	m := NewResults("", "", false, nil, 80, 24)
	m2, _ := m.Update(keyPress('w'))
	// Should not write when hasChanges is false
	if m2.flash != "" {
		t.Errorf("flash = %q, expected empty (no changes)", m2.flash)
	}
}

func TestResultsCopyNoChanges(t *testing.T) {
	m := NewResults("", "", false, nil, 80, 24)
	m2, cmd := m.Update(keyPress('y'))
	_ = m2
	// Should not copy when hasChanges is false
	if cmd != nil {
		t.Error("expected nil cmd when no changes")
	}
}

func TestResultsCopyWithChanges(t *testing.T) {
	m := NewResults("# Changes", "", true, nil, 80, 24)
	m2, cmd := m.Update(keyPress('y'))
	if m2.flash != "Copied to clipboard" {
		t.Errorf("flash = %q", m2.flash)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd for clipboard copy")
	}
}

func TestResultsViewNotReady(t *testing.T) {
	m := ResultsModel{ready: false}
	v := m.View()
	if v != "Loading..." {
		t.Errorf("view = %q, want Loading...", v)
	}
}

func TestResultsViewWithOpenAPIPath(t *testing.T) {
	m := NewResults("# Changes", "/api/openapi.yaml", true, nil, 80, 24)
	v := m.View()
	if !strings.Contains(v, "API Changelog") {
		t.Error("expected view to contain title")
	}
}

func TestResultsViewWithFlash(t *testing.T) {
	m := NewResults("# Changes", "", true, nil, 80, 24)
	m.flash = "Test flash"
	v := m.View()
	_ = v // just verify no panic
}

func TestResultsInitViewportSmallSize(t *testing.T) {
	// Test with very small dimensions to hit min-size guards
	m := NewResults("md", "", true, nil, 30, 10)
	if !m.ready {
		t.Error("expected ready = true")
	}
}

func TestResultsDelegateToViewport(t *testing.T) {
	m := NewResults("# Changes\nLine1\nLine2\nLine3", "", true, nil, 80, 24)
	// Send a random key that viewport might handle
	m2, _ := m.Update(keyPress('j'))
	_ = m2 // just verify no panic
}

// ── RepoSelect tests ────────────────────────────────────────────────

func TestScanRepos(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid git repo
	repoDir := filepath.Join(tmpDir, "my-service")
	os.MkdirAll(filepath.Join(repoDir, ".git"), 0755)

	// Create a non-git dir
	os.MkdirAll(filepath.Join(tmpDir, "not-a-repo"), 0755)

	// Create a hidden dir (should be skipped)
	os.MkdirAll(filepath.Join(tmpDir, ".hidden", ".git"), 0755)

	// Create a file (not a dir, should be skipped)
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("hello"), 0644)

	repos, err := scanRepos(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Name != "my-service" {
		t.Errorf("name = %q, want %q", repos[0].Name, "my-service")
	}
	if repos[0].Path != repoDir {
		t.Errorf("path = %q, want %q", repos[0].Path, repoDir)
	}
}

func TestScanReposEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	repos, err := scanRepos(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("expected 0 repos, got %d", len(repos))
	}
}

func TestScanReposInvalidPath(t *testing.T) {
	_, err := scanRepos("/nonexistent/path")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestScanReposSorted(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"z-repo", "a-repo", "m-repo"} {
		os.MkdirAll(filepath.Join(tmpDir, name, ".git"), 0755)
	}

	repos, err := scanRepos(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 3 {
		t.Fatalf("expected 3 repos, got %d", len(repos))
	}
	if repos[0].Name != "a-repo" {
		t.Errorf("first repo = %q, want a-repo", repos[0].Name)
	}
	if repos[1].Name != "m-repo" {
		t.Errorf("second repo = %q, want m-repo", repos[1].Name)
	}
	if repos[2].Name != "z-repo" {
		t.Errorf("third repo = %q, want z-repo", repos[2].Name)
	}
}

func TestGetGitBranch(t *testing.T) {
	// Non-git directory should return empty string
	tmpDir := t.TempDir()
	branch := getGitBranch(tmpDir)
	if branch != "" {
		t.Errorf("expected empty branch for non-git dir, got %q", branch)
	}
}

// ── Message type tests ──────────────────────────────────────────────

func TestStartChangelogMsgFields(t *testing.T) {
	msg := StartChangelogMsg{
		BaseBranch:    "main",
		FeatureBranch: "feat",
		RepoPath:      "/tmp",
	}
	if msg.BaseBranch != "main" {
		t.Error("BaseBranch mismatch")
	}
	if msg.FeatureBranch != "feat" {
		t.Error("FeatureBranch mismatch")
	}
}

func TestChangelogDoneMsgFields(t *testing.T) {
	msg := ChangelogDoneMsg{
		Markdown:    "# Changes",
		HasChanges:  true,
		OpenAPIPath: "api.yaml",
		Err:         nil,
	}
	if msg.Markdown != "# Changes" {
		t.Error("Markdown mismatch")
	}
}

func TestClipboardMsgType(t *testing.T) {
	msg := clipboardMsg{err: nil}
	if msg.err != nil {
		t.Error("expected nil error")
	}
	msg2 := clipboardMsg{err: fmt.Errorf("fail")}
	if msg2.err == nil {
		t.Error("expected non-nil error")
	}
}

func TestResultsUpdateClipboardMsg(t *testing.T) {
	m := NewResults("# Changes", "", true, nil, 80, 24)
	m2, cmd := m.Update(clipboardMsg{err: nil})
	_ = m2
	if cmd != nil {
		t.Error("expected nil cmd from clipboardMsg")
	}
}

func TestResultsUpdateClipboardMsgError(t *testing.T) {
	m := NewResults("# Changes", "", true, nil, 80, 24)
	m2, cmd := m.Update(clipboardMsg{err: fmt.Errorf("clipboard fail")})
	_ = m2
	if cmd != nil {
		t.Error("expected nil cmd from clipboardMsg with error")
	}
}

func TestCopyToClipboard(t *testing.T) {
	cmd := copyToClipboard("test text")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	// Execute the cmd — it may fail if pbcopy is not available, but it should
	// return a clipboardMsg either way
	result := cmd()
	if _, ok := result.(clipboardMsg); !ok {
		t.Errorf("expected clipboardMsg, got %T", result)
	}
}

func TestGetGitBranch_RealRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Init a real git repo
	cmds := [][]string{
		{"git", "init", "-b", "test-branch"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
	}

	// Need at least one commit for rev-parse to work
	readmePath := filepath.Join(tmpDir, "README.md")
	os.WriteFile(readmePath, []byte("test"), 0644)
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = tmpDir
	addCmd.CombinedOutput()
	commitCmd := exec.Command("git", "commit", "-m", "init")
	commitCmd.Dir = tmpDir
	commitCmd.CombinedOutput()

	branch := getGitBranch(tmpDir)
	if branch != "test-branch" {
		t.Errorf("branch = %q, want %q", branch, "test-branch")
	}
}

func TestNewRepoSelect(t *testing.T) {
	m := NewRepoSelect("changelog", ".", 0, 24)
	v := m.View()
	if v == "" {
		t.Error("NewRepoSelect().View() returned empty string")
	}
}

func TestProgressViewContainsFeature(t *testing.T) {
	m := NewProgress("develop", "feature/api-v2")
	v := m.View()
	if !strings.Contains(v, "feature/api-v2") {
		t.Error("expected view to contain feature branch name")
	}
	if !strings.Contains(v, "develop") {
		t.Error("expected view to contain base branch name")
	}
}

func TestBranchInputViewContent(t *testing.T) {
	m := NewBranchInput("/tmp/repo", "my-app")
	v := m.View()
	if !strings.Contains(v, "my-app") {
		t.Error("expected view to contain repo name")
	}
	if !strings.Contains(v, "Base branch") {
		t.Error("expected view to contain base branch label")
	}
	if !strings.Contains(v, "Feature branch") {
		t.Error("expected view to contain feature branch label")
	}
}

func TestResultsWriteError(t *testing.T) {
	// Create a results model that tries to write to an impossible path
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	t.Cleanup(func() { os.Chdir(origDir) })

	// Make the directory read-only to cause write failure
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	os.MkdirAll(readOnlyDir, 0755)
	os.Chdir(readOnlyDir)
	os.Chmod(readOnlyDir, 0444)
	t.Cleanup(func() { os.Chmod(readOnlyDir, 0755) })

	m := NewResults("# Changes", "", true, nil, 80, 24)
	m2, _ := m.Update(keyPress('w'))
	if !strings.Contains(m2.flash, "Error") {
		t.Errorf("expected error flash, got %q", m2.flash)
	}
}

func TestMenuInit(t *testing.T) {
	m := NewMenu()
	// MenuModel is a type alias for curd.MenuModel, so Init should work
	cmd := m.Init()
	_ = cmd // just verify no panic
}

func TestBranchInputDownOnFeatureNoOp(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	m.focusBase = false
	m2, _ := m.Update(specialKey(tea.KeyDown))
	// down on feature should not switch to base (tab/down only goes forward)
	if m2.focusBase {
		t.Error("down on feature should not switch to base")
	}
}

func TestBranchInputUpOnBaseNoOp(t *testing.T) {
	m := NewBranchInput("/tmp", "repo")
	m2, _ := m.Update(specialKey(tea.KeyUp))
	// up on base should stay on base
	if !m2.focusBase {
		t.Error("up on base should stay on base")
	}
}
