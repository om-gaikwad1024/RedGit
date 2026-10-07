package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/diff"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"github.com/om-gaikwad1024/redgit/internal/refs"
	"path/filepath"
	"strings"
	"time"
)

func Log(args []string) error {
	
	for _, arg := range args {
		if arg == "--graph" || arg == "--author" || arg == "--since" || arg == "--until" || arg == "-n" || arg == "--max-count" {
			return LogAdvanced(args)
		}
	}
	
	
	oneline := false
	showPatch := false
	
	for _, arg := range args {
		if arg == "--oneline" {
			oneline = true
		} else if arg == "-p" || arg == "--patch" {
			showPatch = true
		}
	}
	
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	commitHash, err := refs.ResolveRef(repo.GitDir, "HEAD")
	if err != nil {
		return fmt.Errorf("no commits yet")
	}
	
	for commitHash != "" {
		obj, err := object.ReadObject(repo.GitDir, commitHash)
		if err != nil {
			break
		}
		
		commit, err := object.ParseCommit(obj.Content)
		if err != nil {
			break
		}
		
		if oneline {
			firstLine := strings.Split(commit.Message, "\n")[0]
			fmt.Printf("%s %s\n", commitHash[:7], firstLine)
		} else {
			fmt.Printf("commit %s\n", commitHash)
			fmt.Printf("Author: %s\n", commit.Author)
			
			timestamp := commit.GetTimestamp()
			if timestamp > 0 {
				t := time.Unix(timestamp, 0)
				fmt.Printf("Date:   %s\n", t.Format(time.RFC1123))
			}
			
			fmt.Printf("\n    %s\n", strings.ReplaceAll(commit.Message, "\n", "\n    "))
			
			if showPatch && len(commit.Parents) > 0 {
				fmt.Println()
				showDiff(repo.GitDir, commit.Parents[0], commitHash)
			}
			
			fmt.Println()
		}
		
		if len(commit.Parents) > 0 {
			commitHash = commit.Parents[0]
		} else {
			break
		}
	}
	
	return nil
}

func showDiff(gitDir, oldCommit, newCommit string) {
	oldFiles := getTreeFiles(gitDir, oldCommit)
	newFiles := getTreeFiles(gitDir, newCommit)
	
	allFiles := make(map[string]bool)
	for path := range oldFiles {
		allFiles[path] = true
	}
	for path := range newFiles {
		allFiles[path] = true
	}
	
	for path := range allFiles {
		oldHash := oldFiles[path]
		newHash := newFiles[path]
		
		if oldHash == newHash {
			continue
		}
		
		var oldContent, newContent []byte
		
		if oldHash != "" {
			obj, _ := object.ReadObject(gitDir, oldHash)
			oldContent = obj.Content
		}
		
		if newHash != "" {
			obj, _ := object.ReadObject(gitDir, newHash)
			newContent = obj.Content
		}
		
		if oldHash == "" {
			fmt.Printf("diff --git a/%s b/%s\n", path, path)
			fmt.Printf("new file mode 100644\n")
			fmt.Printf("--- /dev/null\n")
			fmt.Printf("+++ b/%s\n", path)
		} else if newHash == "" {
			fmt.Printf("diff --git a/%s b/%s\n", path, path)
			fmt.Printf("deleted file mode 100644\n")
			fmt.Printf("--- a/%s\n", path)
			fmt.Printf("+++ /dev/null\n")
		} else {
			fmt.Printf("diff --git a/%s b/%s\n", path, path)
			fmt.Printf("--- a/%s\n", path)
			fmt.Printf("+++ b/%s\n", path)
		}
		
		d := diff.ComputeDiff(oldContent, newContent)
		fmt.Print(d.Format())
	}
}

func getTreeFiles(gitDir, commitHash string) map[string]string {
	files := make(map[string]string)
	
	obj, err := object.ReadObject(gitDir, commitHash)
	if err != nil {
		return files
	}
	
	commit, err := object.ParseCommit(obj.Content)
	if err != nil {
		return files
	}
	
	walkTree(gitDir, commit.Tree, "", files)
	return files
}

func walkTree(gitDir, treeHash, prefix string, files map[string]string) {
	obj, err := object.ReadObject(gitDir, treeHash)
	if err != nil {
		return
	}
	
	entries, err := object.ParseTree(obj.Content)
	if err != nil {
		return
	}
	
	for _, entry := range entries {
		fullPath := filepath.Join(prefix, entry.Name)
		
		if entry.IsDirectory() {
			walkTree(gitDir, entry.Hash, fullPath, files)
		} else {
			files[fullPath] = entry.Hash
		}
	}
}