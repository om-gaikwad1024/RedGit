package core

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

func ValidateBranchName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("branch name cannot be empty")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("invalid branch name: contains '..'")
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	if !matched {
		return fmt.Errorf("branch name contains invalid characters")
	}
	return nil
}

func SafePath(base, path string) (string, error) {
	fullPath, err := filepath.Abs(filepath.Join(base, path))
	if err != nil {
		return "", err
	}
	
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	
	if !strings.HasPrefix(fullPath, absBase) {
		return "", fmt.Errorf("path traversal attempt detected")
	}
	
	return fullPath, nil
}

func ValidateCommitMessage(message string) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("commit message cannot be empty")
	}
	return nil
}