package xenv

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadOptions represents options for loading environment variables
type LoadOptions struct {
	Filename string
}

// Load reads the .env file and loads the environment variables
func Load(options ...LoadOptions) error {
	filename := ".env"
	if len(options) > 0 && options[0].Filename != "" {
		filename = options[0].Filename
	}

	// Try to find the .env file in the current directory and parent directories
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	for {
		filePath := filepath.Join(currentDir, filename)
		if _, err := os.Stat(filePath); err == nil {
			return loadEnvFile(filePath)
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break // We've reached the root directory
		}
		currentDir = parentDir
	}

	return fmt.Errorf(".env file not found")
}

func loadEnvFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening .env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid line in .env file: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		value = strings.Trim(value, `"'`)

		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("error setting environment variable %s: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}

	return nil
}
