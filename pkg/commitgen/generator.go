package commitgen

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/genai"
)

// CommitMessageGenerator handles AI-powered commit message generation
type CommitMessageGenerator struct {
	client       *genai.Client
	config       *GeneratorConfig
	systemPrompt string
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
func NewCommitMessageGenerator(config *GeneratorConfig) (*CommitMessageGenerator, error) {
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

	return &CommitMessageGenerator{
		client:       client,
		config:       config,
		systemPrompt: getDefaultSystemPrompt(),
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

