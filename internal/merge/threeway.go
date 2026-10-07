package merge

import (
	"bytes"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"os"
	"path/filepath"
)

type ConflictFile struct {
	Path         string
	BaseContent  []byte
	OursContent  []byte
	TheirsContent []byte
}

func GetTreeFiles(gitDir, commitHash string) (map[string]string, error) {
	// files := make(map[string]string)
	
	obj, err := object.ReadObject(gitDir, commitHash)
	if err != nil {
		return nil, err
	}
	
	commit, err := object.ParseCommit(obj.Content)
	if err != nil {
		return nil, err
	}
	
	return walkTree(gitDir, commit.Tree, ""), nil
}

func walkTree(gitDir, treeHash, prefix string) map[string]string {
	files := make(map[string]string)
	
	obj, err := object.ReadObject(gitDir, treeHash)
	if err != nil {
		return files
	}
	
	entries, err := object.ParseTree(obj.Content)
	if err != nil {
		return files
	}
	
	for _, entry := range entries {
		fullPath := filepath.Join(prefix, entry.Name)
		
		if entry.IsDirectory() {
			subFiles := walkTree(gitDir, entry.Hash, fullPath)
			for k, v := range subFiles {
				files[k] = v
			}
		} else {
			files[fullPath] = entry.Hash
		}
	}
	
	return files
}

func ThreeWayMerge(gitDir, base, ours, theirs string) ([]ConflictFile, error) {
	baseFiles, _ := GetTreeFiles(gitDir, base)
	oursFiles, _ := GetTreeFiles(gitDir, ours)
	theirsFiles, _ := GetTreeFiles(gitDir, theirs)
	
	allPaths := make(map[string]bool)
	for path := range baseFiles {
		allPaths[path] = true
	}
	for path := range oursFiles {
		allPaths[path] = true
	}
	for path := range theirsFiles {
		allPaths[path] = true
	}
	
	var conflicts []ConflictFile
	
	for path := range allPaths {
		baseHash := baseFiles[path]
		ourHash := oursFiles[path]
		theirHash := theirsFiles[path]
	
	if ourHash == theirHash {
		continue
	}
	
	if ourHash == baseHash && theirHash != "" {
		obj, _ := object.ReadObject(gitDir, theirHash)
		os.MkdirAll(filepath.Dir(path), 0755)
		os.WriteFile(path, obj.Content, 0644)
		continue
	}
	
	if theirHash == baseHash && ourHash != "" {
		continue
	}
	
	if baseHash == "" && ourHash != "" && theirHash != "" {
		conflicts = append(conflicts, ConflictFile{Path: path})
		continue
	}
	
	if ourHash != "" && theirHash != "" && ourHash != theirHash {
		var baseContent, ourContent, theirContent []byte
		
		if baseHash != "" {
			obj, _ := object.ReadObject(gitDir, baseHash)
			baseContent = obj.Content
		}
		if ourHash != "" {
			obj, _ := object.ReadObject(gitDir, ourHash)
			ourContent = obj.Content
		}
		if theirHash != "" {
			obj, _ := object.ReadObject(gitDir, theirHash)
			theirContent = obj.Content
		}
		
		conflicts = append(conflicts, ConflictFile{
			Path:          path,
			BaseContent:   baseContent,
			OursContent:   ourContent,
			TheirsContent: theirContent,
		})
	}
}

return conflicts, nil
}
func MarkConflicts(conflicts []ConflictFile) error {
for _, conflict := range conflicts {
var result bytes.Buffer
result.WriteString("<<<<<<< HEAD\n")
result.Write(conflict.OursContent)
result.WriteString("\n=======\n")
result.Write(conflict.TheirsContent)
result.WriteString("\n>>>>>>> MERGE\n")
	os.MkdirAll(filepath.Dir(conflict.Path), 0755)
	if err := os.WriteFile(conflict.Path, result.Bytes(), 0644); err != nil {
		return err
	}
}
return nil
}
func IsFastForward(gitDir, base, target string) bool {
visited := make(map[string]bool)
queue := []string{target}
for len(queue) > 0 {
	current := queue[0]
	queue = queue[1:]
	
	if current == base {
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
func ApplyMerge(gitDir, workingDir, targetCommit string) error {
obj, err := object.ReadObject(gitDir, targetCommit)
if err != nil {
return err
}
commit, err := object.ParseCommit(obj.Content)
if err != nil {
	return err
}

return MaterializeTree(gitDir, workingDir, commit.Tree, "")
}
func MaterializeTree(gitDir, workingDir, treeHash, prefix string) error {
obj, err := object.ReadObject(gitDir, treeHash)
if err != nil {
return err
}
entries, err := object.ParseTree(obj.Content)
if err != nil {
	return err
}

for _, entry := range entries {
	fullPath := filepath.Join(workingDir, prefix, entry.Name)
	
	if entry.IsDirectory() {
		os.MkdirAll(fullPath, 0755)
		MaterializeTree(gitDir, workingDir, entry.Hash, filepath.Join(prefix, entry.Name))
	} else {
		blobObj, err := object.ReadObject(gitDir, entry.Hash)
		if err != nil {
			return err
		}
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, blobObj.Content, os.FileMode(entry.Mode))
	}
}

return nil
}
func ClearWorkingDirectory(workingDir, gitDir string) error {
entries, err := os.ReadDir(workingDir)
if err != nil {
return err
}
for _, entry := range entries {
	if entry.Name() == ".git" {
		continue
	}
	
	path := filepath.Join(workingDir, entry.Name())
	os.RemoveAll(path)
}

return nil
}