package screens

import "github.com/CheeziCrew/curd"

// Re-export shared message types.
type BackToMenuMsg = curd.BackToMenuMsg

// ChangelogDoneMsg is sent when changelog generation completes.
type ChangelogDoneMsg struct {
	Markdown    string
	HasChanges  bool
	OpenAPIPath string
	Err         error
}
