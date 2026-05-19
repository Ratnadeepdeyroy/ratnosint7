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

package plugins

import (
	"context"

	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/runner"
)

// Amass implements the amass plugin.
type Amass struct {
	ResolversPath string
}

func (Amass) Name() string { return "amass" }

func (Amass) InstallSpec() InstallSpec {
	return InstallSpec{
		Name: "amass",
		Cmd:  "go",
		Args: []string{"install", "github.com/owasp-amass/amass/v4/...@v4.2.0"},
	}
}

func (Amass) Preflight(ctx context.Context) error {
	return nil
}

func (a Amass) Run(ctx context.Context, domain string, output chan<- string) error {
	args := []string{"enum", "-passive", "-d", domain}
	if a.ResolversPath != "" {
		args = append(args, "-rf", a.ResolversPath)
	}
	return runner.Run(ctx, "amass", args, output)
}
