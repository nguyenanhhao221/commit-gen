package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, trying to load environment")
	}
	googleAPIKey := os.Getenv("GOOGLE_API_KEY")
	if googleAPIKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(
			`You are a git commit message generator. Analyze the provided git diff and recent git log to create a complete commit message with both subject and body.

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
Output only the commit message, nothing else.`,
			genai.RoleUser,
		),
		ThinkingConfig: &genai.ThinkingConfig{
			IncludeThoughts: false,
			ThinkingBudget:  func() *int32 { v := int32(0); return &v }(), // Disables thinking
		},
	}

	client, _ := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  googleAPIKey,
		Backend: genai.BackendGeminiAPI,
	})

	// Get git diff (staged changes)
	gitDiff, err := getGitDiff()
	if err != nil {
		log.Fatalf("Failed to get git diff: %v", err)
	}

	if strings.TrimSpace(gitDiff) == "" {
		fmt.Println("No staged changes found. Please stage your changes with 'git add' first.")
		return
	}

	// Get recent git log (last 10 commits)
	gitLog, err := getGitLog()
	if err != nil {
		log.Printf("Warning: Failed to get git log: %v", err)
		gitLog = getDefaultCommitExamples()
	}

	// Compose the prompt
	prompt := fmt.Sprintf(
		"Recent git log:\n%s\n\nGit diff:\n%s\n",
		gitLog,
		gitDiff,
	)

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash-lite-preview-06-17",
		genai.Text(prompt),
		config,
	)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(result.Text())
}

// getGitDiff executes 'git diff --staged' and returns the output
func getGitDiff() (string, error) {
	cmd := exec.Command("git", "--no-pager", "diff", "--staged")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	return string(output), nil
}

// getGitLog executes 'git log -10' and returns the output
func getGitLog() (string, error) {
	cmd := exec.Command("git", "log", "-10")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}
	
	// If no commits exist, return empty string to trigger fallback
	if strings.TrimSpace(string(output)) == "" {
		return "", fmt.Errorf("no git history found")
	}
	
	return string(output), nil
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
