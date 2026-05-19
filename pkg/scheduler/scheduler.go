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

// Package scheduler manages the plugin worker pool and job dispatch.
package scheduler

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
)

// Job represents a plugin run for a domain.
type Job struct {
	Plugin interface {
		Name() string
		Run(ctx context.Context, domain string, output chan<- string) error
	}
	Domain string
}

// OptimalWorkers returns the number of workers. Call once at startup.
// Uses runtime.NumCPU; ReadMemStats triggers GC stop-the-world.
func OptimalWorkers() int {
	cpu := runtime.NumCPU()
	if cpu < 2 {
		return 2
	}
	return cpu
}

// RunPluginWorkers starts workers that pull jobs and run plugins.
// Workers write to raw; no plugin closes raw. Caller must close jobs when done.
// When all workers exit, raw is closed.
func RunPluginWorkers(ctx context.Context, jobs <-chan Job, raw chan<- string, workers int) {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					slog.Error("plugin worker panic", "err", r)
					// Drain jobs so the sender goroutine (close(jobs)) can exit
					// and close(raw) can proceed.
					for range jobs {
					}
				}
			}()
			for job := range jobs {
				if err := job.Plugin.Run(ctx, job.Domain, raw); err != nil {
					slog.Warn("plugin run failed", "plugin", job.Plugin.Name(), "domain", job.Domain, "err", err)
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(raw)
	}()
}
