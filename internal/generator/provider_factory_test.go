package generator

import (
	"os"
	"strings"
	"testing"
)

func TestNormalizeProviderDefaultsToClaudeCLI(t *testing.T) {
	if got := normalizeProvider(""); got != ProviderClaudeCLI {
		t.Fatalf("expected %q, got %q", ProviderClaudeCLI, got)
	}
}

func TestNormalizeProviderAlias(t *testing.T) {
	if got := normalizeProvider("claude"); got != ProviderClaudeCLI {
		t.Fatalf("expected alias to resolve to %q, got %q", ProviderClaudeCLI, got)
	}
}

func TestCreateProviderClaudeDefaultModel(t *testing.T) {
	originalLookPath := lookPathFn
	defer func() { lookPathFn = originalLookPath }()

	lookPathFn = func(file string) (string, error) {
		if file != "claude" {
			t.Fatalf("unexpected executable lookup: %q", file)
		}

		return "/usr/bin/claude", nil
	}

	provider, model, timeout, err := createProvider(&Options{Provider: ProviderClaudeCLI})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if provider == nil {
		t.Fatal("expected provider instance")
	}

	if model != "haiku" {
		t.Fatalf("expected default model %q, got %q", "haiku", model)
	}

	if timeout != defaultProviderTimeout {
		t.Fatalf("expected timeout %v, got %v", defaultProviderTimeout, timeout)
	}
}

func TestCreateProviderUnsupported(t *testing.T) {
	_, _, _, err := createProvider(&Options{Provider: "unknown"})
	if err == nil {
		t.Fatal("expected unsupported provider error")
	}

	if !strings.Contains(err.Error(), "unsupported provider") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateProviderGeminiMissingAPIKey(t *testing.T) {
	original, existed := os.LookupEnv("GOOGLE_API_KEY")
	if err := os.Unsetenv("GOOGLE_API_KEY"); err != nil {
		t.Fatalf("failed to unset GOOGLE_API_KEY: %v", err)
	}

	t.Cleanup(func() {
		if !existed {
			_ = os.Unsetenv("GOOGLE_API_KEY")
			return
		}

		_ = os.Setenv("GOOGLE_API_KEY", original)
	})

	_, _, _, err := createProvider(&Options{Provider: ProviderGemini})
	if err == nil {
		t.Fatal("expected missing API key error")
	}

	if !strings.Contains(err.Error(), "requires API key") {
		t.Fatalf("unexpected error: %v", err)
	}
}
