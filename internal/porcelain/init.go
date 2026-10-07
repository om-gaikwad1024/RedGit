package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
)

func Init(args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}
	
	repo, err := core.NewRepository(path)
	if err != nil {
		return err
	}
	
	if err := repo.Init(); err != nil {
		return err
	}
	
	fmt.Printf("Initialized empty Git repository in %s/.git/\n", repo.WorkingDir)
	return nil
}