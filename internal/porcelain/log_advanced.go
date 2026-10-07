package porcelain

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/om-gaikwad1024/redgit/internal/core"
	"github.com/om-gaikwad1024/redgit/internal/object"
	"github.com/om-gaikwad1024/redgit/internal/refs"
)

type LogOptions struct {
	Oneline    bool
	Graph      bool
	Patch      bool
	Author     string
	Since      time.Time
	Until      time.Time
	MaxCount   int
}

func LogAdvanced(args []string) error {
	opts := parseLogOptions(args)
	
	repo, err := core.FindRepository(".")
	if err != nil {
		return err
	}
	
	commitHash, err := refs.ResolveRef(repo.GitDir, "HEAD")
	if err != nil {
		return fmt.Errorf("no commits yet")
	}
	
	commits := collectCommits(repo.GitDir, commitHash, opts)
	
	if opts.Graph {
		displayCommitsWithGraph(repo.GitDir, commits, opts)
	} else {
		displayCommits(repo.GitDir, commits, opts)
	}
	
	return nil
}

func parseLogOptions(args []string) LogOptions {
	opts := LogOptions{}
	
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--oneline":
			opts.Oneline = true
		case "--graph":
			opts.Graph = true
		case "-p", "--patch":
			opts.Patch = true
		case "--author":
			if i+1 < len(args) {
				opts.Author = args[i+1]
				i++
			}
		case "--since":
			if i+1 < len(args) {
				opts.Since, _ = time.Parse("2006-01-02", args[i+1])
				i++
			}
		case "--until":
			if i+1 < len(args) {
				opts.Until, _ = time.Parse("2006-01-02", args[i+1])
				i++
			}
		case "-n", "--max-count":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &opts.MaxCount)
				i++
			}
		}
	}
	
	return opts
}

func collectCommits(gitDir, startCommit string, opts LogOptions) []string {
	var commits []string
	visited := make(map[string]bool)
	queue := []string{startCommit}
	
	for len(queue) > 0 && (opts.MaxCount == 0 || len(commits) < opts.MaxCount) {
		commit := queue[0]
		queue = queue[1:]
		
		if visited[commit] {
			continue
		}
		visited[commit] = true
		
		obj, err := object.ReadObject(gitDir, commit)
		if err != nil {
			continue
		}
		
		parsedCommit, err := object.ParseCommit(obj.Content)
		if err != nil {
			continue
		}
		
		
		if opts.Author != "" && !strings.Contains(parsedCommit.Author, opts.Author) {
			queue = append(queue, parsedCommit.Parents...)
			continue
		}
		
		// Filter by date
		timestamp := parsedCommit.GetTimestamp()
		commitTime := time.Unix(timestamp, 0)
		
		if !opts.Since.IsZero() && commitTime.Before(opts.Since) {
			continue
		}
		
		if !opts.Until.IsZero() && commitTime.After(opts.Until) {
			queue = append(queue, parsedCommit.Parents...)
			continue
		}
		
		commits = append(commits, commit)
		queue = append(queue, parsedCommit.Parents...)
	}
	
	return commits
}

func displayCommits(gitDir string, commits []string, opts LogOptions) {
	for _, commit := range commits {
		obj, _ := object.ReadObject(gitDir, commit)
		parsedCommit, _ := object.ParseCommit(obj.Content)
		
		if opts.Oneline {
			firstLine := strings.Split(parsedCommit.Message, "\n")[0]
			fmt.Printf("%s %s\n", commit[:7], firstLine)
		} else {
			fmt.Printf("commit %s\n", commit)
			fmt.Printf("Author: %s\n", parsedCommit.Author)
			
			timestamp := parsedCommit.GetTimestamp()
			if timestamp > 0 {
				t := time.Unix(timestamp, 0)
				fmt.Printf("Date:   %s\n", t.Format(time.RFC1123))
			}
			
			fmt.Printf("\n    %s\n", strings.ReplaceAll(parsedCommit.Message, "\n", "\n    "))
			
			if opts.Patch && len(parsedCommit.Parents) > 0 {
				fmt.Println()
				showDiff(gitDir, parsedCommit.Parents[0], commit)
			}
			
			fmt.Println()
		}
	}
}

func displayCommitsWithGraph(gitDir string, commits []string, opts LogOptions) {
	graph := buildCommitGraph(gitDir, commits)
	
	for i, commit := range commits {
		obj, _ := object.ReadObject(gitDir, commit)
		parsedCommit, _ := object.ParseCommit(obj.Content)
		
		// Simple graph visualization
		prefix := getGraphPrefix(graph, commit, i)
		
		if opts.Oneline {
			firstLine := strings.Split(parsedCommit.Message, "\n")[0]
			fmt.Printf("%s %s %s\n", prefix, commit[:7], firstLine)
		} else {
			fmt.Printf("%s commit %s\n", prefix, commit)
			fmt.Printf("%s Author: %s\n", prefix, parsedCommit.Author)
			
			timestamp := parsedCommit.GetTimestamp()
			if timestamp > 0 {
				t := time.Unix(timestamp, 0)
				fmt.Printf("%s Date:   %s\n", prefix, t.Format(time.RFC1123))
			}
			
			fmt.Printf("%s\n", prefix)
			fmt.Printf("%s     %s\n", prefix, strings.ReplaceAll(parsedCommit.Message, "\n", "\n"+prefix+"     "))
			fmt.Println()
		}
	}
}

func buildCommitGraph(gitDir string, commits []string) map[string][]string {
	graph := make(map[string][]string)
	
	for _, commit := range commits {
		obj, _ := object.ReadObject(gitDir, commit)
		parsedCommit, _ := object.ParseCommit(obj.Content)
		graph[commit] = parsedCommit.Parents
	}
	
	return graph
}

func getGraphPrefix(graph map[string][]string, commit string, index int) string {
	parents := graph[commit]
	
	if len(parents) == 0 {
		return "*"
	} else if len(parents) == 1 {
		return "*"
	} else {
		
		return "*   "
	}
}