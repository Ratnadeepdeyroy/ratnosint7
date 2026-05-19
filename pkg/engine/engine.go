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

// Package engine orchestrates the scan pipeline.
package engine

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/cache"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/config"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/dedupe"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/errors"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/parser"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/scheduler"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/storage"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/ui"
)

const relayBufSize = 16384

func dbg(enabled bool, msg string, attrs ...any) {
	if enabled {
		slog.Debug(msg, attrs...)
	}
}

type Plugin interface {
	Name() string
	Run(ctx context.Context, domain string, output chan<- string) error
}

type ScanResult struct {
	OutputPath   string
	TotalDomains int
	PluginStats  map[string]PluginStat
	Errors       []*errors.PluginError
	Duration     time.Duration
	FromCache    bool
	CacheAge     time.Duration
}

type PluginStat struct {
	DomainsFound int
	Duration     time.Duration
	Err          error
}

type ScanConfig struct {
	Config       *config.AppConfig
	Plugins      []Plugin
	PluginNames  []string
	Passive      bool
	Active       bool
	Both         bool
	Resolvers    string
	OutputFormat string
	Overwrite    bool
	NoCache      bool
	Events       *ui.Bus
	DebugLogs    bool
}

func runForwardedPlugin(ctx context.Context, domain string, p Plugin, downstream chan<- string,
	events *ui.Bus, dbgEnabled bool) PluginStat {

	name := p.Name()
	if events != nil {
		events.Publish(ui.PluginStarted{Name: name})
	}
	relay := make(chan string, relayBufSize)
	done := make(chan error, 1)
	go func() { done <- p.Run(ctx, domain, relay) }()

	stat := PluginStat{}
	start := time.Now()
	completed := false
	for !completed {
		select {
		case err := <-done:
			completed = true
			stat.Err = err
		case <-ctx.Done():
			completed = true
			stat.Err = ctx.Err()
		case line := <-relay:
			stat.DomainsFound++
			select {
			case downstream <- line:
			case <-ctx.Done():
				completed = true
				stat.Err = ctx.Err()
			}
		}
	}
	for {
		select {
		case line := <-relay:
			stat.DomainsFound++
			select {
			case downstream <- line:
			case <-ctx.Done():
				goto emit
			}
		default:
			goto emit
		}
	}
emit:
	stat.Duration = time.Since(start)
	if dbgEnabled && stat.Err != nil {
		slog.Debug("plugin finished with error", "plugin", name, "err", stat.Err)
	}
	if events != nil {
		events.Publish(ui.PluginFinished{
			Name: name, DomainsFound: stat.DomainsFound,
			Duration: stat.Duration, Err: stat.Err,
		})
	}
	return stat
}

// statPlugin wraps a Plugin to capture per-plugin stats and publish events via the scheduler.
type statPlugin struct {
	inner  Plugin
	result *ScanResult
	mu     *sync.Mutex
	events *ui.Bus
	debug  bool
}

func (sp statPlugin) Name() string { return sp.inner.Name() }
func (sp statPlugin) Run(ctx context.Context, domain string, output chan<- string) error {
	st := runForwardedPlugin(ctx, domain, sp.inner, output, sp.events, sp.debug)
	sp.mu.Lock()
	sp.result.PluginStats[sp.inner.Name()] = st
	sp.mu.Unlock()
	if sp.debug && st.Err != nil {
		slog.Warn("plugin run failed", "plugin", sp.inner.Name(), "err", st.Err)
	}
	return st.Err
}

func statsToUI(m map[string]PluginStat) map[string]ui.PluginStatInfo {
	if len(m) == 0 {
		return nil
	}
	o := make(map[string]ui.PluginStatInfo, len(m))
	for k, v := range m {
		o[k] = ui.PluginStatInfo{DomainsFound: v.DomainsFound, Duration: v.Duration, Err: v.Err}
	}
	return o
}

func failedCount(m map[string]PluginStat) int {
	n := 0
	for _, s := range m {
		if s.Err != nil {
			n++
		}
	}
	return n
}

func Scan(ctx context.Context, domain string, cfg ScanConfig) (*ScanResult, error) {
	start := time.Now()
	result := &ScanResult{PluginStats: make(map[string]PluginStat)}
	scanCfg := cache.ScanConfig{
		PluginNames: cfg.PluginNames, Passive: cfg.Passive,
		Active: cfg.Active, Both: cfg.Both, Resolvers: cfg.Resolvers,
	}
	key := cache.Key(domain, scanCfg)

	if !cfg.NoCache {
		if rc, ok, age, err := cache.Get(domain, key); err == nil && ok {
			var lines []string
			sc := bufio.NewScanner(rc)
			for sc.Scan() {
				line := strings.TrimSpace(sc.Text())
				if line != "" {
					lines = append(lines, line)
				}
			}
			rc.Close()
			if len(lines) > 0 {
				outputPath := storage.OutputPath(domain, cfg.Overwrite, cfg.OutputFormat)
				tw, err := storage.NewWriter(outputPath, cfg.OutputFormat)
				if err != nil {
					return nil, err
				}
				for _, l := range lines {
					if err := tw.Write(l); err != nil {
						tw.Close()
						return nil, err
					}
				}
				if err := tw.Close(); err != nil {
					return nil, err
				}
				result.OutputPath = outputPath
				result.TotalDomains = len(lines)
				result.FromCache = true
				result.CacheAge = age
				result.Duration = time.Since(start)

				if cfg.Events != nil {
					cfg.Events.Publish(ui.CacheHit{Domain: domain, Age: age, TotalDomains: len(lines)})
					cfg.Events.Publish(ui.ScanComplete{
						Domain: domain, OutputPath: outputPath, TotalDomains: len(lines),
						Duration: result.Duration, FromCache: true, CacheAge: age,
						PluginsFailedCount: 0,
					})
				}
				dbg(cfg.DebugLogs, "cache hit", "domain", domain)
				return result, nil
			}
			_ = cache.Invalidate(key)
		}
	}

	outputPath := storage.OutputPath(domain, cfg.Overwrite, cfg.OutputFormat)
	raw := make(chan string, cfg.Config.RawBuffer)
	parsed := make(chan string, cfg.Config.ParsedBuffer)
	uniq := make(chan string, cfg.Config.UniqueBuffer)

	var uniqProg atomic.Int64
	var progCancel context.CancelFunc = func() {}
	if cfg.Events != nil {
		pctx, cancel := context.WithCancel(context.Background())
		progCancel = cancel
		go func() {
			t := time.NewTicker(280 * time.Millisecond)
			defer t.Stop()
			prev := int64(-1)
			for {
				select {
				case <-pctx.Done():
					return
				case <-t.C:
					v := uniqProg.Load()
					if v != prev {
						prev = v
						met := ui.SampleMetrics()
						cfg.Events.Publish(ui.UniqueProgress{Count: int(v), CPUPercent: met.CPUPercent, MemMB: met.MemUsedMB})
					}
				}
			}
		}()
	}
	defer progCancel()

	if cfg.Events != nil {
		cfg.Events.Publish(ui.ScanStarted{
			Domain: domain, ScanID: time.Now().Unix(), PluginCount: len(cfg.Plugins),
		})
	}

	var statMu sync.Mutex
	jobs := make(chan scheduler.Job, len(cfg.Plugins))
	for _, pl := range cfg.Plugins {
		jobs <- scheduler.Job{
			Plugin: statPlugin{inner: pl, result: result, mu: &statMu, events: cfg.Events, debug: cfg.DebugLogs},
			Domain: domain,
		}
	}
	close(jobs)
	scheduler.RunPluginWorkers(ctx, jobs, raw, scheduler.OptimalWorkers())

	parserWorkers := cfg.Config.ParserWorkers
	if parserWorkers < 1 {
		parserWorkers = scheduler.OptimalWorkers() * 2
	}
	var parseWg sync.WaitGroup
	for i := 0; i < parserWorkers; i++ {
		parseWg.Add(1)
		go func() {
			defer parseWg.Done()
			for d := range raw {
				if norm, ok := parser.NormalizeAndValidate(d); ok {
					select {
					case parsed <- norm:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}
	go func() {
		parseWg.Wait()
		close(parsed)
	}()

	ded := dedupe.New()
	go func() {
		for d := range parsed {
			if ded.Add(d) {
				select {
				case uniq <- d:
				case <-ctx.Done():
					for range parsed {
					}
					close(uniq)
					return
				}
			}
		}
		close(uniq)
	}()

	writer, err := storage.NewWriter(outputPath, cfg.OutputFormat)
	if err != nil {
		return nil, err
	}
	count := 0
	for d := range uniq {
		if err := writer.Write(d); err != nil {
			result.Errors = append(result.Errors, errors.NewPluginError("writer", errors.PartialOutput, err))
			break
		}
		count++
		uniqProg.Store(int64(count))
	}
	if err := writer.Close(); err != nil {
		dbg(cfg.DebugLogs, "writer close failed", "err", err)
	}

	result.OutputPath = outputPath
	result.TotalDomains = count
	result.Duration = time.Since(start)

	if count > 0 && !cfg.NoCache {
		if f, err := os.Open(outputPath); err == nil {
			if _, err := f.Seek(0, io.SeekStart); err == nil {
				_ = cache.Set(domain, key, f)
			}
			f.Close()
		}
	}
	failN := failedCount(result.PluginStats)
	if cfg.Events != nil {
		cfg.Events.Publish(ui.ScanComplete{
			Domain: domain, OutputPath: outputPath, TotalDomains: count,
			Duration: result.Duration, FromCache: false,
			PluginStats: statsToUI(result.PluginStats), Errors: result.Errors,
			PluginsFailedCount: failN,
		})
	}
	dbg(cfg.DebugLogs, "scan complete", "domain", domain, "unique", count)
	return result, nil
}
