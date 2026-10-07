package remote

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	
	"github.com/om-gaikwad1024/redgit/internal/object"
	"github.com/om-gaikwad1024/redgit/internal/pack"
)

type HTTPRemote struct {
	URL    string
	Client *http.Client
	Auth   *Credentials
}

type Credentials struct {
	Username string
	Token    string
}

func NewHTTPRemote(url string, auth *Credentials) *HTTPRemote {
	return &HTTPRemote{
		URL: url,
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
			Timeout: 30 * time.Second,
		},
		Auth: auth,
	}
}

func (r *HTTPRemote) DiscoverRefs(service string) (map[string]string, error) {
	url := fmt.Sprintf("%s/info/refs?service=%s", r.URL, service)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	if r.Auth != nil {
		req.SetBasicAuth(r.Auth.Username, r.Auth.Token)
	}
	
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	
	return ParseRefs(resp.Body)
}

func (r *HTTPRemote) FetchPack(wants, haves []string) (*pack.PackFile, error) {
	url := fmt.Sprintf("%s/git-upload-pack", r.URL)
	
	var body bytes.Buffer
	
	for _, want := range wants {
		WritePktLine(&body, fmt.Sprintf("want %s\n", want))
	}
	WritePktLine(&body, "")
	
	for _, have := range haves {
		WritePktLine(&body, fmt.Sprintf("have %s\n", have))
	}
	WritePktLine(&body, "done\n")
	
	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/x-git-upload-pack-request")
	if r.Auth != nil {
		req.SetBasicAuth(r.Auth.Username, r.Auth.Token)
	}
	
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	
	packData, err := skipPktLineHeaders(resp.Body)
	if err != nil {
		return nil, err
	}
	
	return pack.UnpackFile(packData)
}

func (r *HTTPRemote) PushPack(branch string, oldCommit, newCommit string, packData []byte) error {
	url := fmt.Sprintf("%s/git-receive-pack", r.URL)
	
	var body bytes.Buffer
	
	if oldCommit == "" {
		oldCommit = strings.Repeat("0", 40)
	}
	
	cmd := fmt.Sprintf("%s %s refs/heads/%s\x00report-status\n", oldCommit, newCommit, branch)
	WritePktLine(&body, cmd)
	WritePktLine(&body, "")
	
	body.Write(packData)
	
	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/x-git-receive-pack-request")
	if r.Auth != nil {
		req.SetBasicAuth(r.Auth.Username, r.Auth.Token)
	}
	
	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	
	return parseReceivePackResponse(resp.Body)
}

func skipPktLineHeaders(r io.Reader) ([]byte, error) {
	scanner := bufio.NewScanner(r)
	
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) >= 4 && line[:4] == "0000" {
			break
		}
	}
	
	remaining, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	
	if len(remaining) > 8 && string(remaining[1:5]) == "PACK" {
		return remaining[1:], nil
	}
	
	packStart := bytes.Index(remaining, []byte("PACK"))
	if packStart == -1 {
		return nil, fmt.Errorf("no pack data found")
	}
	
	return remaining[packStart:], nil
}

func parseReceivePackResponse(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	
	for scanner.Scan() {
		line := scanner.Text()
		
		if len(line) < 4 {
			continue
		}
		
		content := line[4:]
		
		if strings.Contains(content, "unpack ok") {
			return nil
		}
		
		if strings.HasPrefix(content, "ng ") {
			return fmt.Errorf("push rejected: %s", content[3:])
		}
	}
	
	return scanner.Err()
}

func (r *HTTPRemote) Fetch(localGit string, wants, haves []string) error {
	packFile, err := r.FetchPack(wants, haves)
	if err != nil {
		return err
	}
	
	for _, obj := range packFile.Objects {
		objType := pack.PackTypeToType(obj.Type)
		newObj := &object.Object{
			Type:    object.ObjectType(objType),
			Content: obj.Data,
		}
		object.WriteObject(localGit, newObj)
	}
	
	return nil
}