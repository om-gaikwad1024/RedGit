package main

import (
	"fmt"
	"os"
	
	"github.com/om-gaikwad1024/redgit/internal/plumbing"
	"github.com/om-gaikwad1024/redgit/internal/porcelain"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	
	command := os.Args[1]
	args := os.Args[2:]
	
	var err error
	
	switch command {
	case "init":
		err = porcelain.Init(args)
	case "add":
		err = porcelain.Add(args)
	case "commit":
		err = porcelain.Commit(args)
	case "log":
		err = porcelain.Log(args)
	case "branch":
		err = porcelain.Branch(args)
	case "checkout":
		err = porcelain.Checkout(args)
	case "merge":
		err = porcelain.Merge(args)
	case "hash-object":
		err = plumbing.HashObject(args)
	case "cat-file":
		err = plumbing.CatFile(args)
	case "clone":
		err = porcelain.Clone(args)
	case "fetch":
		err = porcelain.Fetch(args)
	case "push":
		err = porcelain.Push(args)
	case "pull":
		err = porcelain.Pull(args)
	case "status":
		err = porcelain.Status(args)
		case "stash":
		err = porcelain.Stash(args)
	case "rebase":
		err = porcelain.Rebase(args)
	case "tag":
		err = porcelain.Tag(args)
	case "reset":
		err = porcelain.Reset(args)
	case "diff":
		err = porcelain.Diff(args)
	case "remote":
		err = porcelain.Remote(args)
	case "show":
		err = porcelain.Show(args)
	case "gc":
		err = porcelain.GC(args)
	case "rev-parse":
		err = plumbing.RevParse(args)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	usage := `redgit - A Git implementation in Go

Usage:
    redgit <command> [arguments]

Porcelain commands:
    init [directory]              Initialize a new repository
    add <file>...                 Add file(s) to staging area
    commit -m <message>           Create a new commit
    log [options]                 Show commit history
        --oneline                 Compact output
        -p, --patch               Show diffs
        --graph                   Show commit graph
        --author <name>           Filter by author
        --since <date>            Filter by date (YYYY-MM-DD)
        --until <date>            Filter by date
        -n <count>                Limit number of commits
    status                        Show working tree status
    branch                        List branches
    branch <name>                 Create a new branch
    branch -d <name>              Delete a branch
    checkout <branch>             Switch to a branch
    checkout -b <branch>          Create and switch to a new branch
    merge <branch>                Merge a branch into current branch
    rebase <branch>               Rebase current branch onto another
    stash [save|list|pop|apply|drop|clear]  Stash changes
    clone <source> [dest]         Clone a repository
    fetch [remote]                Fetch from remote
    push [remote] [branch]        Push to remote
    pull [remote] [branch]        Pull from remote
	tag                           List tags
    tag <name> [commit]           Create lightweight tag
    tag -a <name> -m <message>    Create annotated tag
    tag -d <name>                 Delete tag
	reset [--soft|--mixed|--hard] <commit>  Reset current HEAD
	diff [--cached] [commit] [commit]  Show changes
	remote [-v]                   List remotes
    remote add <name> <url>       Add remote
    remote remove <name>          Remove remote
    remote show <name>            Show remote info
	show [commit]                 Show commit details
	gc                            Garbage collection and pack objects	

Plumbing commands:
    hash-object [-w] <file>       Compute object hash
    cat-file [-t|-p] <hash>       Show object type or content
	rev-parse <revision>          Parse revision (HEAD^, HEAD~2, etc.)
`
	fmt.Print(usage)
}