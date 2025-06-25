// Package generator provides a high-level interface for commit message generation
package generator

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/genai"
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
	// Use short commit format
	IsShortCommit bool
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
	generator, err := NewCommitMessageGenerator(config, opts.IsShortCommit)
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

// QuickGenerateShort is a convenience function for generating short commit messages
func QuickGenerateShort(apiKey string) (string, error) {
	commitGen, err := New(&Options{
		APIKey:        apiKey,
		IsShortCommit: true,
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

// CommitMessageGenerator handles AI-powered commit message generation
type CommitMessageGenerator struct {
	client        *genai.Client
	config        *GeneratorConfig
	systemPrompt  string
	isShortCommit bool
}

// GeneratorConfig contains configuration for the commit message generator
type GeneratorConfig struct {
	Model   string
	Timeout time.Duration
	APIKey  string
}

// DefaultConfig returns a default configuration
func DefaultConfig() *GeneratorConfig {
	return &GeneratorConfig{
		Model:   "gemini-2.5-flash-lite-preview-06-17", // Fast and Dirty just like we like it
		Timeout: 10 * time.Second,
	}
}

// NewCommitMessageGenerator creates a new commit message generator
func NewCommitMessageGenerator(config *GeneratorConfig, isShortCommit bool) (*CommitMessageGenerator, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  config.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	var systemPrompt string
	if isShortCommit {
		systemPrompt = getShortCommitPrompt()
	} else {
		systemPrompt = getDefaultSystemPrompt()
	}

	return &CommitMessageGenerator{
		client:        client,
		config:        config,
		systemPrompt:  systemPrompt,
		isShortCommit: isShortCommit,
	}, nil
}

// GenerateCommitMessage generates a commit message from git information
func (g *CommitMessageGenerator) GenerateCommitMessage(gitInfo *GitInfo) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.config.Timeout)
	defer cancel()

	// Prepare the prompt
	prompt := buildPrompt(gitInfo)

	// Configure the AI request
	genConfig := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(g.systemPrompt, genai.RoleUser),
		ThinkingConfig: &genai.ThinkingConfig{
			IncludeThoughts: false,
			ThinkingBudget:  func() *int32 { v := int32(0); return &v }(), // Disable thinking
		},
	}

	// Generate the commit message
	result, err := g.client.Models.GenerateContent(
		ctx,
		g.config.Model,
		genai.Text(prompt),
		genConfig,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate commit message: %w", err)
	}

	return result.Text(), nil
}

// Close cleans up resources
func (g *CommitMessageGenerator) Close() error {
	// Add cleanup if needed
	return nil
}

// buildPrompt constructs the prompt for the AI
func buildPrompt(gitInfo *GitInfo) string {
	if gitInfo.HasHistory && gitInfo.RecentCommits != "" {
		return fmt.Sprintf(
			"Recent git log:\n%s\n\nGit diff:\n%s\n",
			gitInfo.RecentCommits,
			gitInfo.StagedDiff,
		)
	}

	// If no history, include default examples
	return fmt.Sprintf(
		"Recent git log:\n%s\n\nGit diff:\n%s\n",
		getDefaultCommitExamples(),
		gitInfo.StagedDiff,
	)
}

// getDefaultSystemPrompt returns the default system prompt
func getDefaultSystemPrompt() string {
	return `You are a git commit message generator. Analyze the provided git diff and recent git log to create a complete commit message with both subject and body.

Format:
- Subject line: type(scope): brief description (max 50 chars)
- Blank line
- Body: Detailed explanation of WHAT, HOW, and WHY (wrap at 72 chars)

Rules for Subject:
1. Use Conventional Commits format: type(scope): description
2. Common types: feat, fix, refactor, chore, docs, style, test, perf, ci, build
3. Keep under 50 characters
4. Use imperative mood (e.g., "add feature" not "added feature")

Rules for Body:
1. Explain WHAT changed (summary of changes)
2. Explain HOW it was implemented (approach/method)
3. Explain WHY it was necessary (motivation/context)
4. Wrap lines at 72 characters
5. Use bullet points for multiple changes
6. Reference issues/tickets if relevant

Example:
feat(auth): add JWT-based user authentication

- Implement JWT token generation and validation
- Add middleware for protecting authenticated routes
- Create user login/logout endpoints with secure session handling

This change enables secure user sessions and replaces the previous
cookie-based authentication which had security vulnerabilities.
The new system provides better scalability and follows industry
best practices for API authentication.

Match the style and tone of recent commits in the git log.
Output only the commit message, nothing else.`
}

// getShortCommitPrompt returns the system prompt for short commit messages
func getShortCommitPrompt() string {
	return `You are a git commit message generator. Analyze the provided git diff and create a single-line commit message.

Rules:
1. Use Conventional Commits format: type(scope): description
2. Common types: feat, fix, refactor, chore, docs, style, test, perf, ci, build
3. Keep under 50 characters total
4. Use imperative mood (e.g., "add feature" not "added feature")
5. Be concise but descriptive
6. NO body text, NO explanations, just the subject line

Examples:
feat(auth): add JWT authentication
fix(db): resolve connection timeout
refactor(api): simplify error handling
docs(readme): update installation steps
test(user): add login validation tests

Output ONLY the commit subject line, nothing else.`
}

// getDefaultCommitExamples provides example commit messages when no git history exists
func getDefaultCommitExamples() string {
	return `Example commit messages for reference:

feat(auth): add JWT-based user authentication

- Implement JWT token generation and validation
- Add middleware for protecting authenticated routes
- Create secure login/logout endpoints

This enables secure user sessions and improves API security
by replacing cookie-based auth with industry-standard JWT tokens.

fix(db): resolve connection timeout issues

- Increase connection pool size from 10 to 50
- Add retry logic for failed connections
- Implement connection health checks

Fixes frequent timeout errors during peak usage periods
that were causing 500 errors for users.

refactor(api): simplify error handling across endpoints

- Create centralized error handler middleware
- Standardize error response format
- Remove duplicate error handling code

Improves code maintainability and provides consistent
error messages to frontend clients.`
}
