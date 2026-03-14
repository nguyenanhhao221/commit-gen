package generator

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/genai"
)

// CommitMessageGenerator handles AI-powered commit message generation.
type CommitMessageGenerator struct {
	client        *genai.Client
	config        *GeneratorConfig
	systemPrompt  string
	isShortCommit bool
}

// GeneratorConfig contains configuration for the commit message generator.
type GeneratorConfig struct {
	Model   string
	Timeout time.Duration
	APIKey  string
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *GeneratorConfig {
	return &GeneratorConfig{
		Model:   "gemini-2.5-flash-lite", // Fast and Dirty just like we like it
		Timeout: 10 * time.Second,
	}
}

// NewCommitMessageGenerator creates a new commit message generator.
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

// GenerateCommitMessage generates a commit message from git information.
func (g *CommitMessageGenerator) GenerateCommitMessage(gitInfo *GitInfo) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.config.Timeout)
	defer cancel()

	prompt := buildPrompt(gitInfo)

	genConfig := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(g.systemPrompt, genai.RoleUser),
		ThinkingConfig: &genai.ThinkingConfig{
			IncludeThoughts: false,
			ThinkingBudget:  func() *int32 { v := int32(0); return &v }(), // Disable thinking
		},
	}

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

// Close cleans up resources.
func (g *CommitMessageGenerator) Close() error {
	return nil
}
