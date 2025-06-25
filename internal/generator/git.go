package generator

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitRepository represents a git repository and provides methods to extract information
type GitRepository struct {
	workingDir string
}

// NewGitRepository creates a new GitRepository instance
// If workingDir is empty, it uses the current directory
func NewGitRepository(workingDir string) *GitRepository {
	return &GitRepository{
		workingDir: workingDir,
	}
}

// GetStagedDiff returns the staged changes in the repository
func (g *GitRepository) GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "--no-pager", "diff", "--staged")
	if g.workingDir != "" {
		cmd.Dir = g.workingDir
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	return string(output), nil
}

// GetRecentCommits returns the last n commit messages from the repository
func (g *GitRepository) GetRecentCommits(count int) (string, error) {
	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", count), "--oneline")
	if g.workingDir != "" {
		cmd.Dir = g.workingDir
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get recent commits: %w", err)
	}

	// If no commits exist, return empty string to trigger fallback
	if strings.TrimSpace(string(output)) == "" {
		return "", fmt.Errorf("no git history found")
	}

	return string(output), nil
}

// GetDetailedCommitHistory returns detailed commit history for context
func (g *GitRepository) GetDetailedCommitHistory(count int) (string, error) {
	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", count))
	if g.workingDir != "" {
		cmd.Dir = g.workingDir
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get detailed commit history: %w", err)
	}

	if strings.TrimSpace(string(output)) == "" {
		return "", fmt.Errorf("no git history found")
	}

	return string(output), nil
}

// HasStagedChanges checks if there are any staged changes
func (g *GitRepository) HasStagedChanges() (bool, error) {
	diff, err := g.GetStagedDiff()
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(diff) != "", nil
}

// GitInfo contains all the git information needed for commit message generation
type GitInfo struct {
	StagedDiff    string
	RecentCommits string
	HasHistory    bool
}

// GetCommitContext gathers all necessary git information in one call
// This is the primary method that consuming applications should use
func (g *GitRepository) GetCommitContext() (*GitInfo, error) {
	// Check for staged changes first
	hasStagedChanges, err := g.HasStagedChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to check for staged changes: %w", err)
	}

	if !hasStagedChanges {
		return nil, fmt.Errorf("no staged changes found")
	}

	// Get staged diff
	diff, err := g.GetStagedDiff()
	if err != nil {
		return nil, err
	}

	// Get recent commits (try detailed first, fall back to simple)
	recentCommits, err := g.GetDetailedCommitHistory(10)
	hasHistory := true
	if err != nil {
		// Try simple format as fallback
		recentCommits, err = g.GetRecentCommits(10)
		if err != nil {
			hasHistory = false
			recentCommits = ""
		}
	}

	return &GitInfo{
		StagedDiff:    diff,
		RecentCommits: recentCommits,
		HasHistory:    hasHistory,
	}, nil
}
