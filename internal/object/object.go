package object

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"github.com/om-gaikwad1024/redgit/internal/core"
)

type ObjectType string

const (
	BlobType   ObjectType = "blob"
	TreeType   ObjectType = "tree"
	CommitType ObjectType = "commit"
)

type Object struct {
	Type    ObjectType
	Content []byte
}

func (o *Object) Hash() string {
	return ComputeHash(string(o.Type), o.Content)
}

func (o *Object) Serialize() []byte {
	header := fmt.Sprintf("%s %d\x00", o.Type, len(o.Content))
	return append([]byte(header), o.Content...)
}

func WriteObject(gitDir string, obj *Object) (string, error) {
	data := obj.Serialize()
	hash := Hash(data)
	
	objDir := filepath.Join(gitDir, "objects", hash[:2])
	if err := os.MkdirAll(objDir, 0755); err != nil {
		return "", err
	}
	
	objPath := filepath.Join(objDir, hash[2:])
	
	if _, err := os.Stat(objPath); err == nil {
		return hash, nil
	}
	
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	
	if err := core.AtomicWrite(objPath, buf.Bytes(), 0444); err != nil {
		return "", err
	}
	
	return hash, nil
}

func ReadObject(gitDir, hash string) (*Object, error) {
	objPath := filepath.Join(gitDir, "objects", hash[:2], hash[2:])
	
	compressed, err := os.ReadFile(objPath)
	if err != nil {
		return nil, fmt.Errorf("object not found: %s", hash)
	}
	
	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	
	nullIdx := bytes.IndexByte(data, 0)
	if nullIdx == -1 {
		return nil, fmt.Errorf("invalid object format")
	}
	
	header := string(data[:nullIdx])
	content := data[nullIdx+1:]
	
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid object header")
	}
	
	return &Object{
		Type:    ObjectType(parts[0]),
		Content: content,
	}, nil
}

func ObjectExists(gitDir, hash string) bool {
	objPath := filepath.Join(gitDir, "objects", hash[:2], hash[2:])
	_, err := os.Stat(objPath)
	return err == nil
}