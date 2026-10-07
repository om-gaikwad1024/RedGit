package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/refs"
)

func Pull(args []string) error {
	remote := "origin"
	branch := ""
	
	if len(args) > 0 {
		remote = args[0]
	}
	if len(args) > 1 {
		branch = args[1]
	}
	
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	if branch == "" {
		branch, err = refs.GetCurrentBranch(repo.GitDir)
		if err != nil {
			return fmt.Errorf("no current branch")
		}
	}
	
	if err := Fetch([]string{remote}); err != nil {
		return err
	}
	
	remoteBranch := fmt.Sprintf("remotes/%s/%s", remote, branch)
	remoteCommit, err := refs.ResolveRef(repo.GitDir, "refs/"+remoteBranch)
	if err != nil {
		return fmt.Errorf("remote branch not found")
	}
	
	currentCommit, _ := refs.ResolveRef(repo.GitDir, "HEAD")
	
	if currentCommit == remoteCommit {
		fmt.Println("Already up to date.")
		return nil
	}
	
	return Merge([]string{remoteBranch})
}