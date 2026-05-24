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

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/config"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/installer"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/ratelimit"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/runner"
)

// ConfigPlugin implements Plugin from ToolConfig (YAML-driven).
type ConfigPlugin struct {
	Config *config.ToolConfig
	Mode   string // "passive", "active", or "both"
}

func (c ConfigPlugin) Name() string {
	return c.Config.Name
}

func (c ConfigPlugin) InstallSpec() InstallSpec {
	args := installer.SplitCommand(c.Config.Install)
	if len(args) == 0 {
		return InstallSpec{Name: c.Config.Name, Cmd: "", Args: nil}
	}
	return InstallSpec{
		Name: c.Config.Name,
		Cmd:  args[0],
		Args: args[1:],
	}
}

func (c ConfigPlugin) Preflight(ctx context.Context) error {
	_ = ctx
	if err := installer.CheckPython312ForTool(*c.Config); err != nil {
		return fmt.Errorf("%s: %w", c.Config.Name, err)
	}
	return nil
}

func (c ConfigPlugin) execCmd(ctx context.Context, rawCmd, domain string, output chan<- string) error {
	rawCmd = strings.ReplaceAll(rawCmd, "{domain}", domain)
	if err := installer.CheckPython312ForTool(*c.Config); err != nil {
		return err
	}
	var workDir string
	if c.Config.WorkDir != "" {
		workDir = filepath.Join(installer.ToolsDir(), c.Config.WorkDir)
		// {venv_python} resolves to the platform-correct interpreter inside the tool's .venv.
		venvPy := installer.VenvPythonBin(filepath.Join(workDir, ".venv"))
		rawCmd = strings.ReplaceAll(rawCmd, "{venv_python}", venvPy)
	}
	args := splitRunCommand(rawCmd)
	if len(args) == 0 {
		return nil
	}
	exe := args[0]
	runArgs := args[1:]
	if workDir != "" && !filepath.IsAbs(exe) && (strings.Contains(exe, "/") || strings.ContainsRune(exe, filepath.Separator)) {
		exe = filepath.Join(workDir, filepath.FromSlash(exe))
	}
	return runner.RunWithDirAndTimeout(ctx, workDir, exe, runArgs, output, c.Config.Timeout)
}

func (c ConfigPlugin) Run(ctx context.Context, domain string, output chan<- string) error {
	if c.Mode == "both" {
		cap := c.Config.RateLimit
		if cap < 1 {
			cap = 1
		}
		limiter := ratelimit.New(ctx, cap, time.Second)

		var wg sync.WaitGroup
		var mu sync.Mutex
		var firstErr error

		launch := func(cmd string) {
			defer wg.Done()
			if err := limiter.Wait(ctx); err != nil {
				return
			}
			if err := c.execCmd(ctx, cmd, domain, output); err != nil && ctx.Err() == nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}

		if c.Config.PassiveRun != "" {
			wg.Add(1)
			go launch(c.Config.PassiveRun)
		}
		if c.Config.ActiveRun != "" {
			wg.Add(1)
			go launch(c.Config.ActiveRun)
		}
		wg.Wait()
		return firstErr
	}

	var runCmd string
	if c.Mode == "active" {
		runCmd = c.Config.ActiveRun
	} else {
		runCmd = c.Config.PassiveRun
	}
	if runCmd == "" {
		return nil
	}
	return c.execCmd(ctx, runCmd, domain, output)
}

func splitRunCommand(s string) []string {
	var args []string
	var current strings.Builder
	for _, r := range s {
		if r == ' ' || r == '\t' {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}
