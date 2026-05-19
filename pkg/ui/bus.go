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

import "sync"

const subChanCap = 512

// Bus distributes UI events from the engine to subscribers (TUI, static renderer).
type Bus struct {
	mu   sync.RWMutex
	subs []chan Event
}

func NewBus() *Bus {
	return &Bus{}
}

func (b *Bus) Subscribe() <-chan Event {
	ch := make(chan Event, subChanCap)
	b.mu.Lock()
	b.subs = append(b.subs, ch)
	b.mu.Unlock()
	return ch
}

func (b *Bus) Publish(e Event) {
	if b == nil || e == nil {
		return
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs {
		ch <- e
	}
}

// CloseSubscribers closes subscriber channels after a scan/UI session ends.
func (b *Bus) CloseSubscribers() {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subs {
		close(ch)
	}
	b.subs = nil
}
