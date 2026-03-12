package ops

import (
	"fmt"

	"github.com/CheeziCrew/gruyere/internal/git"
	"github.com/CheeziCrew/gruyere/internal/openapi"
)

// Input holds the parameters for changelog generation.
type Input struct {
	BaseBranch    string
	FeatureBranch string
	RepoPath      string // defaults to "."
}

// Output holds the result of changelog generation.
type Output struct {
	Markdown    string
	HasChanges  bool
	OpenAPIPath string
}

// GenerateChangelog finds the OpenAPI spec, fetches from both branches, diffs, and renders markdown.
func GenerateChangelog(input Input) (Output, error) {
	repoPath := input.RepoPath
	if repoPath == "" {
		repoPath = "."
	}

	files, err := git.FindOpenAPIFiles(repoPath)
	if err != nil {
		return Output{}, fmt.Errorf("scanning for openapi.yaml: %w", err)
	}
	if len(files) == 0 {
		return Output{}, fmt.Errorf("openapi.yaml not found in %s", repoPath)
	}

	openAPIPath := files[0]

	oldContent, err := git.Show(input.BaseBranch, openAPIPath)
	if err != nil {
		return Output{}, fmt.Errorf("fetching spec from %s: %w", input.BaseBranch, err)
	}

	newContent, err := git.Show(input.FeatureBranch, openAPIPath)
	if err != nil {
		return Output{}, fmt.Errorf("fetching spec from %s: %w", input.FeatureBranch, err)
	}

	oldSpec, err := openapi.ParseSpec(oldContent)
	if err != nil {
		return Output{}, fmt.Errorf("parsing spec from %s: %w", input.BaseBranch, err)
	}

	newSpec, err := openapi.ParseSpec(newContent)
	if err != nil {
		return Output{}, fmt.Errorf("parsing spec from %s: %w", input.FeatureBranch, err)
	}

	result := openapi.Diff(oldSpec, newSpec, input.BaseBranch, input.FeatureBranch, openAPIPath)
	markdown := openapi.RenderMarkdown(result)

	return Output{
		Markdown:    markdown,
		HasChanges:  result.HasChanges,
		OpenAPIPath: openAPIPath,
	}, nil
}
