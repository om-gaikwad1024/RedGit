package porcelain

import (
	"fmt"
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/merge"
	"github.com/om-gaikwad1024/redgit/internal/refs"
	"github.com/om-gaikwad1024/redgit/internal/remote"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

func Clone(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: redgit clone <source> [destination]")
	}
	
	source := args[0]
	dest := filepath.Base(source)
	if len(args) > 1 {
		dest = args[1]
	}
	
	u, err := url.Parse(source)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		return cloneHTTP(source, dest)
	}
	
	return cloneLocal(source, dest)
}

func cloneHTTP(source, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	
	repo, err := core.NewRepository(dest)
	if err != nil {
		return err
	}
	
	if err := repo.Init(); err != nil {
		return err
	}
	
	httpRemote := remote.NewHTTPRemote(source, nil)
	
	remoteRefs, err := httpRemote.DiscoverRefs("git-upload-pack")
	if err != nil {
		return fmt.Errorf("failed to discover refs: %v", err)
	}
	
	var wants []string
	for _, commit := range remoteRefs {
		wants = append(wants, commit)
	}
	
	if len(wants) == 0 {
		return fmt.Errorf("no refs found")
	}
	
	fmt.Printf("Cloning into '%s'...\n", dest)
	
	if err := httpRemote.Fetch(repo.GitDir, wants, []string{}); err != nil {
		return fmt.Errorf("failed to fetch: %v", err)
	}
	
	for branch, commit := range remoteRefs {
		localRef := filepath.Join(repo.GitDir, "refs", "heads", branch)
		os.MkdirAll(filepath.Dir(localRef), 0755)
		os.WriteFile(localRef, []byte(commit+"\n"), 0644)
		
		remoteRef := filepath.Join(repo.GitDir, "refs", "remotes", "origin", branch)
		os.MkdirAll(filepath.Dir(remoteRef), 0755)
		os.WriteFile(remoteRef, []byte(commit+"\n"), 0644)
	}
	
	config, _ := core.LoadConfig(repo)
	config.Set("remote \"origin\"", "url", source)
	config.Set("remote \"origin\"", "fetch", "+refs/heads/*:refs/remotes/origin/*")
	config.Save(repo)
	
	var mainBranch string
	for branch := range remoteRefs {
		if branch == "main" || branch == "master" {
			mainBranch = branch
			break
		}
	}
	if mainBranch == "" {
		for branch := range remoteRefs {
			mainBranch = branch
			break
		}
	}
	
	if mainBranch != "" {
		headContent := fmt.Sprintf("ref: refs/heads/%s\n", mainBranch)
		os.WriteFile(filepath.Join(repo.GitDir, "HEAD"), []byte(headContent), 0644)
		
		headCommit, _ := refs.ResolveRef(repo.GitDir, "HEAD")
		if headCommit != "" {
			currentDir, _ := os.Getwd()
			os.Chdir(dest)
			merge.ApplyMerge(repo.GitDir, repo.WorkingDir, headCommit)
			os.Chdir(currentDir)
		}
	}
	
	fmt.Println("done.")
	return nil
}

func cloneLocal(source, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	
	repo, err := core.NewRepository(dest)
	if err != nil {
		return err
	}
	
	if err := repo.Init(); err != nil {
		return err
	}
	
	sourceGit := filepath.Join(source, ".git")
	if _, err := os.Stat(sourceGit); os.IsNotExist(err) {
		return fmt.Errorf("source is not a git repository")
	}
	
	objectsDir := filepath.Join(sourceGit, "objects")
	destObjectsDir := filepath.Join(repo.GitDir, "objects")
	if err := copyDir(objectsDir, destObjectsDir); err != nil {
		return err
	}
	
	refsDir := filepath.Join(sourceGit, "refs", "heads")
	entries, err := os.ReadDir(refsDir)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			srcFile := filepath.Join(refsDir, entry.Name())
			content, _ := os.ReadFile(srcFile)
			
			localRef := filepath.Join(repo.GitDir, "refs", "heads", entry.Name())
			os.WriteFile(localRef, content, 0644)
			
			remoteRef := filepath.Join(repo.GitDir, "refs", "remotes", "origin", entry.Name())
			os.MkdirAll(filepath.Dir(remoteRef), 0755)
			os.WriteFile(remoteRef, content, 0644)
		}
	}
	
	config, _ := core.LoadConfig(repo)
	config.Set("remote \"origin\"", "url", source)
	config.Set("remote \"origin\"", "fetch", "+refs/heads/*:refs/remotes/origin/*")
	config.Save(repo)
	
	headContent, _ := os.ReadFile(filepath.Join(sourceGit, "HEAD"))
	os.WriteFile(filepath.Join(repo.GitDir, "HEAD"), headContent, 0644)
	
	currentDir, _ := os.Getwd()
	os.Chdir(dest)
	
	headCommit, err := refs.ResolveRef(repo.GitDir, "HEAD")
	if err == nil && headCommit != "" {
		merge.ApplyMerge(repo.GitDir, repo.WorkingDir, headCommit)
	}
	
	os.Chdir(currentDir)
	
	fmt.Printf("Cloning into '%s'...\n", dest)
	fmt.Println("done.")
	
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)
		
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		
		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	
	os.MkdirAll(filepath.Dir(dst), 0755)
	
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	
	_, err = io.Copy(dstFile, srcFile)
	return err
}
