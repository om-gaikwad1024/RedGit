package core

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type GitIgnore struct {
	patterns []string
}

func LoadGitIgnore(repoPath string) (*GitIgnore, error) {
	gi := &GitIgnore{
		patterns: []string{},
	}
	
	gitignorePath := filepath.Join(repoPath, ".gitignore")
	file, err := os.Open(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return gi, nil
		}
		return nil, err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		gi.patterns = append(gi.patterns, line)
	}
	
	return gi, scanner.Err()
}

func (gi *GitIgnore) IsIgnored(path string) bool {
	cleanPath := filepath.ToSlash(path)
	
	for _, pattern := range gi.patterns {
		if matchPattern(pattern, cleanPath) {
			return true
		}
	}
	
	return false
}

func matchPattern(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	
	if strings.HasPrefix(pattern, "!") {
		return false
	}
	
	if strings.HasSuffix(pattern, "/") {
		pattern = strings.TrimSuffix(pattern, "/")
		return strings.HasPrefix(path, pattern+"/") || path == pattern
	}
	
	if strings.Contains(pattern, "/") {
		matched, _ := filepath.Match(pattern, path)
		return matched
	}
	
	parts := strings.Split(path, "/")
	for _, part := range parts {
		matched, _ := filepath.Match(pattern, part)
		if matched {
			return true
		}
	}
	
	matched, _ := filepath.Match(pattern, filepath.Base(path))
	return matched
}