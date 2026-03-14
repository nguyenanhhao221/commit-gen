package generator

import "context"

const (
	ProviderClaudeCLI = "claude-cli"
	ProviderGemini    = "gemini"
)

// GenerateRequest contains provider-agnostic inputs for message generation.
type GenerateRequest struct {
	Model        string
	Prompt       string
	SystemPrompt string
}

// AIProvider abstracts message generation across AI backends.
type AIProvider interface {
	Generate(ctx context.Context, req *GenerateRequest) (string, error)
	Close() error
}
