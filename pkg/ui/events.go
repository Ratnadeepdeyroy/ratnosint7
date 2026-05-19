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

package ui

import (
	"time"

	apperrors "github.com/Ratnadeepdeyroy/ratnosint7/pkg/errors"
)

// ScanStatus indicates high-level dashboard state.
type ScanStatus string

const (
	StatusStarting ScanStatus = "STARTING"
	StatusRunning  ScanStatus = "RUNNING"
	StatusDone     ScanStatus = "DONE"
	StatusCached   ScanStatus = "CACHED"
)

// Event is emitted on the UI bus during scan/update.
type Event interface {
	isUIEvent()
}

// ScanStarted is published when pipeline starts (fresh scan, not cached).
type ScanStarted struct {
	Domain      string
	ScanID      int64
	PluginCount int
}

func (ScanStarted) isUIEvent() {}

// PluginStarted is published before a plugin's Run invokes.
type PluginStarted struct {
	Name string
}

func (PluginStarted) isUIEvent() {}

// PluginFinished is published after a plugin Run returns.
type PluginFinished struct {
	Name         string
	DomainsFound int
	Duration     time.Duration
	Err          error
}

func (PluginFinished) isUIEvent() {}

// PluginSkipped is published when install/preflight omits a tool.
type PluginSkipped struct {
	Name string
	Err  *apperrors.PluginError
}

func (PluginSkipped) isUIEvent() {}

// CacheHit signals a valid cache restore.
type CacheHit struct {
	Domain       string
	Age          time.Duration
	TotalDomains int
}

func (CacheHit) isUIEvent() {}

// UniqueProgress reports deduplicated domains written so far.
type UniqueProgress struct {
	Count      int
	CPUPercent float64
	MemMB      float64
}

func (UniqueProgress) isUIEvent() {}

// ScanComplete is published immediately before Scan returns successfully.
type ScanComplete struct {
	Domain             string
	OutputPath         string
	TotalDomains       int
	Duration           time.Duration
	FromCache          bool
	CacheAge           time.Duration
	PluginStats        map[string]PluginStatInfo
	Errors             []*apperrors.PluginError
	PluginsFailedCount int
}

func (ScanComplete) isUIEvent() {}

// PluginStatInfo carries per-plugin totals for UI.
type PluginStatInfo struct {
	DomainsFound int
	Duration     time.Duration
	Err          error
}

// ToolInstallStarted signals update-tools begun for one binary.
type ToolInstallStarted struct {
	Name string
}

func (ToolInstallStarted) isUIEvent() {}

// ToolInstallFinished signals update-tools result for one tool.
type ToolInstallFinished struct {
	Name string
	Err  error
}

func (ToolInstallFinished) isUIEvent() {}

// UpdateToolsComplete is published when update-tools finishes.
type UpdateToolsComplete struct {
	SuccessCount int
	FailCount    int
}

func (UpdateToolsComplete) isUIEvent() {}
