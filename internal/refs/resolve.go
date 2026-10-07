package refs

import (
	"fmt"
	"strconv"
	"strings"
	
	"github.com/om-gaikwad1024/redgit/internal/object"
)


func ResolveRevision(gitDir, rev string) (string, error) {
	
	if strings.Contains(rev, "^") {
		parts := strings.Split(rev, "^")
		base := parts[0]
		caretCount := len(parts) - 1
		
		
		commit, err := ResolveRef(gitDir, base)
		if err != nil {
			return "", err
		}
		
		
		for i := 0; i < caretCount; i++ {
			obj, err := object.ReadObject(gitDir, commit)
			if err != nil {
				return "", fmt.Errorf("cannot resolve %s: %v", rev, err)
			}
			
			if obj.Type != object.CommitType {
				return "", fmt.Errorf("%s is not a commit", commit)
			}
			
			parsedCommit, err := object.ParseCommit(obj.Content)
			if err != nil {
				return "", err
			}
			
			if len(parsedCommit.Parents) == 0 {
				return "", fmt.Errorf("%s has no parent", commit)
			}
			
			commit = parsedCommit.Parents[0]
		}
		
		return commit, nil
	}
	
	
	if strings.Contains(rev, "~") {
		parts := strings.SplitN(rev, "~", 2)
		base := parts[0]
		
		count := 1
		if len(parts) > 1 && parts[1] != "" {
			var err error
			count, err = strconv.Atoi(parts[1])
			if err != nil {
				return "", fmt.Errorf("invalid revision: %s", rev)
			}
		}
		
		
		commit, err := ResolveRef(gitDir, base)
		if err != nil {
			return "", err
		}
		
		
		for i := 0; i < count; i++ {
			obj, err := object.ReadObject(gitDir, commit)
			if err != nil {
				return "", fmt.Errorf("cannot resolve %s: %v", rev, err)
			}
			
			if obj.Type != object.CommitType {
				return "", fmt.Errorf("%s is not a commit", commit)
			}
			
			parsedCommit, err := object.ParseCommit(obj.Content)
			if err != nil {
				return "", err
			}
			
			if len(parsedCommit.Parents) == 0 {
				return "", fmt.Errorf("%s has no parent", commit)
			}
			
			commit = parsedCommit.Parents[0]
		}
		
		return commit, nil
	}
	
	
	if len(rev) > 0 && rev[len(rev)-1] >= '0' && rev[len(rev)-1] <= '9' {
		
		for i := len(rev) - 1; i >= 0; i-- {
			if rev[i] == '^' {
				base := rev[:i]
				parentNum, _ := strconv.Atoi(rev[i+1:])
				
				commit, err := ResolveRef(gitDir, base)
				if err != nil {
					return "", err
				}
				
				obj, err := object.ReadObject(gitDir, commit)
				if err != nil {
					return "", err
				}
				
				parsedCommit, err := object.ParseCommit(obj.Content)
				if err != nil {
					return "", err
				}
				
				if parentNum < 1 || parentNum > len(parsedCommit.Parents) {
					return "", fmt.Errorf("parent %d does not exist", parentNum)
				}
				
				return parsedCommit.Parents[parentNum-1], nil
			}
			
			if rev[i] < '0' || rev[i] > '9' {
				break
			}
		}
	}
	

	return ResolveRef(gitDir, rev)
}