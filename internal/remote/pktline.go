package remote

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func WritePktLine(w io.Writer, line string) error {
	if line == "" {
		_, err := w.Write([]byte("0000"))
		return err
	}
	
	length := len(line) + 4
	pktLine := fmt.Sprintf("%04x%s", length, line)
	_, err := w.Write([]byte(pktLine))
	return err
}

func ReadPktLine(r *bufio.Reader) (string, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return "", err
	}
	
	length, err := strconv.ParseInt(string(lenBuf), 16, 64)
	if err != nil {
		return "", err
	}
	
	if length == 0 {
		return "", nil
	}
	
	if length < 4 {
		return "", fmt.Errorf("invalid pkt-line length: %d", length)
	}
	
	data := make([]byte, length-4)
	if _, err := io.ReadFull(r, data); err != nil {
		return "", err
	}
	
	return string(data), nil
}

func ParseRefs(r io.Reader) (map[string]string, error) {
	refs := make(map[string]string)
	scanner := bufio.NewScanner(r)
	
	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		
		if len(line) < 4 {
			continue
		}
		
		pktLen := line[:4]
		if pktLen == "0000" {
			continue
		}
		
		content := line[4:]
		
		if strings.HasPrefix(content, "# service=") {
			continue
		}
		
		parts := strings.Fields(content)
		if len(parts) >= 2 {
			commit := parts[0]
			refName := parts[1]
			
			if strings.HasPrefix(refName, "refs/heads/") {
				branch := strings.TrimPrefix(refName, "refs/heads/")
				refs[branch] = commit
			}
		}
	}
	
	return refs, scanner.Err()
}