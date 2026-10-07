package plumbing

import (
	"fmt"
	
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/refs"
)

func RevParse(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: redgit rev-parse <revision>")
	}
	
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	revision := args[0]
	
	
	hash, err := refs.ResolveRevision(repo.GitDir, revision)
	if err != nil {
		return err
	}
	
	fmt.Println(hash)
	return nil
}