package merge

import (
	"github.com/om-gaikwad1024/redgit/internal/object"
)

func FindMergeBase(gitDir, commit1, commit2 string) (string, error) {
	visited1 := make(map[string]bool)
	queue1 := []string{commit1}
	
	for len(queue1) > 0 {
		current := queue1[0]
		queue1 = queue1[1:]
		
		if visited1[current] {
			continue
		}
		visited1[current] = true
		
		obj, err := object.ReadObject(gitDir, current)
		if err != nil {
			continue
		}
		
		if obj.Type == object.CommitType {
			commit, _ := object.ParseCommit(obj.Content)
			queue1 = append(queue1, commit.Parents...)
		}
	}
	
	queue2 := []string{commit2}
	visited2 := make(map[string]bool)
	
	for len(queue2) > 0 {
		current := queue2[0]
		queue2 = queue2[1:]
		
		if visited2[current] {
			continue
		}
		visited2[current] = true
		
		if visited1[current] {
			return current, nil
		}
		
		obj, err := object.ReadObject(gitDir, current)
		if err != nil {
			continue
		}
		
		if obj.Type == object.CommitType {
			commit, _ := object.ParseCommit(obj.Content)
			queue2 = append(queue2, commit.Parents...)
		}
	}
	
	return "", nil
}