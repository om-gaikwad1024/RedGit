package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/diff"
	"github.com/om-gaikwad1024/redgit/internal/index"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"github.com/om-gaikwad1024/redgit/internal/refs"
	"os"
	"path/filepath"
)

func Diff(args []string) error {
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	cached := false
	for _, arg := range args {
		if arg == "--cached" || arg == "--staged" {
			cached = true
			break
		}
	}
	
	if cached {
		return diffCached(repo)
	}
	
	if len(args) >= 2 {
		return diffCommits(repo, args[0], args[1])
	}
	
	if len(args) == 1 {
		return diffCommitToWorking(repo, args[0])
	}
	
	return diffIndexToWorking(repo)
}

func diffIndexToWorking(repo *core.Repository) error {
	idx, err := index.ReadIndex(repo.GitDir)
	if err != nil {
		return err
	}
	
	for _, entry := range idx.Entries {
		filePath := filepath.Join(repo.WorkingDir, entry.Path)
		workingContent, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		
		stagedObj, _ := object.ReadObject(repo.GitDir, entry.Hash)
		stagedContent := stagedObj.Content
		
		blob := object.NewBlob(workingContent)
		workingHash := blob.Hash()
		
		if workingHash != entry.Hash {
			fmt.Printf("diff --git a/%s b/%s\n", entry.Path, entry.Path)
			fmt.Printf("--- a/%s\n", entry.Path)
			fmt.Printf("+++ b/%s\n", entry.Path)
			
			d := diff.ComputeDiff(stagedContent, workingContent)
			fmt.Print(d.Format())
		}
	}
	
	return nil
}

func diffCached(repo *core.Repository) error {
	headCommit, err := refs.ResolveRef(repo.GitDir, "HEAD")
	if err != nil {
		return fmt.Errorf("no commits yet")
	}
	
	headFiles := getTreeFiles(repo.GitDir, headCommit)
	
	idx, err := index.ReadIndex(repo.GitDir)
	if err != nil {
		return err
	}
	
	stagedFiles := make(map[string]string)
	for _, entry := range idx.Entries {
		stagedFiles[entry.Path] = entry.Hash
	}
	
	allFiles := make(map[string]bool)
	for path := range headFiles {
		allFiles[path] = true
	}
	for path := range stagedFiles {
		allFiles[path] = true
	}
	
	for path := range allFiles {
		headHash := headFiles[path]
		stagedHash := stagedFiles[path]
		
		if headHash == stagedHash {
			continue
		}
		
		var headContent, stagedContent []byte
		
		if headHash != "" {
			obj, _ := object.ReadObject(repo.GitDir, headHash)
			headContent = obj.Content
		}
		
		if stagedHash != "" {
			obj, _ := object.ReadObject(repo.GitDir, stagedHash)
			stagedContent = obj.Content
		}
		
		if headHash == "" {
			fmt.Printf("diff --git a/%s b/%s\n", path, path)
			fmt.Printf("new file mode 100644\n")
			fmt.Printf("--- /dev/null\n")
			fmt.Printf("+++ b/%s\n", path)
		} else if stagedHash == "" {
			fmt.Printf("diff --git a/%s b/%s\n", path, path)
			fmt.Printf("deleted file mode 100644\n")
			fmt.Printf("--- a/%s\n", path)
			fmt.Printf("+++ /dev/null\n")
		} else {
			fmt.Printf("diff --git a/%s b/%s\n", path, path)
			fmt.Printf("--- a/%s\n", path)
			fmt.Printf("+++ b/%s\n", path)
		}
		
		d := diff.ComputeDiff(headContent, stagedContent)
		fmt.Print(d.Format())
	}
	
	return nil
}

func diffCommitToWorking(repo *core.Repository, commit string) error {
	resolvedCommit, err := refs.ResolveRef(repo.GitDir, commit)
	if err != nil {
		return fmt.Errorf("failed to resolve '%s'", commit)
	}
	
	commitFiles := getTreeFiles(repo.GitDir, resolvedCommit)
	
	for path, hash := range commitFiles {
		filePath := filepath.Join(repo.WorkingDir, path)
		workingContent, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		
		obj, _ := object.ReadObject(repo.GitDir, hash)
		commitContent := obj.Content
		
		blob := object.NewBlob(workingContent)
		if blob.Hash() != hash {
			fmt.Printf("diff --git a/%s b/%s\n", path, path)
			fmt.Printf("--- a/%s\n", path)
			fmt.Printf("+++ b/%s\n", path)
			
			d := diff.ComputeDiff(commitContent, workingContent)
			fmt.Print(d.Format())
		}
	}
	
	return nil
}

func diffCommits(repo *core.Repository, commit1, commit2 string) error {
	resolved1, err := refs.ResolveRef(repo.GitDir, commit1)
	if err != nil {
		return fmt.Errorf("failed to resolve '%s'", commit1)
	}
	
	resolved2, err := refs.ResolveRef(repo.GitDir, commit2)
	if err != nil {
		return fmt.Errorf("failed to resolve '%s'", commit2)
	}
	
	files1 := getTreeFiles(repo.GitDir, resolved1)
	files2 := getTreeFiles(repo.GitDir, resolved2)
	
	allFiles := make(map[string]bool)
	for path := range files1 {
		allFiles[path] = true
	}
	for path := range files2 {
		allFiles[path] = true
	}
	
	for path := range allFiles {
		hash1 := files1[path]
		hash2 := files2[path]
		
		if hash1 == hash2 {
			continue
		}
		
		var content1, content2 []byte
		
		if hash1 != "" {
			obj, _ := object.ReadObject(repo.GitDir, hash1)
			content1 = obj.Content
		}
		
		if hash2 != "" {
			obj, _ := object.ReadObject(repo.GitDir, hash2)
			content2 = obj.Content
		}
		
		if hash1 == "" {
			fmt.Printf("diff --git a/%s b/%s\n", path, path)
			fmt.Printf("new file mode 100644\n")
			fmt.Printf("--- /dev/null\n")
			fmt.Printf("+++ b/%s\n", path)
		} else if hash2 == "" {
			fmt.Printf("diff --git a/%s b/%s\n", path, path)
			fmt.Printf("deleted file mode 100644\n")
			fmt.Printf("--- a/%s\n", path)
			fmt.Printf("+++ /dev/null\n")
		} else {
			fmt.Printf("diff --git a/%s b/%s\n", path, path)
			fmt.Printf("--- a/%s\n", path)
			fmt.Printf("+++ b/%s\n", path)
		}
		
		d := diff.ComputeDiff(content1, content2)
		fmt.Print(d.Format())
	}
	
	return nil
}

func DiffUnified(args []string) error {
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	if len(args) >= 2 {
		return diffCommitsUnified(repo, args[0], args[1])
	}
	
	return diffIndexToWorkingUnified(repo)
}

func diffIndexToWorkingUnified(repo *core.Repository) error {
	idx, err := index.ReadIndex(repo.GitDir)
	if err != nil {
		return err
	}
	
	for _, entry := range idx.Entries {
		filePath := filepath.Join(repo.WorkingDir, entry.Path)
		workingContent, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		
		stagedObj, _ := object.ReadObject(repo.GitDir, entry.Hash)
		stagedContent := stagedObj.Content
		
		blob := object.NewBlob(workingContent)
		workingHash := blob.Hash()
		
		if workingHash != entry.Hash {
			unified := diff.ComputeUnifiedDiff(stagedContent, workingContent, entry.Path, entry.Path)
			fmt.Print(unified)
		}
	}
	
	return nil
}

func diffCommitsUnified(repo *core.Repository, commit1, commit2 string) error {
	resolved1, err := refs.ResolveRevision(repo.GitDir, commit1)
	if err != nil {
		return fmt.Errorf("failed to resolve '%s'", commit1)
	}
	
	resolved2, err := refs.ResolveRevision(repo.GitDir, commit2)
	if err != nil {
		return fmt.Errorf("failed to resolve '%s'", commit2)
	}
	
	files1 := getTreeFiles(repo.GitDir, resolved1)
	files2 := getTreeFiles(repo.GitDir, resolved2)
	
	allFiles := make(map[string]bool)
	for path := range files1 {
		allFiles[path] = true
	}
	for path := range files2 {
		allFiles[path] = true
	}
	
	for path := range allFiles {
		hash1 := files1[path]
		hash2 := files2[path]
		
		if hash1 == hash2 {
			continue
		}
		
		var content1, content2 []byte
		
		if hash1 != "" {
			obj, _ := object.ReadObject(repo.GitDir, hash1)
			content1 = obj.Content
		}
		
		if hash2 != "" {
			obj, _ := object.ReadObject(repo.GitDir, hash2)
			content2 = obj.Content
		}
		
		unified := diff.ComputeUnifiedDiff(content1, content2, path, path)
		fmt.Print(unified)
	}
	
	return nil
}