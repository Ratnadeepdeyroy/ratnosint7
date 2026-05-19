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

	apperrors "github.com/Ratnadeepdeyroy/ratnosint7/pkg/errors"
)

// FormatPluginError renders categorized errors with Reason/Fix.
func FormatPluginError(st Styles, plugin string, e *apperrors.PluginError, color bool) string {
	if e == nil {
		return fmt.Sprintf("[ERROR] %s: (no details)", plugin)
	}
	title := fmt.Sprintf("[%s] %s", errLabel(e), plugin)
	if color {
		if apperrors.IsFatal(e.Code) {
			title = st.ErrorStyled.Bold(true).Render(title)
		} else {
			title = st.WarnStyled.Bold(true).Render(title)
		}
	}
	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\nReason:\n  ")
	reason := e.Reason
	if reason == "" && e.Err != nil {
		reason = e.Err.Error()
	}
	b.WriteString(reason)
	b.WriteByte('\n')
	if e.Fix != "" {
		b.WriteString("Fix:\n  ")
		b.WriteString(e.Fix)
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func errLabel(e *apperrors.PluginError) string {
	if e == nil {
		return "ERROR"
	}
	if apperrors.IsRetryable(e.Code) || e.Code == apperrors.ContextCancelled {
		return "WARN"
	}
	return "ERROR"
}
