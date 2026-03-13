package git

import "strings"

// Show retrieves file content from a specific git branch.
func Show(branch, filepath string) ([]byte, error) {
	normalized := strings.ReplaceAll(filepath, "\\", "/")
	ref := branch + ":" + normalized
	return gitShowCmd(ref)
}
