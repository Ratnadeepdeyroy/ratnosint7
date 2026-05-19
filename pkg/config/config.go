// Copyright 2026 Ratnadeep.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package config provides configuration loading for ratnosint7.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

// AppConfig holds the application configuration.
type AppConfig struct {
	RawBuffer     int           `yaml:"raw_buffer"`
	ParsedBuffer  int           `yaml:"parsed_buffer"`
	UniqueBuffer  int           `yaml:"unique_buffer"`
	ParserWorkers int           `yaml:"parser_workers"`
	Timeout       time.Duration `yaml:"timeout"`
	Tools         []ToolConfig  `yaml:"tools"`
	Resolvers     []string      `yaml:"resolvers"`
}

// DefaultConfig returns configuration with sensible defaults.
func DefaultConfig() AppConfig {
	cpu := runtime.NumCPU()
	return AppConfig{
		RawBuffer:     8192,
		ParsedBuffer:  4096,
		UniqueBuffer:  1024,
		ParserWorkers: cpu * 2,
		Timeout:       30 * time.Minute,
	}
}

// Load loads configuration from the given path (directory containing configs).
// Loads tools.yaml and resolvers.txt from the path. Optional config.yaml for overrides.
// Merges with defaults for any unset values.
func Load(path string) (*AppConfig, error) {
	cfg := DefaultConfig()

	if path == "" {
		return &cfg, nil
	}

	toolsPath := filepath.Join(path, "tools.yaml")
	tools, err := LoadTools(toolsPath)
	if err == nil {
		cfg.Tools = tools
	}

	resolversPath := filepath.Join(path, "resolvers.txt")
	resolvers, err := LoadResolvers(resolversPath)
	if err == nil && len(resolvers) > 0 {
		cfg.Resolvers = resolvers
	}

	configPath := filepath.Join(path, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if len(data) == 0 {
		return &cfg, nil
	}

	var fileCfg struct {
		RawBuffer     *int `yaml:"raw_buffer"`
		ParsedBuffer  *int `yaml:"parsed_buffer"`
		UniqueBuffer  *int `yaml:"unique_buffer"`
		ParserWorkers *int `yaml:"parser_workers"`
		Timeout       *int `yaml:"timeout_minutes"`
	}

	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if fileCfg.RawBuffer != nil {
		cfg.RawBuffer = *fileCfg.RawBuffer
	}
	if fileCfg.ParsedBuffer != nil {
		cfg.ParsedBuffer = *fileCfg.ParsedBuffer
	}
	if fileCfg.UniqueBuffer != nil {
		cfg.UniqueBuffer = *fileCfg.UniqueBuffer
	}
	if fileCfg.ParserWorkers != nil {
		cfg.ParserWorkers = *fileCfg.ParserWorkers
	}
	if fileCfg.Timeout != nil {
		cfg.Timeout = time.Duration(*fileCfg.Timeout) * time.Minute
	}

	return &cfg, nil
}

// LoadTools loads tools from tools.yaml at the given path.
func LoadTools(path string) ([]ToolConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read tools config: %w", err)
	}

	var fileCfg struct {
		Tools []ToolConfig `yaml:"tools"`
	}
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return nil, fmt.Errorf("parse tools config: %w", err)
	}

	return fileCfg.Tools, nil
}

// LoadResolvers loads resolver IPs from a file (one per line).
func LoadResolvers(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read resolvers: %w", err)
	}

	var resolvers []string
	lines := splitLines(string(data))
	for _, line := range lines {
		line = trimSpace(line)
		if line != "" && line[0] != '#' {
			resolvers = append(resolvers, line)
		}
	}
	return resolvers, nil
}

// ResolversPath returns the path to resolvers.txt relative to config dir.
func ResolversPath(configDir string) string {
	return filepath.Join(configDir, "resolvers.txt")
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
