package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"github.com/om-gaikwad1024/redgit/internal/pack"
	"github.com/om-gaikwad1024/redgit/internal/refs"
	"os"
	"path/filepath"
	"strings"
)

func Push(args []string) error {
	remote := "origin"
	branch := ""
	
	if len(args) == 0 {
		currentBranch, err := refs.GetCurrentBranch("")
		if err != nil {
			return fmt.Errorf("no current branch")
		}
		branch = currentBranch
	} else if len(args) == 1 {
		branch = args[0]
	} else {
		remote = args[0]
		branch = args[1]
	}
	
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	config, err := core.LoadConfig(repo)
	if err != nil {
		return err
	}
	
	remoteURL, ok := config.Get(fmt.Sprintf("remote \"%s\"", remote), "url")
	if !ok {
		return fmt.Errorf("remote '%s' not found", remote)
	}
	
	localCommit, err := refs.ResolveRef(repo.GitDir, "refs/heads/"+branch)
	if err != nil {
		return fmt.Errorf("branch '%s' not found", branch)
	}
	
	remoteGit := filepath.Join(remoteURL, ".git")
	remoteBranchFile := filepath.Join(remoteGit, "refs", "heads", branch)
	
	var remoteCommit string
	if data, err := os.ReadFile(remoteBranchFile); err == nil {
		remoteCommit = strings.TrimSpace(string(data))
	}
	
	if remoteCommit != "" && !isAncestor(repo.GitDir, remoteCommit, localCommit) {
		return fmt.Errorf("non-fast-forward push rejected")
	}
	
	if err := pushObjects(repo.GitDir, remoteGit, localCommit, remoteCommit); err != nil {
		return err
	}
	
	os.MkdirAll(filepath.Dir(remoteBranchFile), 0755)
	if err := os.WriteFile(remoteBranchFile, []byte(localCommit+"\n"), 0644); err != nil {
		return err
	}
	
	fmt.Printf("To %s\n", remoteURL)
	fmt.Printf(" * [new branch]      %s -> %s\n", branch, branch)
	
	return nil
}

func isAncestor(gitDir, ancestor, descendant string) bool {
	visited := make(map[string]bool)
	queue := []string{descendant}
	
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		
		if current == ancestor {
			return true
		}
		
		if visited[current] {
			continue
		}
		visited[current] = true
		
		obj, err := object.ReadObject(gitDir, current)
		if err != nil {
			continue
		}
		
		if obj.Type == object.CommitType {
			commit, _ := object.ParseCommit(obj.Content)
			queue = append(queue, commit.Parents...)
		}
	}
	
	return false
}

func pushObjects(localGit, remoteGit, localCommit, remoteCommit string) error {
	objects, err := collectObjectsToPush(localGit, localCommit, remoteCommit)
	if err != nil {
		return err
	}
	
	if len(objects) == 0 {
		return nil
	}
	
	var packedObjs []pack.PackedObject
	for _, obj := range objects {
		packedObjs = append(packedObjs, pack.PackedObject{
			Type: pack.TypeToPackType(string(obj.Type)),
			Size: uint64(len(obj.Content)),
			Data: obj.Content,
		})
	}
	
	packData, err := pack.CreatePackFile(packedObjs)
	if err != nil {
		return err
	}
	
	packFile, err := pack.UnpackFile(packData)
	if err != nil {
		return err
	}
	
	for _, obj := range packFile.Objects {
		objType := pack.PackTypeToType(obj.Type)
		newObj := &object.Object{
			Type:    object.ObjectType(objType),
			Content: obj.Data,
		}
		hash, _ := object.WriteObject(remoteGit, newObj)
		_ = hash
	}
	
	return nil
}

func collectObjectsToPush(localGit, localCommit, remoteCommit string) ([]fetchObjectInfo, error) {
	visited := make(map[string]bool)
	stopAt := make(map[string]bool)
	
	if remoteCommit != "" {
		stopAt[remoteCommit] = true
	}
	
	var objects []fetchObjectInfo
	queue := []string{localCommit}
	
	for len(queue) > 0 {
		hash := queue[0]
		queue = queue[1:]
		
		if visited[hash] || stopAt[hash] {
			continue
		}
		visited[hash] = true
		
		obj, err := object.ReadObject(localGit, hash)
		if err != nil {
			continue
		}
		
		objects = append(objects, fetchObjectInfo{
			Hash:    hash,
			Type:    obj.Type,
			Content: obj.Content,
		})
		
		if obj.Type == object.CommitType {
			commit, _ := object.ParseCommit(obj.Content)
			queue = append(queue, commit.Tree)
			queue = append(queue, commit.Parents...)
		} else if obj.Type == object.TreeType {
			entries, _ := object.ParseTree(obj.Content)
			for _, entry := range entries {
				queue = append(queue, entry.Hash)
			}
		}
	}
	
	return objects, nil
}
