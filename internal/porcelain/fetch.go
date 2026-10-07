package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"github.com/om-gaikwad1024/redgit/internal/pack"
	"os"
	"path/filepath"
)

func Fetch(args []string) error {
	remote := "origin"
	if len(args) > 0 {
		remote = args[0]
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
	
	remoteGit := filepath.Join(remoteURL, ".git")
	if _, err := os.Stat(remoteGit); os.IsNotExist(err) {
		return fmt.Errorf("remote is not a git repository")
	}
	
	remoteRefsDir := filepath.Join(remoteGit, "refs", "heads")
	entries, err := os.ReadDir(remoteRefsDir)
	if err != nil {
		return err
	}
	
	fmt.Printf("From %s\n", remoteURL)
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		branchName := entry.Name()
		remoteRefFile := filepath.Join(remoteRefsDir, branchName)
		commitHashBytes, _ := os.ReadFile(remoteRefFile)
		commitHash := string(commitHashBytes)
		
		if err := fetchWithPack(repo.GitDir, remoteGit, commitHash); err != nil {
			fmt.Printf("Warning: failed to fetch %s: %v\n", branchName, err)
			continue
		}
		
		localRemoteRef := filepath.Join(repo.GitDir, "refs", "remotes", remote, branchName)
		os.MkdirAll(filepath.Dir(localRemoteRef), 0755)
		os.WriteFile(localRemoteRef, commitHashBytes, 0644)
		
		fmt.Printf(" * branch            %s -> %s/%s\n", branchName, remote, branchName)
	}
	
	return nil
}

func fetchWithPack(localGit, remoteGit, targetCommit string) error {
	objects, err := collectObjects(remoteGit, targetCommit)
	if err != nil {
		return err
	}
	
	if len(objects) == 0 {
		return nil
	}
	
	var packedObjs []pack.PackedObject
	for _, obj := range objects {
		if object.ObjectExists(localGit, obj.Hash) {
			continue
		}
		
		packedObjs = append(packedObjs, pack.PackedObject{
			Type: pack.TypeToPackType(string(obj.Type)),
			Size: uint64(len(obj.Content)),
			Data: obj.Content,
		})
	}
	
	if len(packedObjs) == 0 {
		return nil
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
	object.WriteObject(localGit, newObj)
}

return nil
}


func collectObjects(gitDir, commitHash string) ([]fetchObjectInfo, error) {
	visited := make(map[string]bool)
	var objects []fetchObjectInfo
	queue := []string{commitHash}
	
	for len(queue) > 0 {
		hash := queue[0]
		queue = queue[1:]
		
		if visited[hash] {
			continue
		}
		visited[hash] = true
		
		obj, err := object.ReadObject(gitDir, hash)
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

