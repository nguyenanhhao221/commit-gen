package generator

import (
	"strings"
	"testing"
)

func TestBuildPromptWithHistory(t *testing.T) {
	gitInfo := &GitInfo{
		StagedDiff:    "diff --git a/main.go b/main.go",
		RecentCommits: "feat: previous commit",
		HasHistory:    true,
	}

	prompt := buildPrompt(gitInfo)

	if !strings.Contains(prompt, gitInfo.RecentCommits) {
		t.Fatalf("expected history in prompt: %q", prompt)
	}

	if !strings.Contains(prompt, gitInfo.StagedDiff) {
		t.Fatalf("expected staged diff in prompt: %q", prompt)
	}
}

func TestBuildPromptWithoutHistoryUsesExamples(t *testing.T) {
	gitInfo := &GitInfo{
		StagedDiff: "diff --git a/main.go b/main.go",
		HasHistory: false,
	}

	prompt := buildPrompt(gitInfo)

	if !strings.Contains(prompt, "Example commit messages for reference") {
		t.Fatalf("expected default examples in prompt: %q", prompt)
	}

	if !strings.Contains(prompt, gitInfo.StagedDiff) {
		t.Fatalf("expected staged diff in prompt: %q", prompt)
	}
}

func TestPromptTemplatesContainConventionalCommitRule(t *testing.T) {
	full := getDefaultSystemPrompt()
	short := getShortCommitPrompt()

	expected := "Use Conventional Commits format: type(scope): description"
	if !strings.Contains(full, expected) {
		t.Fatalf("full prompt missing required rule")
	}

	if !strings.Contains(short, expected) {
		t.Fatalf("short prompt missing required rule")
	}
}
