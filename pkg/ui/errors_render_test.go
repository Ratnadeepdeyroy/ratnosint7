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
	"errors"
	"strings"
	"testing"

	apperrors "github.com/Ratnadeepdeyroy/ratnosint7/pkg/errors"
)

func TestFormatPluginErrorNil(t *testing.T) {
	st := NewStyles(ThemeDefault, false)
	s := FormatPluginError(st, "x", nil, false)
	if !strings.Contains(s, "ERROR") || !strings.Contains(s, "x") {
		t.Fatal(s)
	}
}

func TestFormatPluginErrorDetail(t *testing.T) {
	st := NewStyles(ThemeDefault, false)
	in := errors.New("boom")
	pe := apperrors.DetailedPluginError("dnsx", apperrors.InstallFailed, "install failed", "brew install dnsx", in)
	s := FormatPluginError(st, "dnsx", pe, false)
	if !strings.Contains(s, "brew install dnsx") {
		t.Fatal(s)
	}
}
