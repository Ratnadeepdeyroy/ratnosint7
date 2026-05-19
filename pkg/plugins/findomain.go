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

// Findomain implements the findomain plugin.
type Findomain struct{}

func (Findomain) Name() string { return "findomain" }

func (Findomain) InstallSpec() InstallSpec {
	return InstallSpec{
		Name: "findomain",
		Cmd:  "cargo",
		Args: []string{"install", "findomain"},
	}
}

func (Findomain) Preflight(ctx context.Context) error {
	return nil
}

func (f Findomain) Run(ctx context.Context, domain string, output chan<- string) error {
	return runner.Run(ctx, "findomain", []string{"-t", domain, "-q"}, output)
}
