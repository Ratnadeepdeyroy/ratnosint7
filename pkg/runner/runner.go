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

package runner

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/installer"
)

const defaultTimeout = 5 * time.Minute

var debugLogs atomic.Bool

func SetDebugLogs(enabled bool) {
	debugLogs.Store(enabled)
}

func Run(ctx context.Context, name string, args []string, output chan<- string) error {
	return RunWithDir(ctx, "", name, args, output)
}

func RunWithDir(ctx context.Context, dir, name string, args []string, output chan<- string) error {
	return RunWithDirAndTimeout(ctx, dir, name, args, output, 0)
}

func RunWithDirAndTimeout(ctx context.Context, dir, name string, args []string, output chan<- string, toolTimeoutSec int) error {
	timeout := defaultTimeout
	if toolTimeoutSec > 0 {
		timeout = time.Duration(toolTimeoutSec) * time.Second
	}
	runCtx := ctx
	cancel := func() {}
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > timeout {
		runCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	exe := name
	if dir == "" {
		exe = installer.ToolPath(name)
		if exe == "" {
			exe = name
		}
	}
	cmd := exec.CommandContext(runCtx, exe, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", name, err)
	}

	start := time.Now()
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			select {
			case output <- line:
			case <-runCtx.Done():
				if cmd.Process != nil {
					_ = cmd.Process.Kill()
				}
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return fmt.Errorf("%s: timeout after %s", name, timeout)
			}
		}
	}
	if err := cmd.Wait(); err != nil && runCtx.Err() == nil {
		if debugLogs.Load() {
			slog.Debug("tool exited with error", "tool", name, "err", err)
		}
		return err
	}
	if debugLogs.Load() {
		slog.Debug("tool finished", "tool", name, "duration", time.Since(start))
	}
	return nil
}
