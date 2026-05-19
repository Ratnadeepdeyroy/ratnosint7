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

package config

// ToolConfig matches the tools.yaml schema for a single tool.
type ToolConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Install     string `yaml:"install"`
	VersionCmd  string `yaml:"version_cmd"`
	PassiveRun  string `yaml:"passive_run"`
	ActiveRun   string `yaml:"active_run"`
	RateLimit   int    `yaml:"rate_limit"`
	Concurrency int    `yaml:"concurrency"`
	WorkDir     string `yaml:"workdir"`
	Timeout     int    `yaml:"timeout,omitempty"`
}
