package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestShow(t *testing.T) {
	tests := []struct {
		name      string
		branch    string
		filepath  string
		mockFunc  func(ref string) ([]byte, error)
		wantRef   string
		wantData  string
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "success",
			branch:   "main",
			filepath: "api/openapi.yaml",
			wantRef:  "main:api/openapi.yaml",
			mockFunc: func(ref string) ([]byte, error) {
				return []byte("openapi: 3.0.0"), nil
			},
			wantData: "openapi: 3.0.0",
		},
		{
			name:     "error from git",
			branch:   "main",
			filepath: "missing.yaml",
			wantRef:  "main:missing.yaml",
			mockFunc: func(ref string) ([]byte, error) {
				return nil, errors.New("fatal: path not found")
			},
			wantErr:   true,
			errSubstr: "fatal",
		},
		{
			name:     "backslash normalization",
			branch:   "feature",
			filepath: "api\\v2\\openapi.yaml",
			wantRef:  "feature:api/v2/openapi.yaml",
			mockFunc: func(ref string) ([]byte, error) {
				return []byte("normalized"), nil
			},
			wantData: "normalized",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			original := gitShowCmd
			t.Cleanup(func() { gitShowCmd = original })

			var gotRef string
			gitShowCmd = func(ref string) ([]byte, error) {
				gotRef = ref
				return tc.mockFunc(ref)
			}

			data, err := Show(tc.branch, tc.filepath)

			if gotRef != tc.wantRef {
				t.Errorf("ref = %q, want %q", gotRef, tc.wantRef)
			}

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errSubstr != "" && !contains(err.Error(), tc.errSubstr) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.errSubstr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(data) != tc.wantData {
				t.Errorf("data = %q, want %q", string(data), tc.wantData)
			}
		})
	}
}

func initRepo(t *testing.T, dir string) {
	t.Helper()
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
}

func TestGitShowCmd_DefaultFunc_Success(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	os.MkdirAll(filepath.Join(dir, "api"), 0o755)
	os.WriteFile(filepath.Join(dir, "api", "openapi.yaml"), []byte("openapi: 3.0.0"), 0o644)

	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.CombinedOutput()
	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = dir
	cmd.CombinedOutput()

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(origDir) })

	// Call the real default gitShowCmd — do NOT replace it
	data, err := gitShowCmd("main:api/openapi.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(data), "openapi: 3.0.0") {
		t.Errorf("data = %q, expected to contain openapi spec", string(data))
	}
}

func TestGitShowCmd_DefaultFunc_ExitError(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0o644)
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.CombinedOutput()
	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = dir
	cmd.CombinedOutput()

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(origDir) })

	// Call the real default gitShowCmd with a nonexistent file — exercises ExitError path
	_, err := gitShowCmd("main:nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "git show") {
		t.Errorf("error = %q, expected it to contain 'git show'", err.Error())
	}
}

func TestGitShowCmd_DefaultFunc_NonExitError(t *testing.T) {
	// Test the non-ExitError path by temporarily replacing PATH so git can't be found
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	t.Cleanup(func() { os.Setenv("PATH", origPath) })

	_, err := gitShowCmd("main:file.yaml")
	if err == nil {
		t.Fatal("expected error when git is not in PATH")
	}
	// This should be a non-ExitError (exec.ErrNotFound or similar)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
