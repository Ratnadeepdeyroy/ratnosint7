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

package tui

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	appui "github.com/Ratnadeepdeyroy/ratnosint7/pkg/ui"
)

// RunScanDashboard fullscreen UI until ScanComplete or ctx cancelled.
func RunScanDashboard(ctx context.Context, domain string, bus *appui.Bus, theme appui.Theme, color bool) error {
	if bus == nil {
		return errors.New("nil bus")
	}
	sub := bus.Subscribe()
	m := NewDashModel(domain, sub, theme, color)
	opts := []tea.ProgramOption{}
	if ctx != nil {
		opts = append(opts, tea.WithContext(ctx))
	}
	return tea.NewProgram(m, opts...).Start()
}
