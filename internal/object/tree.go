package object

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
)

type TreeEntry struct {
	Mode uint32
	Name string
	Hash string
}

func NewTree(entries []TreeEntry) *Object {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	
	var buf bytes.Buffer
	for _, entry := range entries {
		fmt.Fprintf(&buf, "%o %s\x00", entry.Mode, entry.Name)
		hashBytes, _ := hex.DecodeString(entry.Hash)
		buf.Write(hashBytes)
	}
	
	return &Object{
		Type:    TreeType,
		Content: buf.Bytes(),
	}
}

func ParseTree(content []byte) ([]TreeEntry, error) {
	var entries []TreeEntry
	buf := bytes.NewReader(content)
	
	for buf.Len() > 0 {
		modeEnd := bytes.IndexByte(content[len(content)-buf.Len():], ' ')
		if modeEnd == -1 {
			break
		}
		
		modeStr := string(content[len(content)-buf.Len() : len(content)-buf.Len()+modeEnd])
		mode, err := strconv.ParseUint(modeStr, 8, 32)
		if err != nil {
			return nil, err
		}
		
		buf.Seek(int64(modeEnd+1), 1)
		
		remaining := content[len(content)-buf.Len():]
		nameEnd := bytes.IndexByte(remaining, 0)
		if nameEnd == -1 {
			break
		}
		
		name := string(remaining[:nameEnd])
		buf.Seek(int64(nameEnd+1), 1)
		
		hashBytes := make([]byte, 20)
		buf.Read(hashBytes)
		hash := hex.EncodeToString(hashBytes)
		
		entries = append(entries, TreeEntry{
			Mode: uint32(mode),
			Name: name,
			Hash: hash,
		})
	}
	
	return entries, nil
}

func (e TreeEntry) IsDirectory() bool {
	return e.Mode == 040000
}