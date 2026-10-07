package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/index"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"github.com/om-gaikwad1024/redgit/internal/refs"
	"os"
	"sort"
	"strings"
)

func Commit(args []string) error {
	if len(args) < 2 || args[0] != "-m" {
		return fmt.Errorf("usage: redgit commit -m <message>")
	}
	
	message := args[1]
	if err := core.ValidateCommitMessage(message); err != nil {
		return err
	}

	if message == "" {
		return fmt.Errorf("empty commit message")
	}
	
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	idx, err := index.ReadIndex(repo.GitDir)
	if err != nil {
		return err
	}
	
	if len(idx.Entries) == 0 {
		return fmt.Errorf("nothing to commit")
	}
	
	treeHash, err := buildTree(repo.GitDir, idx.Entries, "")
	if err != nil {
		return err
	}
	
	author := getAuthor()
	
	var parents []string
	currentCommit, err := refs.ResolveRef(repo.GitDir, "HEAD")
	if err == nil && currentCommit != "" {
		parents = append(parents, currentCommit)
	}
	
	commitObj := object.NewCommit(treeHash, parents, author, author, message)
	commitHash, err := object.WriteObject(repo.GitDir, commitObj)
	if err != nil {
		return err
	}
	
	if err := refs.UpdateRef(repo.GitDir, "HEAD", commitHash); err != nil {
		return err
	}
	
	branch, _ := refs.GetCurrentBranch(repo.GitDir)
	if branch != "" {
		fmt.Printf("[%s %s] %s\n", branch, commitHash[:7], message)
	} else {
		fmt.Printf("[%s] %s\n", commitHash[:7], message)
	}
	
	return nil
}

func buildTree(gitDir string, entries []index.Entry, prefix string) (string, error) {
	type dirEntry struct {
		name  string
		hash  string
		mode  uint32
		isDir bool
	}
	
	dirs := make(map[string][]dirEntry)
	
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Path, prefix) && prefix != "" {
			continue
		}
		
		relPath := entry.Path
		if prefix != "" {
			relPath = strings.TrimPrefix(entry.Path, prefix+"/")
		}
		
		parts := strings.Split(relPath, "/")
		
		if len(parts) == 1 {
			dirs[prefix] = append(dirs[prefix], dirEntry{
				name:  parts[0],
				hash:  entry.Hash,
				mode:  entry.Mode,
				isDir: false,
			})
		} else {
			dirName := parts[0]
			found := false
			for _, e := range dirs[prefix] {
				if e.name == dirName && e.isDir {
					found = true
					break
				}
			}
			if !found {
				dirs[prefix] = append(dirs[prefix], dirEntry{
					name:  dirName,
					hash:  "",
					mode:  040000,
					isDir: true,
				})
			}
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
					subPath = path + "/" + de.name
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
		
		sort.Slice(treeEntries, func(i, j int) bool {
			return treeEntries[i].Name < treeEntries[j].Name
		})
		
		treeObj := object.NewTree(treeEntries)
		return object.WriteObject(gitDir, treeObj)
	}
	
	return buildTreeRec(prefix)
}

func getAuthor() string {
	name := os.Getenv("GIT_AUTHOR_NAME")
	if name == "" {
		name = "User"
	}
	
	email := os.Getenv("GIT_AUTHOR_EMAIL")
	if email == "" {
		email = "user@example.com"
	}
	
	return fmt.Sprintf("%s <%s>", name, email)
}