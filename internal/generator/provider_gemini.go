package generator

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/genai"
)

// GeminiProvider generates content with the Gemini API.
type GeminiProvider struct {
	client *genai.Client
	config *GeminiConfig
}

// GeminiConfig contains configuration for the Gemini provider.
type GeminiConfig struct {
	Model   string
	Timeout time.Duration
	APIKey  string
}

// DefaultGeminiConfig returns a default Gemini configuration.
func DefaultGeminiConfig() *GeminiConfig {
	return &GeminiConfig{
		Model:   "gemini-2.5-flash-lite", // Fast and Dirty just like we like it
		Timeout: 10 * time.Second,
	}
}

// NewGeminiProvider creates a new Gemini provider.
func NewGeminiProvider(config *GeminiConfig) (*GeminiProvider, error) {
	if config == nil {
		config = DefaultGeminiConfig()
	}

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

	return &GeminiProvider{
		client: client,
		config: config,
	}, nil
}

// Generate requests content generation from Gemini.
func (g *GeminiProvider) Generate(ctx context.Context, req *GenerateRequest) (string, error) {
	if req == nil {
		return "", fmt.Errorf("generation request is required")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	model := req.Model
	if model == "" {
		model = g.config.Model
	}

	genConfig := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(req.SystemPrompt, genai.RoleUser),
		ThinkingConfig: &genai.ThinkingConfig{
			IncludeThoughts: false,
			ThinkingBudget:  func() *int32 { v := int32(0); return &v }(), // Disable thinking
		},
	}

	result, err := g.client.Models.GenerateContent(
		ctx,
		model,
		genai.Text(req.Prompt),
		genConfig,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate commit message: %w", err)
	}

	return result.Text(), nil
}

// Close cleans up resources.
func (g *GeminiProvider) Close() error {
	return nil
}
