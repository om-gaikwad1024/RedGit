package core

import (
	"io"
	"os"
	"path/filepath"
)

func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	
	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}()
	
	if _, err := tmpFile.Write(data); err != nil {
		return err
	}
	
	if err := tmpFile.Sync(); err != nil {
		return err
	}
	
	if err := tmpFile.Close(); err != nil {
		return err
	}
	
	if err := os.Chmod(tmpPath, perm); err != nil {
		return err
	}
	
	return os.Rename(tmpPath, path)
}

func AtomicCopy(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	
	dir := filepath.Dir(dst)
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	
	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}()
	
	if _, err := io.Copy(tmpFile, srcFile); err != nil {
		return err
	}
	
	if err := tmpFile.Sync(); err != nil {
		return err
	}
	
	if err := tmpFile.Close(); err != nil {
		return err
	}
	
	if err := os.Chmod(tmpPath, srcInfo.Mode()); err != nil {
		return err
	}
	
	return os.Rename(tmpPath, dst)
}