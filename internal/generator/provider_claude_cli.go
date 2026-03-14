package generator

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

var (
	lookPathFn       = exec.LookPath
	commandContextFn = exec.CommandContext
)

// ClaudeCLIProvider generates content by invoking the local Claude CLI.
type ClaudeCLIProvider struct {
	commandPath string
}

// NewClaudeCLIProvider creates a provider backed by the local `claude` executable.
func NewClaudeCLIProvider() (*ClaudeCLIProvider, error) {
	path, err := lookPathFn("claude")
	if err != nil {
		return nil, fmt.Errorf("claude-cli provider requires `claude` in PATH: %w", err)
	}

	return &ClaudeCLIProvider{commandPath: path}, nil
}

// Generate requests content generation from the Claude CLI.
func (p *ClaudeCLIProvider) Generate(ctx context.Context, req *GenerateRequest) (string, error) {
	if req == nil {
		return "", fmt.Errorf("generation request is required")
	}

	args := []string{
		"-p",
		"--output-format", "text",
		"--model", req.Model,
		"--system-prompt", req.SystemPrompt,
		req.Prompt,
	}

	cmd := commandContextFn(ctx, p.commandPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("claude-cli generation failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	message := strings.TrimSpace(string(output))
	if message == "" {
		return "", fmt.Errorf("claude-cli returned empty output")
	}

	return message, nil
}

// Close satisfies the provider interface.
func (p *ClaudeCLIProvider) Close() error {
	return nil
}
