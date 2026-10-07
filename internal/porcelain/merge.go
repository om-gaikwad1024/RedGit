package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/index"
	"github.com/om-gaikwad1024/redgit/internal/merge"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"github.com/om-gaikwad1024/redgit/internal/refs"
	"path/filepath"
	"strings"
)

func Merge(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: redgit merge <branch>")
	}
	
	branchName := args[0]
	
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	ourCommit, err := refs.ResolveRef(repo.GitDir, "HEAD")
	if err != nil {
		return fmt.Errorf("no commits on current branch")
	}
	
	theirCommit, err := refs.ResolveRef(repo.GitDir, "refs/heads/"+branchName)
	if err != nil {
		return fmt.Errorf("branch '%s' not found", branchName)
	}
	
	if ourCommit == theirCommit {
		fmt.Println("Already up to date.")
		return nil
	}
	
	baseCommit, err := merge.FindMergeBase(repo.GitDir, ourCommit, theirCommit)
	if err != nil || baseCommit == "" {
		return fmt.Errorf("no common ancestor found")
	}
	
	if baseCommit == ourCommit {
		if err := merge.ApplyMerge(repo.GitDir, repo.WorkingDir, theirCommit); err != nil {
			return err
		}
		
		if err := refs.UpdateRef(repo.GitDir, "HEAD", theirCommit); err != nil {
			return err
		}
		
		fmt.Println("Fast-forward")
		return nil
	}
	
	if merge.IsFastForward(repo.GitDir, theirCommit, ourCommit) {
		fmt.Println("Already up to date.")
		return nil
	}
	
	conflicts, err := merge.ThreeWayMerge(repo.GitDir, baseCommit, ourCommit, theirCommit)
	if err != nil {
		return err
	}
	
	if len(conflicts) > 0 {
		merge.MarkConflicts(conflicts)
		fmt.Println("Automatic merge failed; fix conflicts and then commit the result.")
		for _, c := range conflicts {
			fmt.Printf("CONFLICT: %s\n", c.Path)
		}
		return fmt.Errorf("merge conflicts")
	}
	
	idx, err := index.ReadIndex(repo.GitDir)
	if err != nil {
		return err
	}
	
	treeHash, err := buildTreeFromIndex(repo.GitDir, idx.Entries)
	if err != nil {
		return err
	}
	
	author := getAuthor()
	mergeMessage := fmt.Sprintf("Merge branch '%s'", branchName)
	
	commitObj := object.NewCommit(treeHash, []string{ourCommit, theirCommit}, author, author, mergeMessage)
	commitHash, err := object.WriteObject(repo.GitDir, commitObj)
	if err != nil {
		return err
	}
	
	if err := refs.UpdateRef(repo.GitDir, "HEAD", commitHash); err != nil {
		return err
	}
	
	fmt.Printf("Merge made by the 'recursive' strategy.\n")
	return nil
}

func buildTreeFromIndex(gitDir string, entries []index.Entry) (string, error) {
	type dirEntry struct {
		name  string
		hash  string
		mode  uint32
		isDir bool
	}
	
	dirs := make(map[string][]dirEntry)
	
	for _, entry := range entries {
		parts := strings.Split(entry.Path, string(filepath.Separator))
		
		if len(parts) == 1 {
			dirs[""] = append(dirs[""], dirEntry{
				name:  parts[0],
				hash:  entry.Hash,
				mode:  entry.Mode,
				isDir: false,
			})
		} else {
			for i := 0; i < len(parts)-1; i++ {
				parent := filepath.Join(parts[:i]...)
				dirName := parts[i]
				
				found := false
				for _, e := range dirs[parent] {
					if e.name == dirName && e.isDir {
						found = true
						break
					}
				}
				if !found {
					dirs[parent] = append(dirs[parent], dirEntry{
						name:  dirName,
						mode:  040000,
						isDir: true,
					})
				}
			}
			
			parent := filepath.Join(parts[:len(parts)-1]...)
			dirs[parent] = append(dirs[parent], dirEntry{
				name:  parts[len(parts)-1],
				hash:  entry.Hash,
				mode:  entry.Mode,
				isDir: false,
			})
		}
	}
	
	var buildTreeRec func(string) (string, error)
	buildTreeRec = func(path string) (string, error) {
		dirEntries := dirs[path]
		var treeEntries []object.TreeEntry
		
		for _, de := range dirEntries {
			if de.isDir {
				subPath := de.name
				if path != "" {
					subPath = filepath.Join(path, de.name)
				}
				subTreeHash, err := buildTreeRec(subPath)
				if err != nil {
					return "", err
				}
				treeEntries = append(treeEntries, object.TreeEntry{
					Mode: 040000,
					Name: de.name,
					Hash: subTreeHash,
				})
			} else {
				treeEntries = append(treeEntries, object.TreeEntry{
					Mode: de.mode,
					Name: de.name,
					Hash: de.hash,
				})
			}
		}
		
		treeObj := object.NewTree(treeEntries)
		return object.WriteObject(gitDir, treeObj)
	}
	
	return buildTreeRec("")

}