package core

import (
	"fmt"
	"os"
	"path/filepath"
)

type Repository struct {
	GitDir     string
	WorkingDir string
}

func NewRepository(path string) (*Repository, error) {
	workingDir, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	
	gitDir := filepath.Join(workingDir, ".git")
	
	return &Repository{
		GitDir:     gitDir,
		WorkingDir: workingDir,
	}, nil
}

func (r *Repository) Init() error {
	dirs := []string{
		r.GitDir,
		filepath.Join(r.GitDir, "objects"),
		filepath.Join(r.GitDir, "refs"),
		filepath.Join(r.GitDir, "refs", "heads"),
		filepath.Join(r.GitDir, "refs", "tags"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	
	headPath := filepath.Join(r.GitDir, "HEAD")
	if err := os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0644); err != nil {
		return err
	}
	
	configPath := filepath.Join(r.GitDir, "config")
	configContent := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return err
	}
	
	return nil
}

func FindRepository(path string) (*Repository, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	
	current := absPath
	for {
		gitDir := filepath.Join(current, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return &Repository{
				GitDir:     gitDir,
				WorkingDir: current,
			}, nil
		}
		
		parent := filepath.Dir(current)
		if parent == current {
			return nil, fmt.Errorf("not a git repository")
		}
		current = parent
	}
}