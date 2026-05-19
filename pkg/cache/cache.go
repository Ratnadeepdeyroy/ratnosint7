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

// Package cache provides result caching with config-aware keys.
package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	cacheDir = ".ratnosint7/cache"
	ttl      = 24 * time.Hour
)

type ScanConfig struct {
	PluginNames []string
	Passive     bool
	Active      bool
	Both        bool
	Resolvers   string
}

func Key(domain string, cfg ScanConfig) string {
	plugins := make([]string, len(cfg.PluginNames))
	copy(plugins, cfg.PluginNames)
	sort.Strings(plugins)
	pluginPart := strings.Join(plugins, ",")
	flags := fmt.Sprintf("passive=%v,active=%v,both=%v,resolvers=%s", cfg.Passive, cfg.Active, cfg.Both, cfg.Resolvers)
	data := domain + "|" + pluginPart + "|" + flags
	h := sha1.Sum([]byte(data))
	return hex.EncodeToString(h[:])
}

func Path(key string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	return filepath.Join(home, cacheDir, key+".txt"), nil
}

// Get returns cached content if valid plus entry age since write.
func Get(domain, key string) (rc io.ReadCloser, ok bool, age time.Duration, err error) {
	p, err := Path(key)
	if err != nil {
		return nil, false, 0, err
	}
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, 0, nil
		}
		return nil, false, 0, err
	}
	fi, err := f.Stat()
	if err != nil || time.Since(fi.ModTime()) > ttl {
		f.Close()
		return nil, false, 0, nil
	}
	return f, true, time.Since(fi.ModTime()), nil
}

func Invalidate(key string) error {
	p, err := Path(key)
	if err != nil {
		return err
	}
	return os.Remove(p)
}

func Set(domain, key string, r io.Reader) error {
	p, err := Path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	tmp := p + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, r); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, p)
}
