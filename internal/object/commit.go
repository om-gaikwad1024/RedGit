package object

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Commit struct {
	Tree      string
	Parents   []string
	Author    string
	Committer string
	Message   string
}

func NewCommit(tree string, parents []string, author, committer, message string) *Object {
	var buf bytes.Buffer
	
	fmt.Fprintf(&buf, "tree %s\n", tree)
	for _, parent := range parents {
		fmt.Fprintf(&buf, "parent %s\n", parent)
	}
	
	timestamp := time.Now().Unix()
	fmt.Fprintf(&buf, "author %s %d +0000\n", author, timestamp)
	fmt.Fprintf(&buf, "committer %s %d +0000\n", committer, timestamp)
	fmt.Fprintf(&buf, "\n%s\n", message)
	
	return &Object{
		Type:    CommitType,
		Content: buf.Bytes(),
	}
}

func ParseCommit(content []byte) (*Commit, error) {
	commit := &Commit{
		Parents: []string{},
	}
	
	lines := strings.Split(string(content), "\n")
	var messageStart int
	
	for i, line := range lines {
		if line == "" {
			messageStart = i + 1
			break
		}
		
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		
		key := parts[0]
		value := parts[1]
		
		switch key {
		case "tree":
			commit.Tree = value
		case "parent":
			commit.Parents = append(commit.Parents, value)
		case "author":
			commit.Author = value
		case "committer":
			commit.Committer = value
		}
	}
	
	if messageStart < len(lines) {
		commit.Message = strings.Join(lines[messageStart:], "\n")
	}
	
	return commit, nil
}

func (c *Commit) GetAuthorName() string {
	parts := strings.Fields(c.Author)
	if len(parts) >= 2 {
		return strings.Join(parts[:len(parts)-2], " ")
	}
	return c.Author
}

func (c *Commit) GetTimestamp() int64 {
	parts := strings.Fields(c.Author)
	if len(parts) >= 2 {
		ts, _ := strconv.ParseInt(parts[len(parts)-2], 10, 64)
		return ts
	}
	return 0
}