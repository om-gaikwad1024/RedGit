package core

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	sections map[string]map[string]string
}

func NewConfig() *Config {
	return &Config{
		sections: make(map[string]map[string]string),
	}
}

func LoadConfig(repo *Repository) (*Config, error) {
	configPath := filepath.Join(repo.GitDir, "config")
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewConfig(), nil
		}
		return nil, err
	}
	defer file.Close()
	
	config := NewConfig()
	var currentSection string
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			if config.sections[currentSection] == nil {
				config.sections[currentSection] = make(map[string]string)
			}
		} else if strings.Contains(line, "=") && currentSection != "" {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			config.sections[currentSection][key] = value
		}
	}
	
	return config, scanner.Err()
}

func (c *Config) Get(section, key string) (string, bool) {
	if sec, ok := c.sections[section]; ok {
		val, exists := sec[key]
		return val, exists
	}
	return "", false
}

func (c *Config) Set(section, key, value string) {
	if c.sections[section] == nil {
		c.sections[section] = make(map[string]string)
	}
	c.sections[section][key] = value
}

func (c *Config) Save(repo *Repository) error {
	configPath := filepath.Join(repo.GitDir, "config")
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	for section, kvs := range c.sections {
		fmt.Fprintf(file, "[%s]\n", section)
		for key, value := range kvs {
			fmt.Fprintf(file, "\t%s = %s\n", key, value)
		}
		fmt.Fprintln(file)
	}
	
	return nil
}

func (c *Config) GetSections() map[string]map[string]string {
	return c.sections
}