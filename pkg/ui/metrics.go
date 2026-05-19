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
	"runtime"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// SystemMetrics for dashboard header sampling.
type SystemMetrics struct {
	CPUPercent   float64
	MemUsedMB    float64
	MemPercent   float64
	NumGoroutine int
}

// SampleMetrics samples memory and CPU (100ms sample window).
func SampleMetrics() SystemMetrics {
	m := SystemMetrics{NumGoroutine: runtime.NumGoroutine()}
	if vm, err := mem.VirtualMemory(); err == nil {
		m.MemUsedMB = float64(vm.Used) / (1024 * 1024)
		m.MemPercent = vm.UsedPercent
	}
	if pct, err := cpu.Percent(100000000, false); err == nil && len(pct) > 0 {
		m.CPUPercent = pct[0]
	}
	return m
}
