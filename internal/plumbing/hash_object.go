package plumbing

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"os"
)

func HashObject(args []string) error {
	write := false
	objType := "blob"
	var filePath string
	
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-w":
			write = true
		case "-t":
			if i+1 < len(args) {
				objType = args[i+1]
				i++
			}
		default:
			filePath = args[i]
		}
	}
	
	if filePath == "" {
		return fmt.Errorf("usage: redgit hash-object [-w] [-t type] <file>")
	}
	
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	
	obj := &object.Object{
		Type:    object.ObjectType(objType),
		Content: content,
	}
	
	if write {
		repo, err := core.FindRepository(".")
		if err != nil {
			return err
		}
		
		hash, err := object.WriteObject(repo.GitDir, obj)
		if err != nil {
			return err
		}
		fmt.Println(hash)
	} else {
		fmt.Println(obj.Hash())
	}
	
	return nil
}