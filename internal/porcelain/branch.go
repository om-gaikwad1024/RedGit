package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/refs"
)

func Branch(args []string) error {
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	if len(args) == 0 {
		return listBranches(repo)
	}
	
	if len(args) >= 2 && args[0] == "-d" {
		return deleteBranch(repo, args[1])
	}
	
	return createBranch(repo, args[0])
}

func listBranches(repo *core.Repository) error {
	branches, err := refs.ListBranches(repo.GitDir)
	if err != nil {
		return err
	}
	
	currentBranch, _ := refs.GetCurrentBranch(repo.GitDir)
	
	for _, branch := range branches {
		if branch == currentBranch {
			fmt.Printf("* %s\n", branch)
		} else {
			fmt.Printf("  %s\n", branch)
		}
	}
	
	return nil
}

func createBranch(repo *core.Repository, name string) error {
	if err := core.ValidateBranchName(name); err != nil {
		return err
	}
	
	currentCommit, err := refs.ResolveRef(repo.GitDir, "HEAD")
	if err != nil {
		return fmt.Errorf("cannot create branch: no commits yet")
	}
	
	return refs.CreateBranch(repo.GitDir, name, currentCommit)
}

func deleteBranch(repo *core.Repository, name string) error {
	currentBranch, _ := refs.GetCurrentBranch(repo.GitDir)
	if name == currentBranch {
		return fmt.Errorf("cannot delete current branch")
	}
	
	return refs.DeleteBranch(repo.GitDir, name)
}