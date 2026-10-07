# REDGIT - Git Implementation in Go

A fully functional Git implementation written in Go from scratch, implementing core Git features including content-addressable storage, branching, merging, and more.

## Features

- ✅ Repository initialization
- ✅ Staging files (add)
- ✅ Creating commits
- ✅ Viewing commit history
- ✅ Branch management
- ✅ Checkout branches
- ✅ Three-way merge with conflict detection
- ✅ Content-addressable object storage
- ✅ SHA-1 hashing
- ✅ Zlib compression

## Installation
```bash
go build -o bin/redgit cmd/redgit/main.go
```

## Usage

### Initialize a repository
```bash
./bin/redgit init my-repo
cd my-repo
```

### Add files
```bash
echo "Hello World" > file.txt
./bin/redgit add file.txt
```

### Commit
```bash
export GIT_AUTHOR_NAME="Om Gaikwad"
export GIT_AUTHOR_EMAIL="om.gaikwad0220@gmail.com"
./bin/redgit commit -m "Initial commit"
```

### View history
```bash
./bin/redgit log
./bin/redgit log --oneline
```

### Branch operations
```bash
./bin/redgit branch feature
./bin/redgit checkout feature
./bin/redgit checkout -b new-feature
```

### Merge
```bash
./bin/redgit checkout main
./bin/redgit merge feature
```

## Architecture

- **Content-Addressable Storage**: All objects (blobs, trees, commits) are stored by SHA-1 hash
- **DAG of Commits**: Commits form a directed acyclic graph
- **Three-Way Merge**: Implements proper merge algorithm with conflict detection
- **Index (Staging Area)**: Binary format compatible with Git

## Project Structure
````
redgit/
├── cmd/redgit/           # CLI entry point
├── internal/
│   ├── core/            # Repository management
│   ├── object/          # Object storage (blob, tree, commit)
│   ├── index/           # Staging area
│   ├── refs/            # Branch/tag management
│   ├── merge/           # Merge algorithms
│   ├── plumbing/        # Low-level commands
│   └── porcelain/       # High-level commands
└── test/                # Tests


License
MIT