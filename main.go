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
			`You are a git commit message generator. Analyze the provided git diff and recent git log to create an appropriate commit message.
Rules:
1. Use Conventional Commits format: type(scope): description
2. Common types: feat, fix, refactor, chore, docs, style, test, perf, ci, build
3. Keep the subject line under 50 characters
4. Use imperative mood (e.g., "add feature" not "added feature")
5. Match the style and tone of recent commits in the git log
6. Focus on WHAT changed and WHY it matters
7. Be specific but concise

Examples:
- feat: add user authentication system
- fix: resolve memory leak in data processing
- refactor: simplify error handling logic
- chore: update dependencies to latest versions

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

	// Get recent git log (last 5 commits)
	gitLog, err := getGitLog()
	if err != nil {
		log.Fatalf("Failed to get git log: %v", err)
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

// getGitLog executes 'git log --oneline -5' and returns the output
func getGitLog() (string, error) {
	cmd := exec.Command("git", "log", "-10")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}
	return string(output), nil
}
