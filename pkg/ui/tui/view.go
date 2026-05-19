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
	"fmt"
	"strings"
	"time"

	appui "github.com/Ratnadeepdeyroy/ratnosint7/pkg/ui"
	"github.com/charmbracelet/lipgloss"
)

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *dashModel) pluginsRunningCount() int {
	n := 0
	for _, r := range m.pluginRows {
		if r.running {
			n++
		}
	}
	return n
}

func (m *dashModel) pluginsDoneCount() int {
	n := 0
	for _, r := range m.pluginRows {
		if r.done {
			n++
		}
	}
	return n
}

func truncView(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 3 {
		return string(r[:maxInt(1, n)])
	}
	return string(r[:n-3]) + "..."
}

func domainsLabelRow(n int) string {
	if n == 1 {
		return "1 subdomain"
	}
	return fmt.Sprintf("%d subdomains", n)
}

func durLabelRow(r pluginRow) string {
	if !r.done {
		return "—"
	}
	if r.dur <= 0 {
		return "0s"
	}
	return r.dur.Round(time.Millisecond).String()
}

func iconRow(r pluginRow) string {
	switch {
	case !r.done && r.running:
		return "[*] "
	case !r.done:
		return "[ ] "
	case r.runErr != nil:
		return "[!] "
	default:
		return "[+] "
	}
}

func footerLine(color bool, w int) string {
	txt := "ctrl+c cancel | stderr shown with --debug"
	if color {
		return lipgloss.NewStyle().Faint(true).Width(w).Render(txt)
	}
	return txt
}

func (m *dashModel) View() string {
	w := maxInt(42, m.width)
	banner := lipgloss.NewStyle().Bold(true).Width(w).Align(lipgloss.Center).Render("ratnosint7 recon engine")
	boxOuter := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1).
		Width(minInt(w, lipgloss.Width(banner)+8))
	header := boxOuter.Render(banner)

	met := m.lastMetrics

	accentLabel := func(s string) string {
		if m.color {
			return m.styles.Accent.Render(s + ":")
		}
		return s + ":"
	}

	statusVal := strings.ToLower(string(m.status))
	statusStr := statusVal
	if m.color {
		switch statusVal {
		case "running":
			statusStr = m.styles.Running.Render(statusVal)
		case "done":
			statusStr = m.styles.Success.Render(statusVal)
		case "cached":
			statusStr = m.styles.WarnStyled.Render(statusVal)
		}
	}

	lines := []string{
		fmt.Sprintf(" %-14s %s", accentLabel("Target"), truncView(m.domain, 48)),
		fmt.Sprintf(" %-14s %d", accentLabel("Scan ID"), m.scanID),
		fmt.Sprintf(" %-14s %d", accentLabel("Active plugins"), m.pluginsRunningCount()),
		fmt.Sprintf(" %-14s %.0f%%", accentLabel("CPU"), met.CPUPercent),
		fmt.Sprintf(" %-14s %.0fMB", accentLabel("Memory"), met.MemUsedMB),
		fmt.Sprintf(" %-14s %s", accentLabel("Status"), statusStr),
		fmt.Sprintf(" %-14s %d", accentLabel("Unique domains"), m.uniqSeen),
		fmt.Sprintf(" %-14s %d", accentLabel("Goroutines"), met.NumGoroutine),
	}
	stats := strings.Join(lines, "\n")

	total := m.pluginPlan
	if total < 1 {
		total = maxInt(1, len(m.pluginsOrdered))
	}
	done := clampInt(m.pluginsDoneCount(), 0, total)

	bar := appui.RenderProgressBar(m.styles, done, total, clampInt(w-28, 16, 36), m.color)

	tools := m.renderTools(w - 8)
	body := lipgloss.JoinVertical(lipgloss.Left, header, "", stats, "", bar, "", tools)

	if m.scanDone != nil {
		body += "\n\n" + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EA8C55")).Render("scan complete")
	}

	footer := footerLine(m.color, w)
	if m.waitingExit {
		txt := "scan complete | press any key to exit"
		if m.color {
			footer = lipgloss.NewStyle().Faint(true).Width(w).Render(txt)
		} else {
			footer = txt
		}
	}
	return body + "\n" + footer + "\n"
}

func (m *dashModel) renderTools(maxInner int) string {
	var b strings.Builder
	for _, name := range m.pluginsOrdered {
		r := m.pluginRows[name]
		pre := ""
		if r.running {
			pre = m.spinner.View() + " "
		}
		warn := ""
		if r.runErr != nil && r.done {
			if m.color {
				warn = "  " + m.styles.ErrorStyled.Render("error")
			} else {
				warn = "  error"
			}
		}
		fmt.Fprintf(&b, "%s%s%-15s %-12s %-10s%s\n", pre, iconRow(r),
			truncView(name, 13), domainsLabelRow(r.domains), durLabelRow(r), warn)
	}
	txt := strings.TrimSuffix(b.String(), "\n")
	if txt == "" {
		txt = " (waiting for tools to start)"
	}
	fr := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Padding(0, 1)
	if maxInner > 10 {
		fr = fr.Width(maxInner)
	}
	return fr.Render(txt)
}
