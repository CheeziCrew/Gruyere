package git

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func createFile(t *testing.T, root, relPath string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte("openapi: 3.0.0"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFindOpenAPIFiles(t *testing.T) {
	root := t.TempDir()
	createFile(t, root, "api/openapi.yaml")
	createFile(t, root, "service/openapi.yaml")

	got, err := FindOpenAPIFiles(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sort.Strings(got)
	want := []string{"api/openapi.yaml", "service/openapi.yaml"}
	sort.Strings(want)

	if len(got) != len(want) {
		t.Fatalf("got %d files, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestFindOpenAPIFiles_SkipsTarget(t *testing.T) {
	root := t.TempDir()
	createFile(t, root, "target/openapi.yaml")

	got, err := FindOpenAPIFiles(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no files, got %v", got)
	}
}

func TestFindOpenAPIFiles_Empty(t *testing.T) {
	root := t.TempDir()

	got, err := FindOpenAPIFiles(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestFindOpenAPIFiles_NestedTarget(t *testing.T) {
	root := t.TempDir()
	createFile(t, root, "src/target/openapi.yaml")

	got, err := FindOpenAPIFiles(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no files (nested target should be skipped), got %v", got)
	}
}

func TestFindOpenAPIFiles_RelativePaths(t *testing.T) {
	root := t.TempDir()
	createFile(t, root, "deep/nested/dir/openapi.yaml")

	got, err := FindOpenAPIFiles(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 file, got %v", got)
	}
	if strings.Contains(got[0], "\\") {
		t.Errorf("path contains backslash: %q", got[0])
	}
	if got[0] != "deep/nested/dir/openapi.yaml" {
		t.Errorf("got %q, want %q", got[0], "deep/nested/dir/openapi.yaml")
	}
}

func TestFindOpenAPIFiles_IgnoresNonYAML(t *testing.T) {
	root := t.TempDir()
	// Create a file named "openapi.json" — should not be found
	full := filepath.Join(root, "api", "openapi.json")
	os.MkdirAll(filepath.Dir(full), 0o755)
	os.WriteFile(full, []byte("{}"), 0o644)

	got, err := FindOpenAPIFiles(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 files, got %v", got)
	}
}

func TestFindOpenAPIFiles_MultipleFiles(t *testing.T) {
	root := t.TempDir()
	createFile(t, root, "a/openapi.yaml")
	createFile(t, root, "b/openapi.yaml")
	createFile(t, root, "c/deep/openapi.yaml")

	got, err := FindOpenAPIFiles(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 files, got %d: %v", len(got), got)
	}
}

func TestFindOpenAPIFiles_RootFile(t *testing.T) {
	root := t.TempDir()
	createFile(t, root, "openapi.yaml")

	got, err := FindOpenAPIFiles(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 file, got %v", got)
	}
	if got[0] != "openapi.yaml" {
		t.Errorf("got %q, want %q", got[0], "openapi.yaml")
	}
}

func TestFindOpenAPIFiles_PermissionError(t *testing.T) {
	root := t.TempDir()
	// Create a directory with no read permission — walk will pass an error for its children
	badDir := filepath.Join(root, "noperm")
	os.MkdirAll(filepath.Join(badDir, "sub"), 0o755)
	createFile(t, root, "noperm/sub/openapi.yaml")
	// Also create a findable one
	createFile(t, root, "good/openapi.yaml")
	// Remove read permission from bad dir
	os.Chmod(badDir, 0o000)
	t.Cleanup(func() { os.Chmod(badDir, 0o755) })

	got, err := FindOpenAPIFiles(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should still find the good one and skip the permission error
	if len(got) != 1 {
		t.Errorf("expected 1 file, got %v", got)
	}
}
