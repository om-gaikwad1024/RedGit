package index

import (
	"os"
	"time"
)

type Entry struct {
	Ctime    time.Time
	Mtime    time.Time
	Dev      uint32
	Ino      uint32
	Mode     uint32
	Uid      uint32
	Gid      uint32
	Size     uint32
	Hash     string
	Flags    uint16
	Path     string
}

func NewEntry(path string, hash string, info os.FileInfo) Entry {
	return Entry{
		Ctime: info.ModTime(),
		Mtime: info.ModTime(),
		Mode:  uint32(info.Mode()),
		Size:  uint32(info.Size()),
		Hash:  hash,
		Path:  path,
	}
}