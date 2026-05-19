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
	"testing"
	"time"
)

func TestBusPublishSubscribe(t *testing.T) {
	b := NewBus()
	ch := b.Subscribe()
	go func() { b.Publish(ScanStarted{Domain: "example.com", ScanID: 42, PluginCount: 2}) }()
	select {
	case ev := <-ch:
		s, ok := ev.(ScanStarted)
		if !ok || s.Domain != "example.com" {
			t.Fatalf("%#v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestCloseSubscribers(t *testing.T) {
	b := NewBus()
	ch := b.Subscribe()
	b.CloseSubscribers()
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected closed")
		}
	case <-time.After(time.Second):
		t.Fatal("close took too long")
	}
}
