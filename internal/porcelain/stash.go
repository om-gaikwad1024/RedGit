package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/index"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"github.com/om-gaikwad1024/redgit/internal/refs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StashEntry struct {
	Message     string
	Timestamp   time.Time
	BaseCommit  string
	StashCommit string
}

func Stash(args []string) error {
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return stashSave(repo, "WIP on branch")
	}

	switch args[0] {
	case "save":
		message := "WIP on branch"
		if len(args) > 1 {
			message = strings.Join(args[1:], " ")
		}
		return stashSave(repo, message)
	case "list":
		return stashList(repo)
	case "pop":
		return stashPop(repo)
	case "apply":
		return stashApply(repo, false)
	case "drop":
		return stashDrop(repo)
	case "clear":
		return stashClear(repo)
	default:
		return fmt.Errorf("unknown stash command: %s", args[0])
	}
}

func stashSave(repo *core.Repository, message string) error {
	currentCommit, err := refs.ResolveRef(repo.GitDir, "HEAD")
	if err != nil {
		return fmt.Errorf("no commits yet")
	}

	idx, err := index.ReadIndex(repo.GitDir)
	if err != nil {
		return err
	}

	if len(idx.Entries) == 0 {
		return fmt.Errorf("no changes to stash")
	}

	treeHash, err := buildTree(repo.GitDir, idx.Entries, "")
	if err != nil {
		return err
	}

	author := getAuthor()
	timestamp := time.Now()

	stashMessage := fmt.Sprintf("WIP on branch: %s (%s)", message, timestamp.Format("2006-01-02 15:04:05"))

	stashCommit := object.NewCommit(treeHash, []string{currentCommit}, author, author, stashMessage)
	stashHash, err := object.WriteObject(repo.GitDir, stashCommit)
	if err != nil {
		return err
	}

	stashDir := filepath.Join(repo.GitDir, "refs", "stash")
	os.MkdirAll(stashDir, 0755)

	stashFile := filepath.Join(stashDir, fmt.Sprintf("stash-%d", timestamp.Unix()))
	stashData := fmt.Sprintf("%s\n%s\n%s\n", stashHash, currentCommit, message)
	if err := os.WriteFile(stashFile, []byte(stashData), 0644); err != nil {
		return err
	}

	// FIX: Cleared slice efficiently without unused loop variable
	idx.Entries = idx.Entries[:0]
	idx.WriteIndex(repo.GitDir)

	workingFiles, _ := filepath.Glob(filepath.Join(repo.WorkingDir, "*"))
	for _, file := range workingFiles {
		if filepath.Base(file) != ".git" {
			os.RemoveAll(file)
		}
	}

	fmt.Printf("Saved working directory and index state: %s\n", message)
	return nil
}

func stashList(repo *core.Repository) error {
	stashDir := filepath.Join(repo.GitDir, "refs", "stash")
	entries, err := os.ReadDir(stashDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No stashes found")
			return nil
		}
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No stashes found")
		return nil
	}

	for i, entry := range entries {
		stashFile := filepath.Join(stashDir, entry.Name())
		data, _ := os.ReadFile(stashFile)
		lines := strings.Split(string(data), "\n")
		if len(lines) >= 3 {
			message := lines[2]
			fmt.Printf("stash@{%d}: %s\n", i, message)
		}
	}

	return nil
}

func stashPop(repo *core.Repository) error {
	if err := stashApply(repo, true); err != nil {
		return err
	}
	return stashDrop(repo)
}

func stashApply(repo *core.Repository, isPop bool) error {
	stashDir := filepath.Join(repo.GitDir, "refs", "stash")
	entries, err := os.ReadDir(stashDir)
	if err != nil || len(entries) == 0 {
		return fmt.Errorf("no stash entries found")
	}

	latestStash := entries[len(entries)-1]
	stashFile := filepath.Join(stashDir, latestStash.Name())
	data, err := os.ReadFile(stashFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("invalid stash format")
	}

	stashHash := strings.TrimSpace(lines[0])

	stashObj, err := object.ReadObject(repo.GitDir, stashHash)
	if err != nil {
		return err
	}

	stashCommit, err := object.ParseCommit(stashObj.Content)
	if err != nil {
		return err
	}

	treeObj, err := object.ReadObject(repo.GitDir, stashCommit.Tree)
	if err != nil {
		return err
	}

	// FIX: Renamed variable to treeEntries to avoid conflict with 'entries' (os.ReadDir)
	// This also fixes the Type Mismatch errors below.
	treeEntries, err := object.ParseTree(treeObj.Content)
	if err != nil {
		return err
	}

	idx, _ := index.ReadIndex(repo.GitDir)

	// FIX: Iterating over treeEntries (Objects) instead of entries (Filesystem)
	for _, entry := range treeEntries {
		// Note: Assuming object.TreeEntry has a method IsDirectory() or you check entry.Mode
		// If IsDirectory() is missing, you might need to check: if entry.Mode != "40000"
		if !entry.IsDirectory() {
			blobObj, _ := object.ReadObject(repo.GitDir, entry.Hash)
			filePath := filepath.Join(repo.WorkingDir, entry.Name)
			os.MkdirAll(filepath.Dir(filePath), 0755)

			// FIX: Assuming entry.Mode is an integer/compatible type based on your original code
			os.WriteFile(filePath, blobObj.Content, os.FileMode(entry.Mode))

			info, _ := os.Stat(filePath)
			idxEntry := index.NewEntry(entry.Name, entry.Hash, info)
			idx.AddEntry(idxEntry)
		}
	}

	idx.WriteIndex(repo.GitDir)

	if isPop {
		fmt.Println("Dropped refs/stash")
	} else {
		fmt.Println("Applied stash")
	}

	return nil
}

func stashDrop(repo *core.Repository) error {
	stashDir := filepath.Join(repo.GitDir, "refs", "stash")
	entries, err := os.ReadDir(stashDir)
	if err != nil || len(entries) == 0 {
		return fmt.Errorf("no stash entries to drop")
	}

	latestStash := entries[len(entries)-1]
	stashFile := filepath.Join(stashDir, latestStash.Name())
	return os.Remove(stashFile)
}

func stashClear(repo *core.Repository) error {
	stashDir := filepath.Join(repo.GitDir, "refs", "stash")
	return os.RemoveAll(stashDir)
}