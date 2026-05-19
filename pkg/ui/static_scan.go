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
	"context"
	"fmt"
	"io"
	"time"

	apperrors "github.com/Ratnadeepdeyroy/ratnosint7/pkg/errors"
)

// ConsumeScan prints events until ScanComplete or ctx cancelled.
func ConsumeScan(ctx context.Context, w io.Writer, bus *Bus, theme Theme, useColor bool) error {
	if bus == nil || w == nil {
		return nil
	}
	st := NewStyles(theme, useColor)
	sub := bus.Subscribe()
	pluginDone := 0
	pluginTotal := 0
	for {
		select {
		case <-ctx.Done():
			msg := "scan cancelled"
			if useColor {
				msg = st.Dim.Render(msg)
			}
			_, err := fmt.Fprintln(w, msg)
			return firstErrNonNil(err, ctx.Err())
		case ev, ok := <-sub:
			if !ok {
				return nil
			}
			switch e := ev.(type) {
			case ScanStarted:
				pluginTotal = e.PluginCount
			case PluginFinished, PluginSkipped:
				// counted after the event line is printed
				_ = e
			}
			done, err := writeScanEvent(w, st, ev, useColor)
			if err != nil {
				return err
			}
			switch ev.(type) {
			case PluginFinished, PluginSkipped:
				pluginDone++
				if pluginTotal > 0 {
					fmt.Fprintf(w, "%s\n", RenderProgressBar(st, pluginDone, pluginTotal, defaultProgressWidth, useColor))
				}
			}
			if done {
				return nil
			}
		}
	}
}

func firstErrNonNil(a, b error) error {
	if a != nil {
		return a
	}
	return b
}

func writeScanEvent(w io.Writer, st Styles, ev Event, color bool) (done bool, err error) {
	switch e := ev.(type) {
	case ScanStarted:
		if color {
			_, err = fmt.Fprintf(w, "%s\n%s %s\n%s %d\n%s %d\n\n",
				st.Title.Render("ratnosint7 recon engine"),
				st.Accent.Render("Target:"), truncateRunesStr(e.Domain, 60),
				st.Accent.Render("Scan ID:"), e.ScanID,
				st.Accent.Render("Plugins:"), e.PluginCount)
		} else {
			_, err = fmt.Fprintf(w, "ratnosint7 recon engine\nTarget: %s\nScan ID: %d\nPlugins: %d\n\n",
				e.Domain, e.ScanID, e.PluginCount)
		}
		return false, err
	case PluginSkipped:
		_, err := fmt.Fprintf(w, "%s\n", PluginSkipLine(st, e.Name, e.Err, color))
		return false, err
	case PluginStarted:
		if color {
			_, err = fmt.Fprintf(w, "%s %-16s %s\n", st.Running.Render("⟳"), e.Name, st.Dim.Render("running…"))
		} else {
			_, err = fmt.Fprintf(w, "[run] %-16s running…\n", e.Name)
		}
		return false, err
	case PluginFinished:
		icon := "✓"
		if color {
			icon = st.Success.Render("✓")
			if e.Err != nil {
				icon = st.ErrorStyled.Render("✗")
			}
		} else if e.Err != nil {
			icon = "✗"
		}
		suffix := ""
		if e.Err != nil {
			suffix = "  (see error log)"
		}
		_, err := fmt.Fprintf(w, "%s %-16s %-13s %-10s%s\n",
			icon, truncateRunesStr(e.Name, 16), pluralDom(e.DomainsFound), fmtDurScan(e.Duration), suffix)
		return false, err
	case UniqueProgress:
		return false, nil
	case CacheHit:
		if color {
			_, err = fmt.Fprintf(w, "%s\n%s\n%s %s\n\n",
				st.WarnStyled.Render("💤 Cache hit"),
				st.Dim.Render("Previous scan reused"),
				st.Accent.Render("Age:"), formatShortDur(e.Age))
		} else {
			_, err = fmt.Fprintf(w, "Cache hit · previous scan reused · age %s\n\n", formatShortDur(e.Age))
		}
		return false, err
	case ScanComplete:
		sf := SummaryFields{
			Domain: e.Domain, OutputPath: e.OutputPath, TotalDomains: e.TotalDomains,
			Duration: e.Duration, FromCache: e.FromCache, CacheAge: e.CacheAge,
			PluginsFailedCount: e.PluginsFailedCount, Errors: e.Errors,
		}
		if _, err = fmt.Fprintf(w, "%s\n", RenderSep(color)); err != nil {
			return false, err
		}
		if _, err = fmt.Fprint(w, RenderSummaryBox(st, sf, color)); err != nil {
			return false, err
		}
		if txt := FormatErrors(st, e.Errors, color); txt != "" {
			fmt.Fprintf(w, "%s", txt)
		}
		return true, nil
	default:
		return false, nil
	}
}

func truncateRunesStr(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 3 {
		return string(r[:max2(1, n)])
	}
	return string(r[:n-3]) + "..."
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func pluralDom(n int) string {
	if n == 1 {
		return "1 subdomain"
	}
	return fmt.Sprintf("%d subdomains", n)
}

func fmtDurScan(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	return d.Round(time.Millisecond).String()
}

// PluginSkipLine formats a skipped tool for static output.
func PluginSkipLine(st Styles, name string, pe *apperrors.PluginError, color bool) string {
	if color && pe != nil {
		return FormatPluginError(st, name, pe, true)
	}
	if pe != nil && pe.Err != nil {
		return fmt.Sprintf("[skip] %-16s %v", name, pe.Err)
	}
	return fmt.Sprintf("[skip] %-16s", name)
}
