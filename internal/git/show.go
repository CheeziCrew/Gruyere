package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Show retrieves file content from a specific git branch.
func Show(branch, filepath string) ([]byte, error) {
	normalized := strings.ReplaceAll(filepath, "\\", "/")
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", branch, normalized))
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git show %s:%s failed: %s", branch, normalized, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, err
	}
	return out, nil
}
