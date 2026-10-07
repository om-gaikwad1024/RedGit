package refs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/om-gaikwad1024/redgit/internal/core"
)

func ResolveRef(gitDir, ref string) (string, error) {
	if strings.HasPrefix(ref, "refs/") {
		refPath := filepath.Join(gitDir, ref)
		data, err := os.ReadFile(refPath)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}
	
	if ref == "HEAD" {
		headPath := filepath.Join(gitDir, "HEAD")
		data, err := os.ReadFile(headPath)
		if err != nil {
			return "", err
		}
		
		content := strings.TrimSpace(string(data))
		if strings.HasPrefix(content, "ref:") {
			refName := strings.TrimSpace(strings.TrimPrefix(content, "ref:"))
			return ResolveRef(gitDir, refName)
		}
		return content, nil
	}
	
	tryRefs := []string{
		ref,
		"refs/" + ref,
		"refs/heads/" + ref,
		"refs/tags/" + ref,
		"refs/remotes/" + ref,
		"refs/remotes/" + ref + "/HEAD",
	}
	
	for _, tryRef := range tryRefs {
		refPath := filepath.Join(gitDir, tryRef)
		data, err := os.ReadFile(refPath)
		if err == nil {
			return strings.TrimSpace(string(data)), nil
		}
	}
	
	return "", fmt.Errorf("ref not found: %s", ref)
}

func UpdateRef(gitDir, ref, hash string) error {
	var refPath string
	
	if ref == "HEAD" {
		headPath := filepath.Join(gitDir, "HEAD")
		data, err := os.ReadFile(headPath)
		if err != nil {
			return err
		}
		
		content := strings.TrimSpace(string(data))
		if strings.HasPrefix(content, "ref:") {
			refName := strings.TrimSpace(strings.TrimPrefix(content, "ref:"))
			refPath = filepath.Join(gitDir, refName)
		} else {
			refPath = headPath
		}
	} else if strings.HasPrefix(ref, "refs/") {
		refPath = filepath.Join(gitDir, ref)
	} else {
		refPath = filepath.Join(gitDir, "refs", "heads", ref)
	}
	
	dir := filepath.Dir(refPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	return core.AtomicWrite(refPath, []byte(hash+"\n"), 0644)
}

func ListBranches(gitDir string) ([]string, error) {
	branchesDir := filepath.Join(gitDir, "refs", "heads")
	entries, err := os.ReadDir(branchesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	
	var branches []string
	for _, entry := range entries {
		if !entry.IsDir() {
			branches = append(branches, entry.Name())
		}
	}
	return branches, nil
}

func GetCurrentBranch(gitDir string) (string, error) {
	headPath := filepath.Join(gitDir, "HEAD")
	data, err := os.ReadFile(headPath)
	if err != nil {
		return "", err
	}
	
	content := strings.TrimSpace(string(data))
	if strings.HasPrefix(content, "ref: refs/heads/") {
		return strings.TrimPrefix(content, "ref: refs/heads/"), nil
	}
	
	return "", fmt.Errorf("detached HEAD")
}

func CreateBranch(gitDir, name, commit string) error {
	branchPath := filepath.Join(gitDir, "refs", "heads", name)
	return os.WriteFile(branchPath, []byte(commit+"\n"), 0644)
}

func DeleteBranch(gitDir, name string) error {
	branchPath := filepath.Join(gitDir, "refs", "heads", name)
	return os.Remove(branchPath)
}

func SetHEAD(gitDir, ref string) error {
	headPath := filepath.Join(gitDir, "HEAD")
	content := fmt.Sprintf("ref: refs/heads/%s\n", ref)
	return os.WriteFile(headPath, []byte(content), 0644)
}