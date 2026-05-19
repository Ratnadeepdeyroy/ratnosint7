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
	"github.com/charmbracelet/lipgloss"
)

// Theme selects lipgloss palettes (orange/grey).
type Theme string

const (
	ThemeDefault Theme = "default"
)

type themePalette struct {
	title, subtitle, accent, dim, border, success, err, warn, run string
}

func colorsForTheme(t Theme) themePalette {
	return themePalette{
		title: "#EA8C55", subtitle: "#A8A29E", accent: "#EA8C55",
		dim: "#78716F", border: "#C0846B", success: "#F4A574",
		err: "#F87171", warn: "#FBBF24", run: "#FB923C",
	}
}

// Styles bundles themed lipgloss styles.
type Styles struct {
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Box         lipgloss.Style
	Accent      lipgloss.Style
	Dim         lipgloss.Style
	Success     lipgloss.Style
	ErrorStyled lipgloss.Style
	WarnStyled  lipgloss.Style
	Running     lipgloss.Style
	Badge       lipgloss.Style
}

func NewStyles(t Theme, color bool) Styles {
	p := colorsForTheme(t)
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.title))
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color(p.accent))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(p.dim))
	success := lipgloss.NewStyle().Foreground(lipgloss.Color(p.success))
	errSt := lipgloss.NewStyle().Foreground(lipgloss.Color(p.err))
	warn := lipgloss.NewStyle().Foreground(lipgloss.Color(p.warn))
	run := lipgloss.NewStyle().Foreground(lipgloss.Color(p.run))
	box := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1).
		BorderForeground(lipgloss.Color(p.border))
	badge := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.accent)).
		BorderForeground(lipgloss.Color(p.border))
	sub := lipgloss.NewStyle().Foreground(lipgloss.Color(p.subtitle)).Faint(true)

	if !color {
		title = lipgloss.NewStyle().Bold(true)
		accent = lipgloss.NewStyle()
		dim = lipgloss.NewStyle()
		success = lipgloss.NewStyle()
		errSt = lipgloss.NewStyle()
		warn = lipgloss.NewStyle()
		run = lipgloss.NewStyle()
		box = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1)
		badge = lipgloss.NewStyle().Bold(true)
		sub = lipgloss.NewStyle()
	}

	return Styles{
		Title: title, Subtitle: sub, Box: box, Accent: accent, Dim: dim,
		Success: success, ErrorStyled: errSt, WarnStyled: warn, Running: run, Badge: badge,
	}
}
