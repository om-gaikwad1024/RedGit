package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/index"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"github.com/om-gaikwad1024/redgit/internal/refs"
	"os"
	"path/filepath"
	"strings"
)

func Status(args []string) error {
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	gitignore, err := core.LoadGitIgnore(repo.WorkingDir)
	if err != nil {
		return err
	}
	
	branch, err := refs.GetCurrentBranch(repo.GitDir)
	if err != nil {
		fmt.Println("HEAD detached")
	} else {
		fmt.Printf("On branch %s\n", branch)
	}
	
	idx, err := index.ReadIndex(repo.GitDir)
	if err != nil {
		return err
	}
	
	stagedFiles := make(map[string]string)
	for _, entry := range idx.Entries {
		stagedFiles[entry.Path] = entry.Hash
	}
	
	workingFiles := make(map[string]bool)
	modifiedFiles := []string{}
	untrackedFiles := []string{}
	
	filepath.Walk(repo.WorkingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		relPath, _ := filepath.Rel(repo.WorkingDir, path)
		
		if relPath == "." {
			return nil
		}
		
		if relPath == ".git" || strings.HasPrefix(relPath, ".git"+string(filepath.Separator)) {
			return filepath.SkipDir
		}
		
		if info.IsDir() {
			return nil
		}
		
		if gitignore.IsIgnored(relPath) {
			return nil
		}
		
		workingFiles[relPath] = true
		
		if stagedHash, exists := stagedFiles[relPath]; exists {
			content, _ := os.ReadFile(path)
			blob := object.NewBlob(content)
			currentHash := blob.Hash()
			
			if currentHash != stagedHash {
				modifiedFiles = append(modifiedFiles, relPath)
			}
		} else {
			untrackedFiles = append(untrackedFiles, relPath)
		}
		
		return nil
	})
	
	deletedFiles := []string{}
	for stagedPath := range stagedFiles {
		if !workingFiles[stagedPath] {
			deletedFiles = append(deletedFiles, stagedPath)
		}
	}
	
	if len(idx.Entries) > 0 {
		fmt.Println("\nChanges to be committed:")
		for _, entry := range idx.Entries {
			fmt.Printf("  new file:   %s\n", entry.Path)
		}
	}
	
	if len(modifiedFiles) > 0 {
		fmt.Println("\nChanges not staged for commit:")
		for _, file := range modifiedFiles {
			fmt.Printf("  modified:   %s\n", file)
		}
	}
	
	if len(deletedFiles) > 0 {
		if len(modifiedFiles) == 0 {
			fmt.Println("\nChanges not staged for commit:")
		}
		for _, file := range deletedFiles {
			fmt.Printf("  deleted:    %s\n", file)
		}
	}
	
	if len(untrackedFiles) > 0 {
		fmt.Println("\nUntracked files:")
		for _, file := range untrackedFiles {
			fmt.Printf("  %s\n", file)
		}
	}
	
	if len(idx.Entries) == 0 && len(modifiedFiles) == 0 && len(untrackedFiles) == 0 && len(deletedFiles) == 0 {
		fmt.Println("nothing to commit, working tree clean")
	}
	
	return nil
}