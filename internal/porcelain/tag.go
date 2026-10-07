package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"github.com/om-gaikwad1024/redgit/internal/refs"
	"os"
	"path/filepath"
	"time"
)

func Tag(args []string) error {
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	if len(args) == 0 {
		return listTags(repo)
	}
	
	if args[0] == "-d" {
		if len(args) < 2 {
			return fmt.Errorf("usage: redgit tag -d <tagname>")
		}
		return deleteTag(repo, args[1])
	}
	
	if args[0] == "-a" {
		if len(args) < 2 {
			return fmt.Errorf("usage: redgit tag -a <tagname> -m <message>")
		}
		
		tagName := args[1]
		message := ""
		
		for i := 2; i < len(args); i++ {
			if args[i] == "-m" && i+1 < len(args) {
				message = args[i+1]
				break
			}
		}
		
		if message == "" {
			return fmt.Errorf("annotated tag requires a message (-m)")
		}
		
		return createAnnotatedTag(repo, tagName, message)
	}
	
	tagName := args[0]
	commit := ""
	if len(args) > 1 {
		commit = args[1]
	}
	
	return createLightweightTag(repo, tagName, commit)
}

func listTags(repo *core.Repository) error {
	tagsDir := filepath.Join(repo.GitDir, "refs", "tags")
	
	if _, err := os.Stat(tagsDir); os.IsNotExist(err) {
		return nil
	}
	
	entries, err := os.ReadDir(tagsDir)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			fmt.Println(entry.Name())
		}
	}
	
	return nil
}

func createLightweightTag(repo *core.Repository, tagName, commitHash string) error {
	
	if tagName == "" {
		return fmt.Errorf("tag name cannot be empty")
	}
	
	
	for _, ch := range tagName {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || 
			(ch >= '0' && ch <= '9') || ch == '.' || ch == '-' || ch == '_') {
			return fmt.Errorf("invalid tag name: contains invalid characters")
		}
	}
	
	if commitHash == "" {
		var err error
		commitHash, err = refs.ResolveRef(repo.GitDir, "HEAD")
		if err != nil {
			return fmt.Errorf("no commits yet")
		}
	} else {
		resolvedHash, err := refs.ResolveRef(repo.GitDir, commitHash)
		if err == nil {
			commitHash = resolvedHash
		}
	}
	
	tagPath := filepath.Join(repo.GitDir, "refs", "tags", tagName)
	os.MkdirAll(filepath.Dir(tagPath), 0755)
	
	return core.AtomicWrite(tagPath, []byte(commitHash+"\n"), 0644)
}

func createAnnotatedTag(repo *core.Repository, tagName, message string) error {
	
	if tagName == "" {
		return fmt.Errorf("tag name cannot be empty")
	}
	
	
	for _, ch := range tagName {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || 
			(ch >= '0' && ch <= '9') || ch == '.' || ch == '-' || ch == '_') {
			return fmt.Errorf("invalid tag name: contains invalid characters")
		}
	}
	
	commitHash, err := refs.ResolveRef(repo.GitDir, "HEAD")
	if err != nil {
		return fmt.Errorf("no commits yet")
	}
	
	author := getAuthor()
	timestamp := time.Now().Unix()
	
	tagContent := fmt.Sprintf("object %s\ntype commit\ntag %s\ntagger %s %d +0000\n\n%s\n",
		commitHash, tagName, author, timestamp, message)
	
	tagObj := &object.Object{
		Type:    "tag",
		Content: []byte(tagContent),
	}
	
	tagHash, err := object.WriteObject(repo.GitDir, tagObj)
	if err != nil {
		return err
	}
	
	tagPath := filepath.Join(repo.GitDir, "refs", "tags", tagName)
	os.MkdirAll(filepath.Dir(tagPath), 0755)
	
	return core.AtomicWrite(tagPath, []byte(tagHash+"\n"), 0644)
}

func deleteTag(repo *core.Repository, tagName string) error {
	tagPath := filepath.Join(repo.GitDir, "refs", "tags", tagName)
	
	if _, err := os.Stat(tagPath); os.IsNotExist(err) {
		return fmt.Errorf("tag '%s' not found", tagName)
	}
	
	return os.Remove(tagPath)
}