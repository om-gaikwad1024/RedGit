package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/index"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"os"
	"path/filepath"
	"strings"
)

func Add(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("nothing specified, nothing added")
	}
	
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	gitignore, err := core.LoadGitIgnore(repo.WorkingDir)
	if err != nil {
		return err
	}
	
	idx, err := index.ReadIndex(repo.GitDir)
	if err != nil {
		return err
	}
	
	for _, path := range args {
		absPath, _ := filepath.Abs(path)
		relPath, _ := filepath.Rel(repo.WorkingDir, absPath)
		
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("pathspec '%s' did not match any files", path)
		}
		
		if info.IsDir() {
			if err := addDirectory(repo, idx, relPath, gitignore); err != nil {
				return err
			}
		} else {
			if !gitignore.IsIgnored(relPath) {
				if err := addFile(repo, idx, relPath); err != nil {
					return err
				}
			}
		}
	}
	
	return idx.WriteIndex(repo.GitDir)
}

func addFile(repo *core.Repository, idx *index.Index, path string) error {
	fullPath := filepath.Join(repo.WorkingDir, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}
	
	blob := object.NewBlob(content)
	hash, err := object.WriteObject(repo.GitDir, blob)
	if err != nil {
		return err
	}
	
	info, err := os.Stat(fullPath)
	if err != nil {
		return err
	}
	
	entry := index.NewEntry(path, hash, info)
	idx.AddEntry(entry)
	
	return nil
}

func addDirectory(repo *core.Repository, idx *index.Index, dirPath string, gitignore *core.GitIgnore) error {
	fullPath := filepath.Join(repo.WorkingDir, dirPath)
	
	return filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		relPath, _ := filepath.Rel(repo.WorkingDir, path)
		
		if strings.HasPrefix(relPath, ".git"+string(filepath.Separator)) {
			return nil
		}
		
		if gitignore.IsIgnored(relPath) {
			return nil
		}
		
		return addFile(repo, idx, relPath)
	})
}