package plumbing

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/object"
)

func CatFile(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: redgit cat-file [-t|-p] <hash>")
	}
	
	flag := args[0]
	hash := args[1]
	
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	obj, err := object.ReadObject(repo.GitDir, hash)
	if err != nil {
		return err
	}
	
	switch flag {
	case "-t":
		fmt.Println(obj.Type)
	case "-p":
		fmt.Print(string(obj.Content))
	default:
		return fmt.Errorf("unknown flag: %s", flag)
	}
	
	return nil
}