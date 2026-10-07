package porcelain

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"strings"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"github.com/om-gaikwad1024/redgit/internal/pack"
)

func GC(args []string) error {
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	fmt.Println("Collecting garbage...")
	
	
	reachable := findReachableObjects(repo.GitDir)
	
	
	loose := findLooseObjects(repo.GitDir)
	
	
	fmt.Printf("Packing %d objects...\n", len(reachable))
	if err := packObjects(repo.GitDir, reachable); err != nil {
		return err
	}
	
	
	removed := 0
	for hash := range loose {
		if !reachable[hash] {
			objPath := filepath.Join(repo.GitDir, "objects", hash[:2], hash[2:])
			os.Remove(objPath)
			removed++
		}
	}
	
	
	for hash := range reachable {
		if loose[hash] {
			objPath := filepath.Join(repo.GitDir, "objects", hash[:2], hash[2:])
			os.Remove(objPath)
		}
	}
	
	fmt.Printf("Removed %d unreachable objects\n", removed)
	fmt.Println("Done.")
	
	return nil
}

func findReachableObjects(gitDir string) map[string]bool {
	reachable := make(map[string]bool)
	visited := make(map[string]bool)
	
	// Start from all refs
	allRefs := getAllRefs(gitDir)
	
	for _, commit := range allRefs {
		walkObjects(gitDir, commit, reachable, visited)
	}
	
	return reachable
}

func getAllRefs(gitDir string) []string {
	var refHashes []string
	
	// Branches
	branchesDir := filepath.Join(gitDir, "refs", "heads")
	filepath.Walk(branchesDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			data, _ := os.ReadFile(path)
			hash := string(data)
			if len(hash) > 0 {
				refHashes = append(refHashes, hash[:len(hash)-1]) // Remove newline
			}
		}
		return nil
	})
	
	// Tags
	tagsDir := filepath.Join(gitDir, "refs", "tags")
	filepath.Walk(tagsDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			data, _ := os.ReadFile(path)
			hash := string(data)
			if len(hash) > 0 {
				refHashes = append(refHashes, hash[:len(hash)-1])
			}
		}
		return nil
	})
	
	// Remote refs
	remotesDir := filepath.Join(gitDir, "refs", "remotes")
	filepath.Walk(remotesDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			data, _ := os.ReadFile(path)
			hash := string(data)
			if len(hash) > 0 {
				refHashes = append(refHashes, hash[:len(hash)-1])
			}
		}
		return nil
	})
	
	// HEAD
	headPath := filepath.Join(gitDir, "HEAD")
	headData, err := os.ReadFile(headPath)
	if err == nil {
		headContent := string(headData)
		if len(headContent) > 5 && headContent[:5] == "ref: " {
			refName := headContent[5 : len(headContent)-1]
			refPath := filepath.Join(gitDir, refName)
			refData, err := os.ReadFile(refPath)
			if err == nil && len(refData) > 0 {
				refHashes = append(refHashes, string(refData[:len(refData)-1]))
			}
		} else if len(headContent) == 41 {
			refHashes = append(refHashes, headContent[:40])
		}
	}
	
	return refHashes
}

func walkObjects(gitDir, hash string, reachable, visited map[string]bool) {
	hash = strings.TrimSpace(hash)
	
	if len(hash) != 40 {
		return
	}
	
	if visited[hash] {
		return
	}
	visited[hash] = true
	reachable[hash] = true
	
	obj, err := object.ReadObject(gitDir, hash)
	if err != nil {
		return
	}
	
	switch obj.Type {
	case object.CommitType:
		commit, _ := object.ParseCommit(obj.Content)
		walkObjects(gitDir, commit.Tree, reachable, visited)
		for _, parent := range commit.Parents {
			walkObjects(gitDir, parent, reachable, visited)
		}
		
	case object.TreeType:
		entries, _ := object.ParseTree(obj.Content)
		for _, entry := range entries {
			walkObjects(gitDir, entry.Hash, reachable, visited)
		}
	}
}

func findLooseObjects(gitDir string) map[string]bool {
	loose := make(map[string]bool)
	
	objectsDir := filepath.Join(gitDir, "objects")
	dirs, _ := os.ReadDir(objectsDir)
	
	for _, dir := range dirs {
		if !dir.IsDir() || len(dir.Name()) != 2 {
			continue
		}
		
		subDir := filepath.Join(objectsDir, dir.Name())
		files, _ := os.ReadDir(subDir)
		
		for _, file := range files {
			if !file.IsDir() && len(file.Name()) == 38 {
				hash := dir.Name() + file.Name()
				loose[hash] = true
			}
		}
	}
	
	return loose
}

func packObjects(gitDir string, objects map[string]bool) error {
	var packObjs []pack.PackedObject
	
	for hash := range objects {
		obj, err := object.ReadObject(gitDir, hash)
		if err != nil {
			continue
		}
		
		packObjs = append(packObjs, pack.PackedObject{
			Type: pack.TypeToPackType(string(obj.Type)),
			Size: uint64(len(obj.Content)),
			Data: obj.Content,
		})
	}
	
	if len(packObjs) == 0 {
		return nil
	}
	
	packData, err := pack.CreatePackFile(packObjs)
	if err != nil {
		return err
	}
	
	
	packDir := filepath.Join(gitDir, "objects", "pack")
	os.MkdirAll(packDir, 0755)
	
	timestamp := time.Now().Unix()
	packPath := filepath.Join(packDir, fmt.Sprintf("pack-%d.pack", timestamp))
	
	return os.WriteFile(packPath, packData, 0644)
}