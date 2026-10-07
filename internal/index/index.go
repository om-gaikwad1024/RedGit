package index

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"github.com/om-gaikwad1024/redgit/internal/core"
)

type Index struct {
	Entries []Entry
}

func NewIndex() *Index {
	return &Index{
		Entries: []Entry{},
	}
}

func ReadIndex(gitDir string) (*Index, error) {
	indexPath := filepath.Join(gitDir, "index")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewIndex(), nil
		}
		return nil, err
	}
	
	if len(data) < 12 {
		return NewIndex(), nil
	}
	
	signature := string(data[:4])
	if signature != "DIRC" {
		return NewIndex(), nil
	}
	
	version := binary.BigEndian.Uint32(data[4:8])
	if version != 2 {
		return nil, fmt.Errorf("unsupported index version: %d", version)
	}
	
	entryCount := binary.BigEndian.Uint32(data[8:12])
	
	index := NewIndex()
	offset := 12
	
	for i := uint32(0); i < entryCount && offset < len(data)-20; i++ {
		if offset+62 > len(data) {
			break
		}
		
		entry := Entry{}
		
		ctimeSec := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 8
		mtimeSec := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 8
		
		entry.Dev = binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
		entry.Ino = binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
		entry.Mode = binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
		entry.Uid = binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
		entry.Gid = binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
		entry.Size = binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
		
		hashBytes := data[offset : offset+20]
		entry.Hash = fmt.Sprintf("%x", hashBytes)
		offset += 20
		
		entry.Flags = binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2
		
		nameLen := entry.Flags & 0xFFF
		if offset+int(nameLen) > len(data) {
			break
		}
		entry.Path = string(data[offset : offset+int(nameLen)])
		offset += int(nameLen)
		
		padLen := (8 - ((62 + int(nameLen)) % 8)) % 8
		offset += padLen
		
		entry.Ctime = entry.Ctime.Add(0)
		entry.Mtime = entry.Mtime.Add(0)
		_, _ = ctimeSec, mtimeSec
		
		index.Entries = append(index.Entries, entry)
	}
	
	return index, nil
}

func (idx *Index) WriteIndex(gitDir string) error {
	var buf bytes.Buffer
	
	buf.WriteString("DIRC")
	binary.Write(&buf, binary.BigEndian, uint32(2))
	binary.Write(&buf, binary.BigEndian, uint32(len(idx.Entries)))
	
	for _, entry := range idx.Entries {
		ctimeSec := uint32(entry.Ctime.Unix())
		ctimeNsec := uint32(0)
		mtimeSec := uint32(entry.Mtime.Unix())
		mtimeNsec := uint32(0)
		
		binary.Write(&buf, binary.BigEndian, ctimeSec)
		binary.Write(&buf, binary.BigEndian, ctimeNsec)
		binary.Write(&buf, binary.BigEndian, mtimeSec)
		binary.Write(&buf, binary.BigEndian, mtimeNsec)
		binary.Write(&buf, binary.BigEndian, entry.Dev)
		binary.Write(&buf, binary.BigEndian, entry.Ino)
		binary.Write(&buf, binary.BigEndian, entry.Mode)
		binary.Write(&buf, binary.BigEndian, entry.Uid)
		binary.Write(&buf, binary.BigEndian, entry.Gid)
		binary.Write(&buf, binary.BigEndian, entry.Size)
		
		hashBytes := make([]byte, 20)
		for i := 0; i < 20 && i*2 < len(entry.Hash); i++ {
			fmt.Sscanf(entry.Hash[i*2:i*2+2], "%02x", &hashBytes[i])
		}
		buf.Write(hashBytes)
		
		nameLen := len(entry.Path)
		if nameLen > 0xFFF {
			nameLen = 0xFFF
		}
		flags := uint16(nameLen)
		binary.Write(&buf, binary.BigEndian, flags)
		
		buf.WriteString(entry.Path)
		
		entryLen := 62 + len(entry.Path)
		padLen := (8 - (entryLen % 8)) % 8
		buf.Write(make([]byte, padLen))
	}
	
	checksum := sha1.Sum(buf.Bytes())
	buf.Write(checksum[:])
	
	indexPath := filepath.Join(gitDir, "index")
	return core.AtomicWrite(indexPath, buf.Bytes(), 0644)
}

func (idx *Index) AddEntry(entry Entry) {
	for i, e := range idx.Entries {
		if e.Path == entry.Path {
			idx.Entries[i] = entry
			return
		}
	}
	idx.Entries = append(idx.Entries, entry)
	sort.Slice(idx.Entries, func(i, j int) bool {
		return idx.Entries[i].Path < idx.Entries[j].Path
	})
}

func (idx *Index) GetEntry(path string) (*Entry, bool) {
	for _, entry := range idx.Entries {
		if entry.Path == path {
			return &entry, true
		}
	}
	return nil, false
}