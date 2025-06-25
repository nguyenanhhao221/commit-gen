package main

import (
	"fmt"
	"log"

	"github.com/nguyenanhhao221/go-google-ai/pkg/commitgen"
)

// Example showing how Lazygit, Nvim, or other tools could integrate with commitgen
func main() {
	fmt.Println("=== Example: Simple Usage ===")
	simpleUsage()

	fmt.Println("\n=== Example: Custom Configuration ===")
	customConfig()

	fmt.Println("\n=== Example: Using Custom Diff Data ===")
	customDiffData()
}

// simpleUsage demonstrates the most basic usage
func simpleUsage() {
	// For simple cases, just use QuickGenerate
	message, err := commitgen.QuickGenerate("your-api-key-here")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	fmt.Printf("Generated message:\n%s\n", message)
}

// customConfig shows how to use custom configuration
func customConfig() {
	commitGen, err := commitgen.New(&commitgen.Options{
		APIKey:     "your-api-key-here",
		WorkingDir: "/path/to/git/repo", // Different repository
		Model:      "gemini-1.5-pro",    // Different model
		CustomPrompt: `Create a concise commit message.
		
		Rules:
		- Keep it under 50 characters
		- Use imperative mood
		- Focus on the main change
		
		Format: type: description`,
	})
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer commitGen.Close()

	message, err := commitGen.Generate()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Custom configured message:\n%s\n", message)
}

// customDiffData shows how Lazygit or IDE plugins could pass their own git data
func customDiffData() {
	commitGen, err := commitgen.New(&commitgen.Options{
		APIKey: "your-api-key-here",
	})
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer commitGen.Close()

	// This is how Lazygit could pass already-retrieved diff data
	diff := `diff --git a/src/auth.go b/src/auth.go
index 1234567..abcdefg 100644
--- a/src/auth.go
+++ b/src/auth.go
@@ -10,6 +10,10 @@ func Login(username, password string) error {
 	if username == "" || password == "" {
 		return errors.New("username and password required")
 	}
+	
+	// Add rate limiting
+	if !rateLimiter.Allow() {
+		return errors.New("too many login attempts")
+	}
 
 	return authenticateUser(username, password)
 }`

	history := `feat(auth): implement JWT authentication
fix(db): resolve connection pool issues
refactor(api): standardize error responses`

	message, err := commitGen.GenerateFromDiff(diff, history)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Message from custom diff:\n%s\n", message)
}

// For Nvim Lua integration, the usage would be:
/*
-- In your Neovim Lua config
local function generate_commit_message()
  local handle = io.popen('go run examples/nvim_integration.go')
  local result = handle:read("*a")
  handle:close()
  
  -- Insert the result into the commit message buffer
  vim.api.nvim_put({result}, "l", true, true)
end

vim.keymap.set('n', '<leader>gc', generate_commit_message, {desc = 'Generate commit message'})
*/

// For Lazygit integration, you could add this to your config:
/*
customCommands:
  - key: 'C'
    command: 'go run /path/to/examples/simple_usage.go'
    description: 'Generate AI commit message'
    context: 'files'
    loadingText: 'Generating commit message...'
*/ 