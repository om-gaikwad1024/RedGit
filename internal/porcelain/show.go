package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/diff"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"github.com/om-gaikwad1024/redgit/internal/refs"
	"strings"
	"time"
)

func Show(args []string) error {
	target := "HEAD"
	if len(args) > 0 {
		target = args[0]
	}
	
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	
	commitHash, err := refs.ResolveRevision(repo.GitDir, target)
	if err != nil {
		return fmt.Errorf("failed to resolve '%s'", target)
	}
	
	obj, err := object.ReadObject(repo.GitDir, commitHash)
	if err != nil {
		return err
	}
	
	commit, err := object.ParseCommit(obj.Content)
	if err != nil {
		return err
	}
	
	fmt.Printf("commit %s\n", commitHash)
	fmt.Printf("Author: %s\n", commit.Author)
	
	timestamp := commit.GetTimestamp()
	if timestamp > 0 {
		t := time.Unix(timestamp, 0)
		fmt.Printf("Date:   %s\n", t.Format(time.RFC1123))
	}
	
	fmt.Printf("\n    %s\n\n", strings.ReplaceAll(commit.Message, "\n", "\n    "))
	
	if len(commit.Parents) > 0 {
		showDiff(repo.GitDir, commit.Parents[0], commitHash)
	} else {
		showInitialCommit(repo.GitDir, commitHash)
	}
	
	return nil
}

func showInitialCommit(gitDir, commitHash string) {
	files := getTreeFiles(gitDir, commitHash)
	
	for path, hash := range files {
		fmt.Printf("diff --git a/%s b/%s\n", path, path)
		fmt.Printf("new file mode 100644\n")
		fmt.Printf("--- /dev/null\n")
		fmt.Printf("+++ b/%s\n", path)
		
		obj, _ := object.ReadObject(gitDir, hash)
		d := diff.ComputeDiff([]byte{}, obj.Content)
		fmt.Print(d.Format())
	}
}