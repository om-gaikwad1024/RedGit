package porcelain

import "github.com/om-gaikwad1024/redgit/internal/object"

type fetchObjectInfo struct {
	Hash    string
	Type    object.ObjectType
	Content []byte
}