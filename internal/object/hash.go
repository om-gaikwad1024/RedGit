package object

import (
	"crypto/sha1"
	"encoding/hex"
)

func Hash(data []byte) string {
	h := sha1.Sum(data)
	return hex.EncodeToString(h[:])
}

func ComputeHash(objType string, content []byte) string {
	header := []byte(objType + " " + string(len(content)) + "\x00")
	data := append(header, content...)
	return Hash(data)
}