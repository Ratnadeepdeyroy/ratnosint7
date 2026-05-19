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

package dedupe

import (
	"sync"
	"testing"
)

func TestDedupe(t *testing.T) {
	d := New()
	if !d.Add("example.com") {
		t.Error("first add should return true")
	}
	if d.Add("example.com") {
		t.Error("duplicate add should return false")
	}
	if !d.Add("sub.example.com") {
		t.Error("new domain add should return true")
	}
}

func TestDedupeConcurrent(t *testing.T) {
	d := New()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			domain := "example.com"
			if n%2 == 0 {
				domain = "sub.example.com"
			}
			d.Add(domain)
		}(i)
	}
	wg.Wait()
}

func BenchmarkDedupe(b *testing.B) {
	d := New()
	domains := []string{"example.com", "sub.example.com", "a.example.com", "b.example.com"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Add(domains[i%len(domains)])
	}
}
