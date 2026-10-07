package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"os"
	"path/filepath"
)

func Remote(args []string) error {
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	if len(args) == 0 {
		return listRemotes(repo)
	}
	
	switch args[0] {
	case "add":
		if len(args) < 3 {
			return fmt.Errorf("usage: redgit remote add <name> <url>")
		}
		return addRemote(repo, args[1], args[2])
	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("usage: redgit remote remove <name>")
		}
		return removeRemote(repo, args[1])
	case "show":
		if len(args) < 2 {
			return fmt.Errorf("usage: redgit remote show <name>")
		}
		return showRemote(repo, args[1])
	case "-v", "--verbose":
		return listRemotesVerbose(repo)
	default:
		return fmt.Errorf("unknown remote command: %s", args[0])
	}
}

func listRemotes(repo *core.Repository) error {
	config, err := core.LoadConfig(repo)
	if err != nil {
		return err
	}
	
	sections := config.GetSections()
	for section := range sections {
		if len(section) > 8 && section[:7] == "remote " {
			remoteName := section[8 : len(section)-1]
			fmt.Println(remoteName)
		}
	}
	
	return nil
}

func listRemotesVerbose(repo *core.Repository) error {
	config, err := core.LoadConfig(repo)
	if err != nil {
		return err
	}
	
	sections := config.GetSections()
	for section := range sections {
		if len(section) > 8 && section[:7] == "remote " {
			remoteName := section[8 : len(section)-1]
			url, _ := config.Get(section, "url")
			fmt.Printf("%s\t%s (fetch)\n", remoteName, url)
			fmt.Printf("%s\t%s (push)\n", remoteName, url)
		}
	}
	
	return nil
}
func addRemote(repo *core.Repository, name, url string) error {
	config, err := core.LoadConfig(repo)
	if err != nil {
		return err
	}
	
	section := fmt.Sprintf("remote \"%s\"", name)
	config.Set(section, "url", url)
	config.Set(section, "fetch", fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", name))
	
	return config.Save(repo)
}

func removeRemote(repo *core.Repository, name string) error {
	config, err := core.LoadConfig(repo)
	if err != nil {
		return err
	}
	
	section := fmt.Sprintf("remote \"%s\"", name)
	sections := config.GetSections()
	delete(sections, section)
	
	remotePath := filepath.Join(repo.GitDir, "refs", "remotes", name)
	os.RemoveAll(remotePath)
	
	return config.Save(repo)
}

func showRemote(repo *core.Repository, name string) error {
	config, err := core.LoadConfig(repo)
	if err != nil {
		return err
	}
	
	section := fmt.Sprintf("remote \"%s\"", name)
	url, ok := config.Get(section, "url")
	if !ok {
		return fmt.Errorf("remote '%s' not found", name)
	}
	
	fetch, _ := config.Get(section, "fetch")
	
	fmt.Printf("* remote %s\n", name)
	fmt.Printf("  Fetch URL: %s\n", url)
	fmt.Printf("  Push  URL: %s\n", url)
	if fetch != "" {
	fmt.Printf("  Fetch spec: %s\n", fetch)
}

return nil
}