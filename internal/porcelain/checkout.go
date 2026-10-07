package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/index"
	"github.com/om-gaikwad1024/redgit/internal/merge"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"github.com/om-gaikwad1024/redgit/internal/refs"
	"os"
	"path/filepath"
)

func Checkout(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: redgit checkout [-b] <branch>")
	}
	
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	createBranch := false
	branchName := args[0]
	
	if args[0] == "-b" {
		if len(args) < 2 {
			return fmt.Errorf("usage: redgit checkout -b <branch>")
		}
		createBranch = true
		branchName = args[1]
	}
	
	if createBranch {
		currentCommit, err := refs.ResolveRef(repo.GitDir, "HEAD")
		if err != nil {
			return fmt.Errorf("cannot create branch: no commits yet")
		}
		
		if err := refs.CreateBranch(repo.GitDir, branchName, currentCommit); err != nil {
			return err
		}
	}
	
	targetCommit, err := refs.ResolveRef(repo.GitDir, "refs/heads/"+branchName)
	if err != nil {
		return fmt.Errorf("branch '%s' not found", branchName)
	}
	
	if err := merge.ClearWorkingDirectory(repo.WorkingDir, repo.GitDir); err != nil {
		return err
	}
	
	if err := merge.ApplyMerge(repo.GitDir, repo.WorkingDir, targetCommit); err != nil {
		return err
	}
	
	targetObj, _ := object.ReadObject(repo.GitDir, targetCommit)
	targetCommitParsed, _ := object.ParseCommit(targetObj.Content)
	
	idx := index.NewIndex()
	if err := rebuildIndexFromTree(repo.GitDir, repo.WorkingDir, targetCommitParsed.Tree, "", idx); err != nil {
		return err
	}
	
	if err := idx.WriteIndex(repo.GitDir); err != nil {
		return err
	}
	
	if err := refs.SetHEAD(repo.GitDir, branchName); err != nil {
		return err
	}
	
	fmt.Printf("Switched to branch '%s'\n", branchName)
	return nil
}

func rebuildIndexFromTree(gitDir, workingDir, treeHash, prefix string, idx *index.Index) error {
	treeObj, err := object.ReadObject(gitDir, treeHash)
	if err != nil {
		return err
	}
	
	entries, err := object.ParseTree(treeObj.Content)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		fullPath := filepath.Join(prefix, entry.Name)
		
		if entry.IsDirectory() {
			if err := rebuildIndexFromTree(gitDir, workingDir, entry.Hash, fullPath, idx); err != nil {
				return err
			}
		} else {
			absPath := filepath.Join(workingDir, fullPath)
			info, err := os.Stat(absPath)
			if err != nil {
				continue
			}
			
			idxEntry := index.NewEntry(fullPath, entry.Hash, info)
			idx.AddEntry(idxEntry)
		}
	}
	
	return nil
}