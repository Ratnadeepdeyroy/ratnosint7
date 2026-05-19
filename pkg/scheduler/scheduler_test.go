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

package scheduler

import (
	"context"
	"testing"
)

func TestOptimalWorkers(t *testing.T) {
	w := OptimalWorkers()
	if w < 1 {
		t.Errorf("OptimalWorkers() = %d, want >= 1", w)
	}
}

func TestRunPluginWorkers(t *testing.T) {
	ctx := context.Background()
	jobs := make(chan Job, 2)
	raw := make(chan string, 10)

	mock := &mockPlugin{name: "mock"}
	jobs <- Job{Plugin: mock, Domain: "example.com"}
	jobs <- Job{Plugin: mock, Domain: "test.com"}
	close(jobs)

	RunPluginWorkers(ctx, jobs, raw, 2)

	count := 0
	for range raw {
		count++
	}
	if count != 4 {
		t.Errorf("expected 4 results, got %d", count)
	}
}

type mockPlugin struct {
	name string
}

func (m *mockPlugin) Name() string { return m.name }

func (m *mockPlugin) Run(ctx context.Context, domain string, output chan<- string) error {
	output <- domain
	output <- "sub." + domain
	return nil
}

func BenchmarkScheduler(b *testing.B) {
	ctx := context.Background()
	mock := &mockPlugin{name: "mock"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jobs := make(chan Job, 10)
		raw := make(chan string, 100)
		for j := 0; j < 10; j++ {
			jobs <- Job{Plugin: mock, Domain: "example.com"}
		}
		close(jobs)
		RunPluginWorkers(ctx, jobs, raw, 4)
		for range raw {
		}
	}
}
