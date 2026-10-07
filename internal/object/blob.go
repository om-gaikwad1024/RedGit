package object

func NewBlob(content []byte) *Object {
	return &Object{
		Type:    BlobType,
		Content: content,
	}
}