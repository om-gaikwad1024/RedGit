package unit

import (
	"testing"
	"github.com/om-gaikwad1024/redgit/internal/object"
)

func TestBlobHash(t *testing.T) {
	content := []byte("Hello World")
	blob := object.NewBlob(content)
	
	hash := blob.Hash()
	expected := "5e1c309dae7f45e0f39b1bf3ac3cd9db12e7d689"
	
	if hash != expected {
		t.Errorf("Expected hash %s, got %s", expected, hash)
	}
}

func TestTreeCreation(t *testing.T) {
	entries := []object.TreeEntry{
		{Mode: 0100644, Name: "file.txt", Hash: "abc123"},
		{Mode: 040000, Name: "dir", Hash: "def456"},
	}
	
	tree := object.NewTree(entries)
	if tree.Type != object.TreeType {
		t.Errorf("Expected tree type, got %s", tree.Type)
	}
}