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

// Package parser provides domain normalization and validation.
package parser

import (
	"regexp"
	"strings"
)

const (
	maxDomainLen = 253
	maxLabelLen  = 63
)

// bareNumericVersion rejects tool stdout like "v1.1", "2.6.7" mistakenly parsed as hostname.
var bareNumericVersionPattern = regexp.MustCompile(`^[vV]?[0-9]+(?:\.[0-9]+)+(?:\.[0-9]+)*$`)

// isBareNumericVersionNoise returns true for lines such as semver-only strings without a real hostname.
func isBareNumericVersionNoise(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	s = strings.TrimSuffix(strings.ToLower(s), ".")
	return bareNumericVersionPattern.MatchString(s)
}

// IsValidDomain validates a domain per RFC 1123 style rules used for scanners.
func IsValidDomain(s string) bool {
	s = strings.TrimSpace(s)
	if isBareNumericVersionNoise(s) {
		return false
	}
	if len(s) == 0 || len(s) > maxDomainLen {
		return false
	}
	labels := strings.Split(s, ".")
	if len(labels) < 2 {
		return false
	}

	for _, label := range labels {
		if len(label) == 0 || len(label) > maxLabelLen {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, r := range label {
			if !(r == '-' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
				return false
			}
		}
	}

	return true
}

func normalize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ".")
	return strings.ToLower(s)
}

// NormalizeAndValidate normalizes, drops tool version artifacts, validates host shape.
func NormalizeAndValidate(s string) (string, bool) {
	norm := normalize(s)
	if norm == "" {
		return "", false
	}
	if isBareNumericVersionNoise(norm) {
		return "", false
	}
	if strings.Contains(norm, "/") || strings.Contains(norm, "@") || strings.ContainsAny(norm, " \t\n\r") {
		return "", false
	}
	if isIPAddress(norm) || isNumericOnly(norm) {
		return "", false
	}
	if !IsValidDomain(norm) {
		return "", false
	}
	return norm, true
}

func isIPAddress(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}

func isNumericOnly(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
