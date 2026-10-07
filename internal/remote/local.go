package remote

import (
	"os"
	"path/filepath"
)

type LocalRemote struct {
	Name string
	URL  string
}

func (r *LocalRemote) Fetch() error {
	return nil
}

func (r *LocalRemote) Push(branch string) error {
	return nil
}

func (r *LocalRemote) ListRefs() (map[string]string, error) {
	refs := make(map[string]string)
	
	refsDir := filepath.Join(r.URL, ".git", "refs", "heads")
	entries, err := os.ReadDir(refsDir)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			data, _ := os.ReadFile(filepath.Join(refsDir, entry.Name()))
			refs[entry.Name()] = string(data)
		}
	}
	
	return refs, nil
}