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

package parser

import "testing"

func TestIsValidDomain(t *testing.T) {
	tests := []struct {
		name   string
		domain string
		want   bool
	}{
		{"valid", "example.com", true},
		{"valid subdomain", "sub.example.com", true},
		{"valid punycode", "xn--domain.com", true},
		{"single label", "example", false},
		{"semver noise", "v1.1", false},
		{"wildcard", "*.example.com", false},
		{"trailing dot", "example.com.", false},
		{"empty labels", "example..com", false},
		{"leading hyphen", "-example.com", false},
		{"trailing hyphen", "example-.com", false},
		{"uppercase", "Example.COM", false},
		{"too long", string(make([]byte, 254)), false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidDomain(tt.domain)
			if got != tt.want {
				t.Errorf("IsValidDomain(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}

func TestNormalizeAndValidate(t *testing.T) {
	tests := []struct {
		in    string
		want  string
		valid bool
	}{
		{"  example.com  ", "example.com", true},
		{"Example.COM", "example.com", true},
		{"example.com.", "example.com", true},
		{"invalid", "", false},
		{"v1.1 semver noise", "", false},
		{"2.14.333 release-ish", "", false},
	}
	for _, tt := range tests {
		got, ok := NormalizeAndValidate(tt.in)
		if ok != tt.valid || got != tt.want {
			t.Errorf("NormalizeAndValidate(%q) = (%q, %v), want (%q, %v)", tt.in, got, ok, tt.want, tt.valid)
		}
	}
}

func BenchmarkParser(b *testing.B) {
	domains := []string{"example.com", "sub.example.com", "a.b.c.example.com", "xn--test.com"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := domains[i%len(domains)]
		NormalizeAndValidate(d)
	}
}
