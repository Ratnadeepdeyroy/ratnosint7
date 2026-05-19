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

// Assetfinder implements the assetfinder plugin.
type Assetfinder struct{}

func (Assetfinder) Name() string { return "assetfinder" }

func (Assetfinder) InstallSpec() InstallSpec {
	return InstallSpec{
		Name: "assetfinder",
		Cmd:  "go",
		Args: []string{"install", "github.com/tomnomnom/assetfinder@v0.1.1"},
	}
}

func (Assetfinder) Preflight(ctx context.Context) error {
	return nil
}

func (a Assetfinder) Run(ctx context.Context, domain string, output chan<- string) error {
	return runner.Run(ctx, "assetfinder", []string{domain}, output)
}
