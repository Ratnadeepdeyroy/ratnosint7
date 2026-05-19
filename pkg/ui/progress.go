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
)

const (
	defaultProgressWidth = 28
	progressFillRune     = '█'
	progressEmptyRune  = '░'
)

// RenderProgressBar returns a labeled plugin progress line for CLI output.
func RenderProgressBar(st Styles, done, total, width int, color bool) string {
	if total < 1 {
		total = 1
	}
	if width < 10 {
		width = defaultProgressWidth
	}
	if done < 0 {
		done = 0
	}
	if done > total {
		done = total
	}

	fill := width * done / total
	filled := strings.Repeat(string(progressFillRune), fill)
	empty := strings.Repeat(string(progressEmptyRune), width-fill)

	var bar string
	if color {
		bar = st.Dim.Render("[") +
			st.Accent.Render(filled) +
			st.Dim.Render(empty) +
			st.Dim.Render("]")
	} else {
		bar = "[" + filled + empty + "]"
	}

	pct := 0
	if total > 0 {
		pct = done * 100 / total
	}

	label := "Plugins"
	if color {
		label = st.Dim.Render(label)
	}

	return fmt.Sprintf("  %-8s %s  %3d%%  %d/%d", label, bar, pct, done, total)
}
