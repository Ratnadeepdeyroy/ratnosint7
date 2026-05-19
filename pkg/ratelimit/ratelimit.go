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

// Package ratelimit provides a token bucket rate limiter.
package ratelimit

import (
	"context"
	"time"
)

// RateLimiter is a token bucket rate limiter.
type RateLimiter struct {
	tokens chan struct{}
}

// New creates a RateLimiter with the given capacity.
// The refill goroutine stops when ctx is cancelled.
func New(ctx context.Context, capacity int, refillRate time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens: make(chan struct{}, capacity),
	}
	for i := 0; i < capacity; i++ {
		rl.tokens <- struct{}{}
	}
	go rl.refill(ctx, refillRate)
	return rl
}

func (rl *RateLimiter) refill(ctx context.Context, rate time.Duration) {
	ticker := time.NewTicker(rate)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			select {
			case rl.tokens <- struct{}{}:
			default:
			}
		}
	}
}

// Wait blocks until a token is available or ctx is cancelled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
