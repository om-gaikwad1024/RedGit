package remote

import (
	"fmt"
)

type Remote interface {
	Fetch() error
	Push(branch string) error
	ListRefs() (map[string]string, error)
}

type RemoteConfig struct {
	Name string
	URL  string
}

func NewRemote(config RemoteConfig) (Remote, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("remote URL cannot be empty")
	}
	
	return &LocalRemote{
		Name: config.Name,
		URL:  config.URL,
	}, nil
}