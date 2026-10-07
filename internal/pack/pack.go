package pack

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
)

type PackFile struct {
	Version uint32
	Objects []PackedObject
}

type PackedObject struct {
	Type uint8
	Size uint64
	Data []byte
}

const (
	OBJ_COMMIT = 1
	OBJ_TREE   = 2
	OBJ_BLOB   = 3
	OBJ_TAG    = 4
)

func CreatePackFile(objects []PackedObject) ([]byte, error) {
	var buf bytes.Buffer
	
	buf.WriteString("PACK")
	binary.Write(&buf, binary.BigEndian, uint32(2))
	binary.Write(&buf, binary.BigEndian, uint32(len(objects)))
	
	for _, obj := range objects {
		if err := writePackedObject(&buf, obj); err != nil {
			return nil, err
		}
	}
	
	hash := sha1.Sum(buf.Bytes())
	buf.Write(hash[:])
	
	return buf.Bytes(), nil
}

func writePackedObject(buf *bytes.Buffer, obj PackedObject) error {
	writeVariableLengthSize(buf, obj.Type, obj.Size)
	
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	if _, err := w.Write(obj.Data); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	
	buf.Write(compressed.Bytes())
	return nil
}

func writeVariableLengthSize(buf *bytes.Buffer, objType uint8, size uint64) {
	b := byte((objType << 4) | (byte(size) & 0x0F))
	size >>= 4
	
	if size > 0 {
		b |= 0x80
	}
	buf.WriteByte(b)
	
	for size > 0 {
		b = byte(size & 0x7F)
		size >>= 7
		if size > 0 {
			b |= 0x80
		}
		buf.WriteByte(b)
	}
}

func UnpackFile(data []byte) (*PackFile, error) {
	buf := bytes.NewReader(data)
	
	sig := make([]byte, 4)
	if _, err := buf.Read(sig); err != nil {
		return nil, err
	}
	if string(sig) != "PACK" {
		return nil, fmt.Errorf("invalid pack signature")
	}
	
	var version, count uint32
	binary.Read(buf, binary.BigEndian, &version)
	binary.Read(buf, binary.BigEndian, &count)
	
	pack := &PackFile{
		Version: version,
		Objects: make([]PackedObject, 0, count),
	}
	
	for i := uint32(0); i < count; i++ {
		obj, err := readPackedObject(buf)
		if err != nil {
			return nil, err
		}
		pack.Objects = append(pack.Objects, obj)
	}
	
	return pack, nil
}

func readPackedObject(buf io.Reader) (PackedObject, error) {
	objType, size, err := readVariableLengthSize(buf)
	if err != nil {
		return PackedObject{}, err
	}
	
	r, err := zlib.NewReader(buf)
	if err != nil {
		return PackedObject{}, err
	}
	defer r.Close()
	
	data, err := io.ReadAll(r)
	if err != nil {
		return PackedObject{}, err
	}
	
	return PackedObject{
		Type: objType,
		Size: size,
		Data: data,
	}, nil
}

func readVariableLengthSize(buf io.Reader) (uint8, uint64, error) {
	var b [1]byte
	if _, err := buf.Read(b[:]); err != nil {
		return 0, 0, err
	}
	
	objType := (b[0] >> 4) & 0x07
	size := uint64(b[0] & 0x0F)
	shift := uint64(4)
	
	for b[0]&0x80 != 0 {
		if _, err := buf.Read(b[:]); err != nil {
			return 0, 0, err
		}
		size |= uint64(b[0]&0x7F) << shift
		shift += 7
	}
	
	return objType, size, nil
}

func TypeToPackType(objType string) uint8 {
	switch objType {
	case "commit":
		return OBJ_COMMIT
	case "tree":
		return OBJ_TREE
	case "blob":
		return OBJ_BLOB
	case "tag":
		return OBJ_TAG
	default:
		return OBJ_BLOB
	}
}

func PackTypeToType(packType uint8) string {
	switch packType {
	case OBJ_COMMIT:
		return "commit"
	case OBJ_TREE:
		return "tree"
	case OBJ_BLOB:
		return "blob"
	case OBJ_TAG:
		return "tag"
	default:
		return "blob"
	}
}