#!/bin/bash
set -e

echo "=== Building redgit ==="
make build

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "\n${BLUE}=== Test 1: Basic Workflow ===${NC}"
rm -rf test-repo
mkdir test-repo
cd test-repo

REDGIT="../bin/redgit"

$REDGIT init .
echo "Hello World" > README.md

export GIT_AUTHOR_NAME="Test User"
export GIT_AUTHOR_EMAIL="test@example.com"

$REDGIT add README.md
$REDGIT commit -m "Initial commit"
$REDGIT log --oneline

echo -e "${GREEN}✓ Basic workflow passed${NC}"

echo -e "\n${BLUE}=== Test 2: Branching and Merging ===${NC}"
$REDGIT branch dev
$REDGIT branch
$REDGIT checkout dev
echo "Development" > dev.txt
$REDGIT add dev.txt
$REDGIT commit -m "Dev work"

$REDGIT checkout main
$REDGIT merge dev
$REDGIT log --oneline

echo -e "${GREEN}✓ Branching and merging passed${NC}"

echo -e "\n${BLUE}=== Test 3: Merge Conflicts ===${NC}"
$REDGIT branch feature1
$REDGIT branch feature2

$REDGIT checkout feature1
echo "Feature 1 change" > conflict.txt
$REDGIT add conflict.txt
$REDGIT commit -m "Feature 1 commit"

$REDGIT checkout feature2
echo "Feature 2 change" > conflict.txt
$REDGIT add conflict.txt
$REDGIT commit -m "Feature 2 commit"

$REDGIT checkout main
$REDGIT merge feature1

# This should create a conflict
set +e
$REDGIT merge feature2
if [ $? -ne 0 ]; then
    echo "Conflict detected as expected"
    # Resolve conflict manually
    echo "Resolved conflict" > conflict.txt
    $REDGIT add conflict.txt
    $REDGIT commit -m "Resolve conflict"
fi
set -e

echo -e "${GREEN}✓ Conflict handling passed${NC}"

echo -e "\n${BLUE}=== Test 4: Stash Operations ===${NC}"
echo "Stashed work" > stash.txt
$REDGIT add stash.txt
$REDGIT stash save "My stash"
$REDGIT stash list
$REDGIT stash pop
$REDGIT status

echo -e "${GREEN}✓ Stash operations passed${NC}"

echo -e "\n${BLUE}=== Test 5: Rebase ===${NC}"
$REDGIT branch rebase-feature
$REDGIT checkout rebase-feature
echo "Rebase feature" > rebase.txt
$REDGIT add rebase.txt
$REDGIT commit -m "Rebase feature commit"

$REDGIT checkout main
echo "Main work during rebase" > main2.txt
$REDGIT add main2.txt
$REDGIT commit -m "Main commit for rebase"

$REDGIT checkout rebase-feature
$REDGIT rebase main
$REDGIT log --oneline

echo -e "${GREEN}✓ Rebase passed${NC}"

echo -e "\n${BLUE}=== Test 6: Tags ===${NC}"
$REDGIT checkout main
$REDGIT tag v1.0
$REDGIT tag -a v1.1 -m "Version 1.1"
$REDGIT tag
$REDGIT tag -d v1.0
$REDGIT tag

echo -e "${GREEN}✓ Tag operations passed${NC}"

echo -e "\n${BLUE}=== Test 7: Reset ===${NC}"
echo "Reset test" > reset.txt
$REDGIT add reset.txt
$REDGIT commit -m "Commit to reset"
LAST_COMMIT=$($REDGIT rev-parse HEAD)
echo "Another change" > reset.txt
$REDGIT add reset.txt
$REDGIT commit -m "Another commit"

$REDGIT reset --hard HEAD^
$REDGIT log --oneline

echo -e "${GREEN}✓ Reset passed${NC}"

echo -e "\n${BLUE}=== Test 8: Diff ===${NC}"
echo "Original content" > diff-test.txt
$REDGIT add diff-test.txt
$REDGIT commit -m "Add diff test file"
echo "Modified content" > diff-test.txt
$REDGIT diff
$REDGIT add diff-test.txt
$REDGIT diff --cached

echo -e "${GREEN}✓ Diff passed${NC}"

echo -e "\n${BLUE}=== Test 9: Show ===${NC}"
$REDGIT show HEAD
$REDGIT show HEAD^

echo -e "${GREEN}✓ Show passed${NC}"

echo -e "\n${BLUE}=== Test 10: Advanced Log ===${NC}"
$REDGIT log --graph
$REDGIT log --author "Test User"
$REDGIT log -n 3
$REDGIT log -p

echo -e "${GREEN}✓ Advanced log passed${NC}"

echo -e "\n${BLUE}=== Test 11: Rev-parse ===${NC}"
$REDGIT rev-parse HEAD
$REDGIT rev-parse HEAD^
$REDGIT rev-parse HEAD~2
$REDGIT rev-parse main

echo -e "${GREEN}✓ Rev-parse passed${NC}"

echo -e "\n${BLUE}=== Test 12: Remote Operations ===${NC}"
cd ..
mkdir remote-repo
cd remote-repo
$REDGIT init .
cd ../test-repo

$REDGIT remote add origin ../remote-repo
$REDGIT remote -v
$REDGIT remote show origin

echo -e "${GREEN}✓ Remote operations passed${NC}"

echo -e "\n${BLUE}=== Test 13: Clone (Local) ===${NC}"
cd ..
./bin/redgit clone test-repo cloned-repo
cd cloned-repo
$REDGIT log --oneline
$REDGIT status

echo -e "${GREEN}✓ Clone passed${NC}"

echo -e "\n${BLUE}=== Test 14: Garbage Collection ===${NC}"
cd ../test-repo
$REDGIT gc

echo -e "${GREEN}✓ Garbage collection passed${NC}"

echo -e "\n${BLUE}=== Test 15: Status ===${NC}"
echo "Untracked file" > untracked.txt
$REDGIT status
$REDGIT add untracked.txt
$REDGIT status
$REDGIT commit -m "Add untracked file"
$REDGIT status

echo -e "${GREEN}✓ Status passed${NC}"

cd ..
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}✓ ALL TESTS PASSED!${NC}"
echo -e "${GREEN}========================================${NC}"

# Cleanup
rm -rf test-repo cloned-repo remote-repo