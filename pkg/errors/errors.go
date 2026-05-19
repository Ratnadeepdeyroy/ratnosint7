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

// Package errors defines typed errors for ratnosint7.
package errors

import "fmt"

type ErrorCode string

const (
	ToolNotFound     ErrorCode = "ToolNotFound"
	NetworkTimeout   ErrorCode = "NetworkTimeout"
	PartialOutput    ErrorCode = "PartialOutput"
	ContextCancelled ErrorCode = "ContextCancelled"
	InstallFailed    ErrorCode = "InstallFailed"
	Unknown          ErrorCode = "Unknown"
)

type PluginError struct {
	Plugin string
	Code   ErrorCode
	Reason string
	Fix    string
	Err    error
}

func (e *PluginError) Error() string {
	if e == nil {
		return ""
	}
	if e.Reason != "" {
		if e.Err != nil {
			return fmt.Sprintf("%s [%s]: %s — %v", e.Plugin, e.Code, e.Reason, e.Err)
		}
		return fmt.Sprintf("%s [%s]: %s", e.Plugin, e.Code, e.Reason)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s [%s]: %v", e.Plugin, e.Code, e.Err)
	}
	return fmt.Sprintf("%s [%s]", e.Plugin, e.Code)
}

func (e *PluginError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewPluginError(plugin string, code ErrorCode, err error) *PluginError {
	pe := &PluginError{Plugin: plugin, Code: code, Err: err}
	if err != nil && pe.Reason == "" {
		pe.Reason = err.Error()
	}
	return pe
}

func DetailedPluginError(plugin string, code ErrorCode, reason, fix string, err error) *PluginError {
	return &PluginError{Plugin: plugin, Code: code, Reason: reason, Fix: fix, Err: err}
}

func IsFatal(code ErrorCode) bool {
	return code == ToolNotFound || code == InstallFailed
}

func IsRetryable(code ErrorCode) bool {
	return code == NetworkTimeout
}

