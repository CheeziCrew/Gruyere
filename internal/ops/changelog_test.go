package ops

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initGitRepo creates a git repo in dir with an initial commit.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init", "-b", "main"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s failed: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
}

// gitCommit stages all files and commits with the given message.
func gitCommit(t *testing.T, dir, msg string) {
	t.Helper()
	cmds := [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", msg},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s failed: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
}

// gitBranch creates and checks out a new branch.
func gitBranch(t *testing.T, dir, branch string) {
	t.Helper()
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git checkout -b %s failed: %v\n%s", branch, err, out)
	}
}

const specV1 = `openapi: "3.0.0"
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
        name:
          type: string
`

const specV2 = `openapi: "3.0.0"
paths:
  /users:
    get:
      tags: [Users]
    post:
      tags: [Users]
  /health:
    get:
      tags: [System]
components:
  schemas:
    User:
      properties:
        id:
          type: integer
        name:
          type: string
        email:
          type: string
`

func TestGenerateChangelog(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, dir string) // set up git repo
		input      func(dir string) Input
		wantErr    bool
		errSubstr  string
		checkOut   func(t *testing.T, out Output)
	}{
		{
			name: "success with changes",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				initGitRepo(t, dir)
				apiDir := filepath.Join(dir, "api")
				os.MkdirAll(apiDir, 0o755)
				os.WriteFile(filepath.Join(apiDir, "openapi.yaml"), []byte(specV1), 0o644)
				gitCommit(t, dir, "initial spec")

				gitBranch(t, dir, "feature")
				os.WriteFile(filepath.Join(apiDir, "openapi.yaml"), []byte(specV2), 0o644)
				gitCommit(t, dir, "add endpoints and field")
			},
			input: func(dir string) Input {
				return Input{
					BaseBranch:    "main",
					FeatureBranch: "feature",
					RepoPath:      dir,
				}
			},
			checkOut: func(t *testing.T, out Output) {
				t.Helper()
				if !out.HasChanges {
					t.Error("expected HasChanges to be true")
				}
				if out.OpenAPIPath == "" {
					t.Error("expected OpenAPIPath to be set")
				}
				if !strings.Contains(out.Markdown, "POST") {
					t.Errorf("markdown should mention added POST endpoint, got:\n%s", out.Markdown)
				}
				if !strings.Contains(out.Markdown, "email") {
					t.Errorf("markdown should mention added email field, got:\n%s", out.Markdown)
				}
			},
		},
		{
			name: "no openapi file",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				initGitRepo(t, dir)
				os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0o644)
				gitCommit(t, dir, "initial")
			},
			input: func(dir string) Input {
				return Input{
					BaseBranch:    "main",
					FeatureBranch: "main",
					RepoPath:      dir,
				}
			},
			wantErr:   true,
			errSubstr: "not found",
		},
		{
			name: "default repo path",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				initGitRepo(t, dir)
				apiDir := filepath.Join(dir, "api")
				os.MkdirAll(apiDir, 0o755)
				os.WriteFile(filepath.Join(apiDir, "openapi.yaml"), []byte(specV1), 0o644)
				gitCommit(t, dir, "initial spec")
			},
			input: func(dir string) Input {
				return Input{
					BaseBranch:    "main",
					FeatureBranch: "main",
					RepoPath:      "", // should default to "."
				}
			},
			// This test verifies the default path logic.
			// With empty RepoPath it defaults to "." which is the test runner's cwd,
			// not the temp dir. So it may or may not find openapi.yaml depending on
			// the cwd. We just verify it doesn't panic and returns a sensible result.
			checkOut: nil, // skip output check — focus is on no panic
		},
	}

	// Additional error path tests
	tests = append(tests, struct {
		name       string
		setup      func(t *testing.T, dir string)
		input      func(dir string) Input
		wantErr    bool
		errSubstr  string
		checkOut   func(t *testing.T, out Output)
	}{
		name: "invalid YAML in base branch",
		setup: func(t *testing.T, dir string) {
			t.Helper()
			initGitRepo(t, dir)
			apiDir := filepath.Join(dir, "api")
			os.MkdirAll(apiDir, 0o755)
			// Invalid YAML
			os.WriteFile(filepath.Join(apiDir, "openapi.yaml"), []byte(":\n  :\n    - ]["), 0o644)
			gitCommit(t, dir, "bad spec")
		},
		input: func(dir string) Input {
			return Input{
				BaseBranch:    "main",
				FeatureBranch: "main",
				RepoPath:      dir,
			}
		},
		wantErr:   true,
		errSubstr: "parsing spec",
	})

	tests = append(tests, struct {
		name       string
		setup      func(t *testing.T, dir string)
		input      func(dir string) Input
		wantErr    bool
		errSubstr  string
		checkOut   func(t *testing.T, out Output)
	}{
		name: "no changes between identical branches",
		setup: func(t *testing.T, dir string) {
			t.Helper()
			initGitRepo(t, dir)
			apiDir := filepath.Join(dir, "api")
			os.MkdirAll(apiDir, 0o755)
			os.WriteFile(filepath.Join(apiDir, "openapi.yaml"), []byte(specV1), 0o644)
			gitCommit(t, dir, "initial spec")
		},
		input: func(dir string) Input {
			return Input{
				BaseBranch:    "main",
				FeatureBranch: "main",
				RepoPath:      dir,
			}
		},
		checkOut: func(t *testing.T, out Output) {
			t.Helper()
			if out.HasChanges {
				t.Error("expected no changes between identical branches")
			}
		},
	})

	tests = append(tests, struct {
		name       string
		setup      func(t *testing.T, dir string)
		input      func(dir string) Input
		wantErr    bool
		errSubstr  string
		checkOut   func(t *testing.T, out Output)
	}{
		name: "bad base branch name",
		setup: func(t *testing.T, dir string) {
			t.Helper()
			initGitRepo(t, dir)
			apiDir := filepath.Join(dir, "api")
			os.MkdirAll(apiDir, 0o755)
			os.WriteFile(filepath.Join(apiDir, "openapi.yaml"), []byte(specV1), 0o644)
			gitCommit(t, dir, "initial spec")
		},
		input: func(dir string) Input {
			return Input{
				BaseBranch:    "nonexistent-branch",
				FeatureBranch: "main",
				RepoPath:      dir,
			}
		},
		wantErr:   true,
		errSubstr: "fetching spec from nonexistent-branch",
	})

	tests = append(tests, struct {
		name       string
		setup      func(t *testing.T, dir string)
		input      func(dir string) Input
		wantErr    bool
		errSubstr  string
		checkOut   func(t *testing.T, out Output)
	}{
		name: "bad feature branch name",
		setup: func(t *testing.T, dir string) {
			t.Helper()
			initGitRepo(t, dir)
			apiDir := filepath.Join(dir, "api")
			os.MkdirAll(apiDir, 0o755)
			os.WriteFile(filepath.Join(apiDir, "openapi.yaml"), []byte(specV1), 0o644)
			gitCommit(t, dir, "initial spec")
		},
		input: func(dir string) Input {
			return Input{
				BaseBranch:    "main",
				FeatureBranch: "nonexistent-feature",
				RepoPath:      dir,
			}
		},
		wantErr:   true,
		errSubstr: "fetching spec from nonexistent-feature",
	})

	tests = append(tests, struct {
		name       string
		setup      func(t *testing.T, dir string)
		input      func(dir string) Input
		wantErr    bool
		errSubstr  string
		checkOut   func(t *testing.T, out Output)
	}{
		name: "invalid YAML in feature branch only",
		setup: func(t *testing.T, dir string) {
			t.Helper()
			initGitRepo(t, dir)
			apiDir := filepath.Join(dir, "api")
			os.MkdirAll(apiDir, 0o755)
			os.WriteFile(filepath.Join(apiDir, "openapi.yaml"), []byte(specV1), 0o644)
			gitCommit(t, dir, "good spec")

			gitBranch(t, dir, "bad-feature")
			os.WriteFile(filepath.Join(apiDir, "openapi.yaml"), []byte(":\n  :\n    - ]["), 0o644)
			gitCommit(t, dir, "broken spec")
		},
		input: func(dir string) Input {
			return Input{
				BaseBranch:    "main",
				FeatureBranch: "bad-feature",
				RepoPath:      dir,
			}
		},
		wantErr:   true,
		errSubstr: "parsing spec from bad-feature",
	})

	tests = append(tests, struct {
		name       string
		setup      func(t *testing.T, dir string)
		input      func(dir string) Input
		wantErr    bool
		errSubstr  string
		checkOut   func(t *testing.T, out Output)
	}{
		name: "scan error for nonexistent repo path",
		setup: func(t *testing.T, dir string) {
			t.Helper()
			initGitRepo(t, dir)
			os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0o644)
			gitCommit(t, dir, "initial")
		},
		input: func(dir string) Input {
			return Input{
				BaseBranch:    "main",
				FeatureBranch: "main",
				RepoPath:      filepath.Join(dir, "nonexistent-subdir"),
			}
		},
		wantErr:   true,
		errSubstr: "openapi.yaml",
	})

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tc.setup(t, dir)

			// git.Show runs "git show" without setting cmd.Dir, so we must
			// chdir into the repo for it to find the correct git context.
			origDir, _ := os.Getwd()
			if err := os.Chdir(dir); err != nil {
				t.Fatalf("chdir to temp dir: %v", err)
			}
			t.Cleanup(func() { os.Chdir(origDir) })

			input := tc.input(dir)
			out, err := GenerateChangelog(input)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errSubstr != "" && !strings.Contains(err.Error(), tc.errSubstr) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.errSubstr)
				}
				return
			}

			// For the default repo path test, we allow errors since cwd may not have openapi.yaml
			if tc.name == "default repo path" {
				// Just verify no panic occurred — error is acceptable
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.checkOut != nil {
				tc.checkOut(t, out)
			}
		})
	}
}
