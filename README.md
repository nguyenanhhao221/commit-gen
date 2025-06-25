# CommitGen - AI-Powered Git Commit Message Generator

A **core library** and CLI tool for generating intelligent Git commit messages using Google's Gemini AI. Designed to be integrated into various tools like Lazygit, Neovim, VS Code, or used standalone.

## Features

- 🤖 **AI-Powered Analysis**: Uses Google Gemini to understand your code changes
- 📦 **Core Library**: Clean Go package that can be integrated into any application
- 🔌 **Multiple Integrations**: CLI tool, IDE plugins, git hooks, or custom applications
- 📝 **Conventional Commits**: Generates properly formatted commit messages following industry standards
- 🔍 **Context Aware**: Analyzes both your diff and recent git history for consistent style
- ⚡ **Fast & Lightweight**: Minimal dependencies and clean architecture
- 🎯 **Detailed Messages**: Creates both concise subject lines and informative commit bodies
- 🛠️ **Configurable**: Custom prompts, models, and working directories

## How It Works

1. Analyzes your `git diff --staged` (your staged changes)
2. Reviews your recent git log to match your project's commit style
3. Uses Google's Gemini AI to generate a complete commit message with:
   - Proper subject line with conventional commit format
   - Detailed body explaining what, how, and why
   - Consistent tone matching your project's history

## Installation

### Prerequisites

- Go 1.24 or later
- Git repository
- Google AI API key ([Get one here](https://ai.google.dev/))

### Setup

1. Clone this repository:

```bash
git clone <repository-url>
cd go-google-ai
```

2. Install dependencies:

```bash
go mod tidy
```

3. Set up your Google AI API key:

```bash
# Option 1: Environment variable
export GOOGLE_API_KEY="your-api-key-here"

# Option 2: Create .env file
echo "GOOGLE_API_KEY=your-api-key-here" > .env
```

4. Build the binary:

```bash
go build -o commit-gen main.go
```

## Usage

### Basic Usage

1. Stage your changes:

```bash
git add .
```

2. Generate commit message:

```bash
# Full commit message (subject + body)
./commit-gen

# Short commit message (subject only)
./commit-gen -short
```

3. Use the output for your commit:

```bash
# Full commit
git commit -m "$(./commit-gen)"

# Quick short commit
git commit -m "$(./commit-gen -short)"
```

### Example Output

**Full commit message** (`./commit-gen`):

```
feat(auth): add JWT-based user authentication

- Implement JWT token generation and validation
- Add middleware for protecting authenticated routes  
- Create user login/logout endpoints with secure session handling

This change enables secure user sessions and replaces the previous
cookie-based authentication which had security vulnerabilities.
The new system provides better scalability and follows industry
best practices for API authentication.
```

**Short commit message** (`./commit-gen -short`):

```
feat(auth): add JWT-based user authentication
```

### Using as a Library

```go
package main

import (
    "fmt"
    "log"
    "github.com/nguyenanhhao221/go-google-ai/pkg/commitgen"
)

func main() {
    // Simple usage
    message, err := commitgen.QuickGenerate("your-api-key")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(message)
    
    // Quick short commit
    shortMessage, err := commitgen.QuickGenerateShort("your-api-key")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(shortMessage)
    
    // Advanced usage
    commitGen, err := commitgen.New(&commitgen.Options{
        APIKey:        "your-api-key",
        WorkingDir:    "/path/to/repo",
        Model:         "gemini-1.5-pro",
        IsShortCommit: true, // Generate short commits
    })
    if err != nil {
        log.Fatal(err)
    }
    defer commitGen.Close()
    
    message, err = commitGen.Generate()
    // ... use message
}
```

### Integration Examples

**Lazygit Custom Command**:

```yaml
# ~/.config/lazygit/config.yml
customCommands:
  - key: 'C'
    command: 'your-binary-path'
    description: 'Generate AI commit message'
    context: 'files'
```

**Neovim Lua Plugin**:

```lua
-- Add to your Neovim config
local function generate_commit_message()
  local handle = io.popen('your-binary-path')
  local result = handle:read("*a")
  handle:close()
  vim.api.nvim_put({result}, "l", true, true)
end

vim.keymap.set('n', '<leader>gc', generate_commit_message)
```

**Git Alias** (CLI):

```bash
git config --global alias.smart-commit '!git commit -m "$(commit-gen)"'
# Usage: git smart-commit
```

**Git Hook**: Add to `.git/hooks/prepare-commit-msg` for automatic suggestions

## Configuration

The AI prompt is currently embedded in the code but follows these rules:

- **Subject line**: `type(scope): description` (max 50 chars)
- **Types**: feat, fix, refactor, chore, docs, style, test, perf, ci, build
- **Body**: Explains what, how, and why (wrapped at 72 chars)

## Error Handling

- **No staged changes**: The tool will prompt you to stage changes first
- **No API key**: Clear error message with setup instructions  
- **No git history**: Falls back to example commit formats
- **API timeout**: 10-second timeout prevents hanging

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

[Add your preferred license]

## Roadmap

- [ ] Configuration file support
- [ ] Custom prompt templates
- [ ] Multiple AI provider support
- [ ] Git hook automation
- [ ] Team-specific commit conventions
