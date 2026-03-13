package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// gitShowCmd is the function used to run git show. Replaceable in tests.
var gitShowCmd = func(ref string) ([]byte, error) {
	cmd := exec.Command("git", "show", ref)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git show %s failed: %s", ref, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, err
	}
	return out, nil
}
