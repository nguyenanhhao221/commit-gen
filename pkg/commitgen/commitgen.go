// Package commitgen provides AI-powered git commit message generation.
// This package is designed to be used as a core library by various applications
// such as CLI tools, IDE plugins, or git integrations.
package commitgen

import (
	"fmt"
	"os"
)

// CommitGen provides a high-level interface for commit message generation
type CommitGen struct {
	generator *CommitMessageGenerator
	repo      *GitRepository
}

// Options contains configuration options for CommitGen
type Options struct {
	// WorkingDir is the git repository directory (empty for current dir)
	WorkingDir string
	// APIKey for the AI service
	APIKey string
	// Model to use for generation (optional, uses default if empty)
	Model string
}

// New creates a new CommitGen instance
func New(opts *Options) (*CommitGen, error) {
	if opts == nil {
		opts = &Options{}
	}

	// Get API key from options or environment
	apiKey := opts.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key not provided in options or GOOGLE_API_KEY environment variable")
	}

	// Set up generator config
	config := DefaultConfig()
	config.APIKey = apiKey
	if opts.Model != "" {
		config.Model = opts.Model
	}

	// Create generator
	generator, err := NewCommitMessageGenerator(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create generator: %w", err)
	}

	// Create git repository handler
	repo := NewGitRepository(opts.WorkingDir)

	return &CommitGen{
		generator: generator,
		repo:      repo,
	}, nil
}

// Generate creates a commit message for the current staged changes
func (c *CommitGen) Generate() (string, error) {
	// Get git context
	gitInfo, err := c.repo.GetCommitContext()
	if err != nil {
		return "", err
	}

	// Generate commit message
	message, err := c.generator.GenerateCommitMessage(gitInfo)
	if err != nil {
		return "", err
	}

	return message, nil
}

// GenerateFromDiff creates a commit message from provided diff and optional history
// This is useful for applications that want to provide their own git data
func (c *CommitGen) GenerateFromDiff(diff, history string) (string, error) {
	gitInfo := &GitInfo{
		StagedDiff:    diff,
		RecentCommits: history,
		HasHistory:    history != "",
	}

	return c.generator.GenerateCommitMessage(gitInfo)
}

// HasStagedChanges checks if there are staged changes in the repository
func (c *CommitGen) HasStagedChanges() (bool, error) {
	return c.repo.HasStagedChanges()
}

// GetGitInfo returns the git information that would be used for generation
// This is useful for debugging or for applications that want to preview the data
func (c *CommitGen) GetGitInfo() (*GitInfo, error) {
	return c.repo.GetCommitContext()
}

// Close cleans up resources
func (c *CommitGen) Close() error {
	return c.generator.Close()
}

// QuickGenerate is a convenience function for simple use cases
// It creates a CommitGen instance, generates a message, and cleans up
func QuickGenerate(apiKey string) (string, error) {
	commitGen, err := New(&Options{
		APIKey: apiKey,
	})
	if err != nil {
		return "", err
	}
	defer commitGen.Close()

	return commitGen.Generate()
}

// QuickGenerateWithOptions is like QuickGenerate but with more options
func QuickGenerateWithOptions(opts *Options) (string, error) {
	commitGen, err := New(opts)
	if err != nil {
		return "", err
	}
	defer commitGen.Close()

	return commitGen.Generate()
}

