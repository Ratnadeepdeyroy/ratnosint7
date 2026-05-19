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
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	apperrors "github.com/Ratnadeepdeyroy/ratnosint7/pkg/errors"
)

// SummaryFields holds totals for the boxed scan summary.
type SummaryFields struct {
	Domain             string
	OutputPath         string
	TotalDomains       int
	Duration           time.Duration
	FromCache          bool
	CacheAge           time.Duration
	PluginsFailedCount int
	Errors             []*apperrors.PluginError
}

// RenderSummaryBox returns a framed scan summary block.
func RenderSummaryBox(st Styles, sf SummaryFields, color bool) string {
	var lines []string
	if color {
		lines = append(lines, st.Title.Render("Scan Summary"))
	} else {
		lines = append(lines, "Scan Summary")
	}
	add := func(label, val string) {
		if color {
			lbl := lipgloss.NewStyle().Foreground(lipgloss.Color("#A8A29E")).Render(label + ":")
			lines = append(lines, fmt.Sprintf("%s %s", lbl, val))
			return
		}
		lines = append(lines, fmt.Sprintf("%-18s%s", label+":", val))
	}
	add("Target", sf.Domain)
	add("Total subdomains", fmt.Sprintf("%d", sf.TotalDomains))
	add("Duration", sf.Duration.Round(time.Millisecond).String())
	add("Output file", sf.OutputPath)
	if sf.FromCache {
		cv := "💤 cache hit · true"
		if sf.CacheAge > 0 {
			cv += fmt.Sprintf(" · age %s", formatShortDur(sf.CacheAge))
		}
		add("Cached", cv)
	} else {
		add("Cached", "false")
	}
	if sf.PluginsFailedCount > 0 {
		add("Failed tools", fmt.Sprintf("%d", sf.PluginsFailedCount))
	}
	body := strings.Join(lines, "\n")
	return st.Box.Render(body) + "\n"
}

func formatShortDur(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	return d.String()
}

// FormatErrors prints writer/pipeline PluginErrors beneath the summary.
func FormatErrors(st Styles, errs []*apperrors.PluginError, color bool) string {
	if len(errs) == 0 {
		return ""
	}
	var b strings.Builder
	for _, e := range errs {
		if e == nil {
			continue
		}
		b.WriteString(FormatPluginError(st, e.Plugin, e, color))
		b.WriteByte('\n')
	}
	return b.String()
}

// RenderSep is a divider before boxed summary on static terminals.
func RenderSep(color bool) string {
	if color {
		return "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	}
	return "------------------------------------"
}
