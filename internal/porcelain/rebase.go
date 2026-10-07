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
	"strings"
)

func Rebase(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: redgit rebase <branch>")
	}
	
	targetBranch := args[0]
	
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	currentCommit, err := refs.ResolveRef(repo.GitDir, "HEAD")
	if err != nil {
		return fmt.Errorf("no commits on current branch")
	}
	
	targetCommit, err := refs.ResolveRef(repo.GitDir, "refs/heads/"+targetBranch)
	if err != nil {
		return fmt.Errorf("branch '%s' not found", targetBranch)
	}
	
	if currentCommit == targetCommit {
		fmt.Println("Current branch is up to date.")
		return nil
	}
	
	mergeBase, err := merge.FindMergeBase(repo.GitDir, currentCommit, targetCommit)
	if err != nil || mergeBase == "" {
		return fmt.Errorf("no common ancestor found")
	}
	
	if mergeBase == currentCommit {
		if err := refs.UpdateRef(repo.GitDir, "HEAD", targetCommit); err != nil {
			return err
		}
		if err := merge.ApplyMerge(repo.GitDir, repo.WorkingDir, targetCommit); err != nil {
			return err
		}
		fmt.Println("Fast-forwarded to", targetBranch)
		return nil
	}
	
	commitsToReplay := getCommitList(repo.GitDir, mergeBase, currentCommit)
	
	currentBase := targetCommit
	
	if err := merge.ApplyMerge(repo.GitDir, repo.WorkingDir, targetCommit); err != nil {
		return err
	}
	
	for _, commitHash := range commitsToReplay {
		commitObj, err := object.ReadObject(repo.GitDir, commitHash)
		if err != nil {
			return err
		}
		
		commit, err := object.ParseCommit(commitObj.Content)
		if err != nil {
			return err
		}
		
		if err := merge.ApplyMerge(repo.GitDir, repo.WorkingDir, commitHash); err != nil {
			fmt.Printf("Warning: Could not apply commit %s cleanly\n", commitHash[:7])
		}
		
		idx, _ := index.ReadIndex(repo.GitDir)
		treeHash, err := buildTreeFromIndexRebase(repo.GitDir, idx.Entries)
		if err != nil {
			return err
		}
		
		newCommit := object.NewCommit(treeHash, []string{currentBase}, commit.Author, commit.Committer, commit.Message)
		newCommitHash, err := object.WriteObject(repo.GitDir, newCommit)
		if err != nil {
			return err
		}
		
		fmt.Printf("Reapplying: %s %s\n", commitHash[:7], strings.Split(commit.Message, "\n")[0])
		
		currentBase = newCommitHash
	}
	
	if err := refs.UpdateRef(repo.GitDir, "HEAD", currentBase); err != nil {
		return err
	}
	
	if err := merge.ApplyMerge(repo.GitDir, repo.WorkingDir, currentBase); err != nil {
		return err
	}
	
	currentObj, _ := object.ReadObject(repo.GitDir, currentBase)
	currentCommitParsed, _ := object.ParseCommit(currentObj.Content)
	
	idx := index.NewIndex()
	if err := rebuildIndexFromTreeRebase(repo.GitDir, repo.WorkingDir, currentCommitParsed.Tree, "", idx); err != nil {
		return err
	}
	idx.WriteIndex(repo.GitDir)
	
	fmt.Printf("Successfully rebased onto %s\n", targetBranch)
	return nil
}

func buildTreeFromIndexRebase(gitDir string, entries []index.Entry) (string, error) {
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

func rebuildIndexFromTreeRebase(gitDir, workingDir, treeHash, prefix string, idx *index.Index) error {
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
			if err := rebuildIndexFromTreeRebase(gitDir, workingDir, entry.Hash, fullPath, idx); err != nil {
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

func getCommitList(gitDir, base, head string) []string {
	var commits []string
	visited := make(map[string]bool)
	
	var walk func(string)
	walk = func(commit string) {
		if commit == base || visited[commit] {
			return
		}
		visited[commit] = true
		
		obj, err := object.ReadObject(gitDir, commit)
		if err != nil {
			return
		}
		
		c, err := object.ParseCommit(obj.Content)
		if err != nil {
			return
		}
		
		for _, parent := range c.Parents {
			walk(parent)
		}
		
		commits = append(commits, commit)
	}
	
	walk(head)
	return commits
}