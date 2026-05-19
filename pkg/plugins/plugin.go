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

// Package plugins provides tool adapters for subdomain enumeration.
package plugins

import "context"

// InstallSpec describes how to install a plugin.
type InstallSpec struct {
	Name string
	Cmd  string
	Args []string
}

// Plugin is the interface for enumeration tools.
type Plugin interface {
	Name() string
	InstallSpec() InstallSpec
	Preflight(ctx context.Context) error
	Run(ctx context.Context, domain string, output chan<- string) error
}
