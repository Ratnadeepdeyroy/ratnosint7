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

// Package dedupe provides sharded map deduplication for domains.
package dedupe

import (
	"sync"
)

const numShards = 32

type shard struct {
	mu sync.Mutex
	m  map[string]struct{}
}

// Dedupe is a sharded map for write-heavy deduplication.
type Dedupe struct {
	shards [numShards]shard
}

// New creates a new Dedupe.
func New() *Dedupe {
	d := &Dedupe{}
	for i := range d.shards {
		d.shards[i].m = make(map[string]struct{})
	}
	return d
}

// shardIndex computes FNV-1a inline — no allocation per call.
func (d *Dedupe) shardIndex(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h % numShards
}

// Add adds a domain. Returns true if newly seen, false if duplicate.
func (d *Dedupe) Add(domain string) bool {
	idx := d.shardIndex(domain)
	sh := &d.shards[idx]
	sh.mu.Lock()
	_, exists := sh.m[domain]
	if !exists {
		sh.m[domain] = struct{}{}
	}
	sh.mu.Unlock()
	return !exists
}
