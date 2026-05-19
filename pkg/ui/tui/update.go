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
	tea "github.com/charmbracelet/bubbletea"

	appui "github.com/Ratnadeepdeyroy/ratnosint7/pkg/ui"
)

type busEnvelope struct {
	Event appui.Event
}

type metricTickMsg struct{}

func listenCmd(ch <-chan appui.Event) tea.Cmd {
	return func() tea.Msg { return busEnvelope{Event: <-ch} }
}

func metricTickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(time.Time) tea.Msg {
		return metricTickMsg{}
	})
}

func (m *dashModel) Init() tea.Cmd {
	return tea.Batch(listenCmd(m.busCh), m.spinner.Tick, metricTickCmd())
}

func (m *dashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var spin tea.Cmd
	m.spinner, spin = m.spinner.Update(msg)
	switch msg.(type) {
	case tea.WindowSizeMsg:
		w := msg.(tea.WindowSizeMsg)
		m.width, m.height = w.Width, w.Height
		return m.follow(spin)
	case tea.KeyMsg:
		if m.waitingExit {
			return m, tea.Quit
		}
		if msg.(tea.KeyMsg).String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m.follow(spin)
	case busEnvelope:
		m.apply(msg.(busEnvelope).Event)
		if m.scanDone != nil {
			m.waitingExit = true
			return m, nil
		}
		return m, tea.Batch(listenCmd(m.busCh), spin, metricTickCmd())
	case metricTickMsg:
		m.lastMetrics = appui.SampleMetrics()
		return m, metricTickCmd()
	case spinner.TickMsg:
		return m.follow(spin)
	default:
		return m.follow(spin)
	}
}

func (m *dashModel) follow(spin tea.Cmd) (tea.Model, tea.Cmd) {
	if m.scanDone != nil && spin != nil {
		return m, spin
	}
	return m, spin
}

func (m *dashModel) apply(ev appui.Event) {
	switch e := ev.(type) {
	case appui.ScanStarted:
		m.domain, m.scanID, m.pluginPlan = e.Domain, e.ScanID, e.PluginCount
		m.status = appui.StatusRunning
	case appui.PluginSkipped:
		m.pluginRows[e.Name] = pluginRow{name: e.Name, done: true, runErr: e.Err}
		addOrder(&m.pluginsOrdered, e.Name)
	case appui.PluginStarted:
		m.pluginRows[e.Name] = pluginRow{name: e.Name, running: true}
		addOrder(&m.pluginsOrdered, e.Name)
		m.status = appui.StatusRunning
	case appui.PluginFinished:
		r := m.pluginRows[e.Name]
		r.running, r.done = false, true
		r.domains, r.dur, r.runErr = e.DomainsFound, e.Duration, e.Err
		m.pluginRows[e.Name] = r
	case appui.UniqueProgress:
		m.uniqSeen = e.Count
		m.lastMetrics.CPUPercent = e.CPUPercent
		m.lastMetrics.MemUsedMB = e.MemMB
	case appui.CacheHit:
		m.status = appui.StatusCached
		m.uniqSeen = e.TotalDomains
	case appui.ScanComplete:
		d := e
		m.scanDone = &d
		m.uniqSeen = e.TotalDomains
		if e.FromCache {
			m.status = appui.StatusCached
		} else {
			m.status = appui.StatusDone
		}
		for nm, pi := range e.PluginStats {
			m.pluginRows[nm] = pluginRow{name: nm, done: true, domains: pi.DomainsFound, dur: pi.Duration, runErr: pi.Err}
			addOrder(&m.pluginsOrdered, nm)
		}
	}
}

func addOrder(slice *[]string, n string) {
	for _, x := range *slice {
		if x == n {
			return
		}
	}
	*slice = append(*slice, n)
}
