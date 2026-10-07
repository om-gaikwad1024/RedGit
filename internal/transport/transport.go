package transport

import (
	"io"
)

type Transport interface {
	Connect(url string) error
	Send(data []byte) error
	Receive() ([]byte, error)
	Close() error
}

type FileTransport struct {
	reader io.Reader
	writer io.Writer
}

func NewFileTransport() *FileTransport {
	return &FileTransport{}
}

func (t *FileTransport) Connect(url string) error {
	return nil
}

func (t *FileTransport) Send(data []byte) error {
	if t.writer != nil {
		_, err := t.writer.Write(data)
		return err
	}
	return nil
}

func (t *FileTransport) Receive() ([]byte, error) {
	return nil, nil
}

func (t *FileTransport) Close() error {
	return nil
}