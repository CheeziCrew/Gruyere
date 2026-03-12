package screens

import "github.com/CheeziCrew/curd"

// Re-export shared message types.
type BackToMenuMsg = curd.BackToMenuMsg
type RepoSelectDoneMsg = curd.RepoSelectDoneMsg

// StartChangelogMsg is sent when the user confirms both branches.
type StartChangelogMsg struct {
	BaseBranch    string
	FeatureBranch string
	RepoPath      string
}

// ChangelogDoneMsg is sent when changelog generation completes.
type ChangelogDoneMsg struct {
	Markdown    string
	HasChanges  bool
	OpenAPIPath string
	Err         error
}
