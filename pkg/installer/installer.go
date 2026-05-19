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

// Package installer manages tool installation.
package installer

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/config"
)

// GoBinPath returns GOPATH/bin (where go install puts binaries).
func GoBinPath() string {
	out, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "go", "bin")
	}
	return filepath.Join(strings.TrimSpace(string(out)), "bin")
}

// ToolsDir returns ~/.ratnosint7/tools for downloaded binaries.
func ToolsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ratnosint7", "tools")
}

// PathWithGoBin returns PATH with common tool dirs prepended (GOPATH/bin, tools, cargo/bin, homebrew).
func PathWithGoBin() string {
	goBin := GoBinPath()
	home, _ := os.UserHomeDir()
	toolsDir := ToolsDir()
	cargoBin := filepath.Join(home, ".cargo", "bin")
	brewBins := []string{"/opt/homebrew/bin", "/usr/local/bin"}
	path := os.Getenv("PATH")
	localBin := filepath.Join(home, ".local", "bin")
	prepend := []string{goBin, toolsDir, cargoBin, localBin}
	for _, b := range brewBins {
		if _, err := os.Stat(b); err == nil {
			prepend = append(prepend, b)
			break
		}
	}
	joined := strings.Join(prepend, string(os.PathListSeparator))
	if path == "" {
		return joined
	}
	return joined + string(os.PathListSeparator) + path
}

// ToolExists checks if a tool is in PATH or common tool dirs.
func ToolExists(name string) bool {
	return ToolPath(name) != ""
}

// ToolExistsWithWorkDir checks if a tool with workdir is installed (e.g. sublist3r).
func ToolExistsWithWorkDir(name, workDir string) bool {
	if workDir == "" {
		return ToolExists(name)
	}
	p := filepath.Join(ToolsDir(), workDir)
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

// ToolPath returns the full path to the executable, or empty if not found.
func ToolPath(name string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	home, _ := os.UserHomeDir()
	candidates := []string{name}
	if runtime.GOOS == "windows" && !strings.HasSuffix(name, ".exe") {
		candidates = append(candidates, name+".exe")
	}
	for _, dir := range []string{
		GoBinPath(),
		ToolsDir(),
		filepath.Join(home, ".cargo", "bin"),
		filepath.Join(home, ".local", "bin"),
		"/opt/homebrew/bin",
		"/usr/local/bin",
	} {
		for _, n := range candidates {
			p := filepath.Join(dir, n)
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				return p
			}
		}
	}
	return ""
}

// Install runs the install command.
func Install(ctx context.Context, name, cmd string, args []string) error {
	c := exec.CommandContext(ctx, cmd, args...)
	var stderr bytes.Buffer
	c.Stderr = &stderr
	if err := c.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("install %s: %w\n%s", name, err, stderr.String())
		}
		return fmt.Errorf("install %s: %w", name, err)
	}
	return nil
}

// InstallFindomainFromGitHub downloads findomain from GitHub releases (no cargo/brew).
func InstallFindomainFromGitHub(ctx context.Context) error {
	asset := findomainAsset()
	if asset == "" {
		return fmt.Errorf("findomain: unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	url := "https://github.com/Findomain/Findomain/releases/download/10.0.1/" + asset
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download findomain: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download findomain: %s", resp.Status)
	}
	dir := ToolsDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmpZip := filepath.Join(dir, "findomain.zip")
	f, err := os.Create(tmpZip)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		os.Remove(tmpZip)
		return err
	}
	defer os.Remove(tmpZip)
	r, err := zip.OpenReader(tmpZip)
	if err != nil {
		return err
	}
	defer r.Close()
	dest := filepath.Join(dir, "findomain")
	if runtime.GOOS == "windows" {
		dest = filepath.Join(dir, "findomain.exe")
	}
	for _, zf := range r.File {
		base := filepath.Base(zf.Name)
		if base == "findomain" || base == "findomain.exe" {
			if zf.FileInfo().IsDir() {
				continue
			}
			rc, err := zf.Open()
			if err != nil {
				return err
			}
			out, err := os.Create(dest)
			if err != nil {
				rc.Close()
				return err
			}
			_, err = io.Copy(out, rc)
			rc.Close()
			out.Close()
			if err != nil {
				return err
			}
			if runtime.GOOS != "windows" {
				_ = os.Chmod(dest, 0755)
			}
			return nil
		}
	}
	return fmt.Errorf("findomain: no binary in zip")
}

func findomainAsset() string {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "findomain-osx-arm64.zip"
		}
		return "findomain-osx-x86_64.zip"
	case "linux":
		if runtime.GOARCH == "386" {
			return "findomain-linux-i386.zip"
		}
		return "findomain-linux.zip"
	case "windows":
		if runtime.GOARCH == "386" {
			return "findomain-windows-i686.exe.zip"
		}
		return "findomain-windows.exe.zip"
	}
	return ""
}

// InstallFromConfig runs the install command from ToolConfig.
// Supports "cmd1 && cmd2" for multi-step installs (e.g. git clone && cd X && pip install).
// If git clone would fail (dir exists), remove old dir first for idempotent re-installs.
func InstallFromConfig(ctx context.Context, tc config.ToolConfig) error {
	parts := strings.Split(tc.Install, " && ")
	toolsDir := ToolsDir()
	if err := os.MkdirAll(toolsDir, 0755); err != nil {
		return err
	}

	if err := checkTugareconInstallPrereqs(tc); err != nil {
		return err
	}

	// Pre-emptively remove clone dirs for idempotent re-installs.
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "git clone") {
			args := SplitCommand(part)
			if len(args) >= 3 {
				// Extract repo name from URL: https://github.com/user/RepoName.git -> RepoName
				repoURL := args[len(args)-1]  // Last arg is the URL
				repoName := filepath.Base(repoURL)
				repoName = strings.TrimSuffix(repoName, ".git")
				cloneDir := filepath.Join(toolsDir, repoName)
				if _, err := os.Stat(cloneDir); err == nil {
					// Dir exists, remove it for clean re-install
					if rmErr := os.RemoveAll(cloneDir); rmErr != nil {
						return fmt.Errorf("remove old clone dir %s: %w", cloneDir, rmErr)
					}
				}
			}
		}
	}

	var workDir string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "cd ") {
			dir := strings.TrimSpace(part[3:])
			workDir = filepath.Join(toolsDir, dir)
			continue
		}
		args := SplitCommand(part)
		if len(args) == 0 {
			continue
		}
		c := exec.CommandContext(ctx, args[0], args[1:]...)
		if workDir != "" {
			c.Dir = workDir
		} else if strings.HasPrefix(part, "git clone") {
			c.Dir = toolsDir
		}
		var stderr bytes.Buffer
		c.Stderr = &stderr
		if err := c.Run(); err != nil {
			if stderr.Len() > 0 {
				return fmt.Errorf("install %s: %w\n%s", tc.Name, err, formatInstallStderr(tc, stderr.String()))
			}
			return fmt.Errorf("install %s: %w", tc.Name, err)
		}
	}
	if strings.EqualFold(tc.Name, "tugarecon") {
		if workDir == "" {
			workDir = filepath.Join(toolsDir, "tugarecon")
		}
		if err := installTugareconVenv(ctx, workDir); err != nil {
			return err
		}
	}
	return nil
}

const python312Bin = "python3.12"

// TugareconVenvPython returns the venv interpreter path for tugarecon, or empty if missing.
func TugareconVenvPython() string {
	return filepath.Join(ToolsDir(), "tugarecon", ".venv", "bin", "python")
}

func checkTugareconInstallPrereqs(tc config.ToolConfig) error {
	if !strings.EqualFold(tc.Name, "tugarecon") {
		return nil
	}
	return tugareconPython312Missing(tc.Name)
}

// CheckPython312ForTool validates tugarecon is ready to run (Python 3.12 + local venv).
func CheckPython312ForTool(tc config.ToolConfig) error {
	if !strings.EqualFold(tc.Name, "tugarecon") {
		return nil
	}
	if err := tugareconPython312Missing(tc.Name); err != nil {
		return err
	}
	venvPy := TugareconVenvPython()
	if st, err := os.Stat(venvPy); err != nil || st.IsDir() {
		return fmt.Errorf(
			"%s is not installed (missing virtualenv) — install Python 3.12, then run: ratnosint7 update-tools",
			tc.Name,
		)
	}
	return nil
}

func tugareconPython312Missing(toolName string) error {
	if _, err := exec.LookPath(python312Bin); err != nil {
		return fmt.Errorf(
			"%s requires Python 3.12 but %q was not found on PATH — install: brew install python@3.12",
			toolName, python312Bin,
		)
	}
	return nil
}

func installTugareconVenv(ctx context.Context, workDir string) error {
	py312, err := exec.LookPath(python312Bin)
	if err != nil {
		return fmt.Errorf(
			"install tugarecon: Python 3.12 is required but %q was not found on PATH — install: brew install python@3.12",
			python312Bin,
		)
	}
	if err := runInDir(ctx, workDir, py312, []string{"-m", "venv", ".venv"}); err != nil {
		return fmt.Errorf("install tugarecon: create virtualenv: %w", err)
	}
	venvPy := filepath.Join(workDir, ".venv", "bin", "python")
	if err := runInDir(ctx, workDir, venvPy, []string{"-m", "pip", "install", "-r", "requirements.txt"}); err != nil {
		return fmt.Errorf("install tugarecon: pip in virtualenv: %w", err)
	}
	return nil
}

func runInDir(ctx context.Context, dir, exe string, args []string) error {
	c := exec.CommandContext(ctx, exe, args...)
	c.Dir = dir
	var stderr bytes.Buffer
	c.Stderr = &stderr
	if err := c.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("%w\n%s", err, stderr.String())
		}
		return err
	}
	return nil
}

func formatInstallStderr(tc config.ToolConfig, stderr string) string {
	if strings.Contains(stderr, "externally-managed-environment") {
		return stderr + "\n(hint: tugarecon installs into a local .venv — rebuild ratnosint7 and run update-tools again)"
	}
	if strings.EqualFold(tc.Name, "tugarecon") && strings.Contains(stderr, "dnspython>=2.8.0") {
		return stderr + "\n(hint: tugarecon needs Python 3.12 — install: brew install python@3.12, then update-tools)"
	}
	return stderr
}

// SplitCommand splits an install command into cmd and args.
func SplitCommand(s string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	for _, r := range s {
		switch {
		case r == '"' || r == '\'':
			inQuote = !inQuote
		case (r == ' ' || r == '\t') && !inQuote:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

// SuggestedInstallFix returns a concise fix hint shown when installs fail or binaries are missing.
func SuggestedInstallFix(toolName string) string {
	n := strings.ToLower(strings.TrimSpace(toolName))
	if n == "" {
		return "Install the tool from upstream docs and ensure it is on PATH."
	}
	switch n {
	case "findomain":
		return "brew install findomain — or use ratnosint7's bundled GitHub download path for findomain."
	case "subfinder":
		return "go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest — or brew install subfinder."
	case "amass":
		return "brew install amass — or install upstream release for your OS."
	case "assetfinder":
		return "go install github.com/tomnomnom/assetfinder@latest"
	case "dnsx":
		return "go install github.com/projectdiscovery/dnsx/cmd/dnsx@latest — or brew where available."
	case "tugarecon":
		return "Python 3.12 required (brew install python@3.12) — update-tools creates ~/.ratnosint7/tools/tugarecon/.venv"
	case "turbolist3r":
		return "git clone https://github.com/fleetcaptain/Turbolist3r.git ~/.ratnosint7/tools/Turbolist3r && cd ~/.ratnosint7/tools/Turbolist3r && pip3 install -r requirements.txt"
	default:
		return fmt.Sprintf("brew install %s — or install from upstream and ensure PATH includes the binary.", n)
	}
}
