package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nguyenanhhao221/go-google-ai/pkg/commitgen"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file, using system environment")
	}

	shortCommit := flag.Bool("short", false, "Just generate short commit title")
	flag.Parse()

	// Create commit generator
	commitGen, err := commitgen.New(&commitgen.Options{
		IsShortCommit: *shortCommit,
		// API key will be loaded from GOOGLE_API_KEY environment variable
		// WorkingDir defaults to current directory
	})
	if err != nil {
		log.Fatalf("Failed to initialize commit generator: %v", err)
	}
	defer commitGen.Close()

	// Check for staged changes first
	hasChanges, err := commitGen.HasStagedChanges()
	if err != nil {
		log.Fatalf("Failed to check for staged changes: %v", err)
	}

	if !hasChanges {
		fmt.Println("No staged changes found. Please stage your changes with 'git add' first.")
		os.Exit(1)
	}

	// Generate commit message
	message, err := commitGen.Generate()
	if err != nil {
		log.Fatalf("Failed to generate commit message: %v", err)
	}

	// Output the generated commit message
	fmt.Println(message)
}
