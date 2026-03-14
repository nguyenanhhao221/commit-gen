package generator

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const defaultProviderTimeout = 10 * time.Second

func normalizeProvider(provider string) string {
	value := strings.TrimSpace(strings.ToLower(provider))
	if value == "" {
		return ProviderGemini
	}

	return value
}

func createProvider(opts *Options) (AIProvider, string, time.Duration, error) {
	providerName := normalizeProvider(opts.Provider)

	switch providerName {
	case ProviderGemini:
		config := DefaultGeminiConfig()
		if opts.Model != "" {
			config.Model = opts.Model
		}

		apiKey := strings.TrimSpace(opts.APIKey)
		if apiKey == "" {
			apiKey = os.Getenv("GOOGLE_API_KEY")
		}
		if apiKey == "" {
			return nil, "", 0, fmt.Errorf("gemini provider requires API key in options or GOOGLE_API_KEY environment variable")
		}

		config.APIKey = apiKey
		provider, err := NewGeminiProvider(config)
		if err != nil {
			return nil, "", 0, fmt.Errorf("failed to initialize gemini provider: %w", err)
		}

		return provider, config.Model, config.Timeout, nil
	case ProviderClaudeCLI:
		provider, err := NewClaudeCLIProvider()
		if err != nil {
			return nil, "", 0, err
		}

		model := strings.TrimSpace(opts.Model)
		if model == "" {
			model = "haiku"
		}

		return provider, model, defaultProviderTimeout, nil
	default:
		return nil, "", 0, fmt.Errorf("unsupported provider %q (supported: %s, %s)", providerName, ProviderGemini, ProviderClaudeCLI)
	}
}
