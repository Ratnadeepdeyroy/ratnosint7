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
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"

	appui "github.com/Ratnadeepdeyroy/ratnosint7/pkg/ui"
)

type pluginRow struct {
	name    string
	running bool
	done    bool
	domains int
	dur     time.Duration
	runErr  error
}

type dashModel struct {
	busCh          <-chan appui.Event
	styles         appui.Styles
	color          bool
	domain         string
	status         appui.ScanStatus
	scanID         int64
	pluginPlan     int
	pluginsOrdered []string
	pluginRows     map[string]pluginRow
	uniqSeen       int
	spinner        spinner.Model
	scanDone       *appui.ScanComplete
	waitingExit    bool
	width, height  int
	lastMetrics    appui.SystemMetrics
}

func NewDashModel(domain string, busCh <-chan appui.Event, theme appui.Theme, color bool) *dashModel {
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	if color {
		spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#EA8C55"))
	}
	return &dashModel{
		busCh: busCh, color: color, domain: domain,
		styles: appui.NewStyles(theme, color), pluginRows: make(map[string]pluginRow),
		status: appui.StatusStarting, spinner: spin, width: 80, height: 24,
	}
}
