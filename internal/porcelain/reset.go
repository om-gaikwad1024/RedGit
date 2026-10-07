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

func Reset(args []string) error {
	mode := "mixed"
	var target string
	
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--soft":
			mode = "soft"
		case "--mixed":
			mode = "mixed"
		case "--hard":
			mode = "hard"
		default:
			target = args[i]
		}
	}
	
	if target == "" {
		target = "HEAD"
	}
	
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	
	targetCommit, err := refs.ResolveRevision(repo.GitDir, target)
	if err != nil {
		return fmt.Errorf("failed to resolve '%s': %v", target, err)
	}
	
	if err := refs.UpdateRef(repo.GitDir, "HEAD", targetCommit); err != nil {
		return err
	}
	
	if mode == "soft" {
		return nil
	}
	
	targetObj, err := object.ReadObject(repo.GitDir, targetCommit)
	if err != nil {
		return err
	}
	
	targetCommitParsed, err := object.ParseCommit(targetObj.Content)
	if err != nil {
		return err
	}
	
	idx := index.NewIndex()
	if err := rebuildIndexFromTreeReset(repo.GitDir, repo.WorkingDir, targetCommitParsed.Tree, "", idx); err != nil {
		return err
	}
	
	if err := idx.WriteIndex(repo.GitDir); err != nil {
		return err
	}
	
	if mode == "hard" {
		if err := merge.ClearWorkingDirectory(repo.WorkingDir, repo.GitDir); err != nil {
			return err
		}
		
		if err := merge.ApplyMerge(repo.GitDir, repo.WorkingDir, targetCommit); err != nil {
			return err
		}
	}
	
	fmt.Printf("HEAD is now at %s\n", targetCommit[:7])
	return nil
}

func rebuildIndexFromTreeReset(gitDir, workingDir, treeHash, prefix string, idx *index.Index) error {
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
			if err := rebuildIndexFromTreeReset(gitDir, workingDir, entry.Hash, fullPath, idx); err != nil {
				return err
			}
		} else {
			absPath := filepath.Join(workingDir, fullPath)
			info, err := os.Stat(absPath)
			if err != nil {
				info = nil
			}
			
			if info != nil {
				idxEntry := index.NewEntry(fullPath, entry.Hash, info)
				idx.AddEntry(idxEntry)
			}
		}
	}
	
	return nil
}	