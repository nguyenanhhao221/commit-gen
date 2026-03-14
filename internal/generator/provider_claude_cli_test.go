package generator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

func TestClaudeCLIProviderGenerate(t *testing.T) {
	originalCommandContext := commandContextFn
	defer func() { commandContextFn = originalCommandContext }()

	commandContextFn = fakeExecCommandContext

	setHelperEnv(t, "EXPECTED_MODEL", "haiku")
	setHelperEnv(t, "EXPECTED_SYSTEM_PROMPT", "system prompt")
	setHelperEnv(t, "EXPECTED_PROMPT", "prompt body")
	setHelperEnv(t, "HELPER_OUTPUT", "  feat(cli): add provider flag\n")
	setHelperEnv(t, "HELPER_EXIT_CODE", "0")

	provider := &ClaudeCLIProvider{commandPath: "/usr/bin/claude"}

	message, err := provider.Generate(context.Background(), &GenerateRequest{
		Model:        "haiku",
		SystemPrompt: "system prompt",
		Prompt:       "prompt body",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if message != "feat(cli): add provider flag" {
		t.Fatalf("unexpected message: %q", message)
	}
}

func TestClaudeCLIProviderGenerateCommandError(t *testing.T) {
	originalCommandContext := commandContextFn
	defer func() { commandContextFn = originalCommandContext }()

	commandContextFn = fakeExecCommandContext

	setHelperEnv(t, "EXPECTED_MODEL", "haiku")
	setHelperEnv(t, "EXPECTED_SYSTEM_PROMPT", "system prompt")
	setHelperEnv(t, "EXPECTED_PROMPT", "prompt body")
	setHelperEnv(t, "HELPER_OUTPUT", "")
	setHelperEnv(t, "HELPER_STDERR", "simulated failure")
	setHelperEnv(t, "HELPER_EXIT_CODE", "3")

	provider := &ClaudeCLIProvider{commandPath: "/usr/bin/claude"}

	_, err := provider.Generate(context.Background(), &GenerateRequest{
		Model:        "haiku",
		SystemPrompt: "system prompt",
		Prompt:       "prompt body",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "simulated failure") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClaudeCLIProviderGenerateNilRequest(t *testing.T) {
	provider := &ClaudeCLIProvider{commandPath: "/usr/bin/claude"}

	_, err := provider.Generate(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClaudeCLIProviderHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	separator := -1
	for i, arg := range args {
		if arg == "--" {
			separator = i
			break
		}
	}

	if separator == -1 || separator+2 >= len(args) {
		fmt.Fprint(os.Stderr, "invalid helper invocation")
		os.Exit(2)
	}

	commandPath := args[separator+1]
	commandArgs := args[separator+2:]

	if commandPath != "/usr/bin/claude" {
		fmt.Fprintf(os.Stderr, "unexpected command path: %s", commandPath)
		os.Exit(2)
	}

	expected := []string{
		"-p",
		"--output-format",
		"text",
		"--model",
		os.Getenv("EXPECTED_MODEL"),
		"--system-prompt",
		os.Getenv("EXPECTED_SYSTEM_PROMPT"),
		os.Getenv("EXPECTED_PROMPT"),
	}

	if len(expected) != len(commandArgs) {
		fmt.Fprintf(os.Stderr, "unexpected args length: %d", len(commandArgs))
		os.Exit(2)
	}

	for i := range expected {
		if expected[i] != commandArgs[i] {
			fmt.Fprintf(os.Stderr, "arg mismatch at %d: expected %q, got %q", i, expected[i], commandArgs[i])
			os.Exit(2)
		}
	}

	if stderr := os.Getenv("HELPER_STDERR"); stderr != "" {
		fmt.Fprint(os.Stderr, stderr)
	}

	fmt.Print(os.Getenv("HELPER_OUTPUT"))

	exitCode, err := strconv.Atoi(os.Getenv("HELPER_EXIT_CODE"))
	if err != nil {
		exitCode = 0
	}

	os.Exit(exitCode)
}

func fakeExecCommandContext(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestClaudeCLIProviderHelperProcess", "--", command}
	cs = append(cs, args...)

	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")

	return cmd
}

func setHelperEnv(t *testing.T, key, value string) {
	t.Helper()

	original, existed := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed setting env %s: %v", key, err)
	}

	t.Cleanup(func() {
		if !existed {
			_ = os.Unsetenv(key)
			return
		}

		_ = os.Setenv(key, original)
	})
}
