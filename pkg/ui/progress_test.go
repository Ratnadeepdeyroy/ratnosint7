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
	"strings"
	"testing"
)

func TestRenderProgressBar(t *testing.T) {
	st := NewStyles(ThemeDefault, false)
	got := RenderProgressBar(st, 4, 13, 20, false)
	if !strings.Contains(got, "Plugins") {
		t.Fatalf("missing label: %q", got)
	}
	if !strings.Contains(got, "30%") {
		t.Fatalf("missing percent: %q", got)
	}
	if !strings.Contains(got, "4/13") {
		t.Fatalf("missing counter: %q", got)
	}
	if !strings.Contains(got, "█") || !strings.Contains(got, "░") {
		t.Fatalf("missing block chars: %q", got)
	}
}

func TestRenderProgressBarClamps(t *testing.T) {
	st := NewStyles(ThemeDefault, false)
	got := RenderProgressBar(st, 99, 10, 12, false)
	if !strings.Contains(got, "10/10") {
		t.Fatalf("expected clamped total: %q", got)
	}
	if !strings.Contains(got, "100%") {
		t.Fatalf("expected 100%%: %q", got)
	}
}
