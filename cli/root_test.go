package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CheeziCrew/gruyere/internal/ops"
)

func TestBuildCLI(t *testing.T) {
	cmd := BuildCLI()
	if cmd.Use != "gruyere <base-branch> <feature-branch>" {
		t.Errorf("Use = %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
	if !cmd.SilenceErrors {
		t.Error("expected SilenceErrors = true")
	}
	if !cmd.SilenceUsage {
		t.Error("expected SilenceUsage = true")
	}

	// Check flags exist
	f := cmd.Flags()
	if f.Lookup("output") == nil {
		t.Error("expected --output flag")
	}
	if f.Lookup("prepend") == nil {
		t.Error("expected --prepend flag")
	}
}

func TestBuildCLI_RequiresExactArgs(t *testing.T) {
	cmd := BuildCLI()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error with no args")
	}
}

func TestWriteOutput_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "out.md")

	err := writeOutput("# Changes\n", path, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	if string(content) != "# Changes\n" {
		t.Errorf("content = %q", string(content))
	}
}

func TestWriteOutput_Prepend(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "out.md")

	os.WriteFile(path, []byte("existing content"), 0644)

	err := writeOutput("new stuff", path, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "new stuff") {
		t.Error("expected new content")
	}
	if !strings.Contains(string(content), "existing content") {
		t.Error("expected existing content preserved")
	}
	// new content should come first
	idx1 := strings.Index(string(content), "new stuff")
	idx2 := strings.Index(string(content), "existing content")
	if idx1 > idx2 {
		t.Error("new content should be prepended before existing")
	}
}

func TestWriteOutput_PrependNoExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nonexistent.md")

	err := writeOutput("first content", path, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "first content") {
		t.Error("expected content written")
	}
}

func TestWriteOutput_ErrorOnBadPath(t *testing.T) {
	err := writeOutput("test", "/nonexistent/dir/file.md", false)
	if err == nil {
		t.Error("expected error writing to invalid path")
	}
}

func TestRunChangelog_Error(t *testing.T) {
	original := generateChangelogFn
	t.Cleanup(func() { generateChangelogFn = original })

	generateChangelogFn = func(input ops.Input) (ops.Output, error) {
		return ops.Output{}, fmt.Errorf("mock error")
	}

	err := runChangelog("main", "feat", "", false)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "mock error" {
		t.Errorf("error = %q, want %q", err.Error(), "mock error")
	}
}

func TestRunChangelog_NoChanges(t *testing.T) {
	original := generateChangelogFn
	t.Cleanup(func() { generateChangelogFn = original })

	generateChangelogFn = func(input ops.Input) (ops.Output, error) {
		return ops.Output{HasChanges: false}, nil
	}

	err := runChangelog("main", "feat", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunChangelog_WithChangesToStdout(t *testing.T) {
	original := generateChangelogFn
	t.Cleanup(func() { generateChangelogFn = original })

	generateChangelogFn = func(input ops.Input) (ops.Output, error) {
		return ops.Output{
			Markdown:   "# Changes\n",
			HasChanges: true,
		}, nil
	}

	err := runChangelog("main", "feat", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunChangelog_WithChangesToFile(t *testing.T) {
	original := generateChangelogFn
	t.Cleanup(func() { generateChangelogFn = original })

	generateChangelogFn = func(input ops.Input) (ops.Output, error) {
		return ops.Output{
			Markdown:   "# Changes\n",
			HasChanges: true,
		}, nil
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "out.md")

	err := runChangelog("main", "feat", outPath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(outPath)
	if string(content) != "# Changes\n" {
		t.Errorf("content = %q", string(content))
	}
}

func TestRunChangelog_WithChangesToFilePrepend(t *testing.T) {
	original := generateChangelogFn
	t.Cleanup(func() { generateChangelogFn = original })

	generateChangelogFn = func(input ops.Input) (ops.Output, error) {
		return ops.Output{
			Markdown:   "# New\n",
			HasChanges: true,
		}, nil
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "out.md")
	os.WriteFile(outPath, []byte("# Old\n"), 0644)

	err := runChangelog("main", "feat", outPath, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(outPath)
	if !strings.Contains(string(content), "# New") || !strings.Contains(string(content), "# Old") {
		t.Errorf("content = %q", string(content))
	}
}

func TestDefaultGenerateChangelogFn(t *testing.T) {
	// Call the default generateChangelogFn without mocking —
	// it will fail since we're not in a git repo, but exercises the default function body.
	_, err := generateChangelogFn(ops.Input{
		BaseBranch:    "main",
		FeatureBranch: "main",
		RepoPath:      t.TempDir(),
	})
	// Expected to error since temp dir has no openapi.yaml
	if err == nil {
		t.Error("expected error from default generateChangelogFn with no openapi files")
	}
}

func TestBuildCLI_ExecuteRunsRunChangelog(t *testing.T) {
	original := generateChangelogFn
	t.Cleanup(func() { generateChangelogFn = original })

	var gotInput ops.Input
	generateChangelogFn = func(input ops.Input) (ops.Output, error) {
		gotInput = input
		return ops.Output{HasChanges: false}, nil
	}

	cmd := BuildCLI()
	cmd.SetArgs([]string{"main", "feature/x"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotInput.BaseBranch != "main" {
		t.Errorf("BaseBranch = %q, want %q", gotInput.BaseBranch, "main")
	}
	if gotInput.FeatureBranch != "feature/x" {
		t.Errorf("FeatureBranch = %q, want %q", gotInput.FeatureBranch, "feature/x")
	}
}
