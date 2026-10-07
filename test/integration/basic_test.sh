#!/bin/bash
set -e

echo "Running integration tests..."

rm -rf test-repo
mkdir test-repo
cd test-repo

../bin/redgit init .
echo "✓ Init test passed"

echo "Hello World" > file1.txt
../bin/redgit add file1.txt
echo "✓ Add test passed"

export GIT_AUTHOR_NAME="Test User"
export GIT_AUTHOR_EMAIL="test@example.com"
../bin/redgit commit -m "First commit"
echo "✓ Commit test passed"

../bin/redgit log > /dev/null
echo "✓ Log test passed"

../bin/redgit branch feature
../bin/redgit branch | grep feature > /dev/null
echo "✓ Branch test passed"

echo "Feature work" > file2.txt
../bin/redgit add file2.txt
../bin/redgit commit -m "Feature commit"

../bin/redgit checkout main
echo "✓ Checkout test passed"

../bin/redgit merge feature
echo "✓ Merge test passed"

cd ..
rm -rf test-repo

echo "All integration tests passed!"