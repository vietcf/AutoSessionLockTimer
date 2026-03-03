//go:build windows

package main

import (
	"sync"
	"time"
)

type TimerState struct {
	mu          sync.Mutex
	cfg         Config
	lastUnlock  time.Time
	running     bool
	isLocked    bool
	shouldExit  bool
	lockedFired bool
}

func newTimerState(cfg Config) *TimerState {
	return &TimerState{
		cfg:        cfg,
		running:    cfg.Enabled,
		isLocked:   false,
		lastUnlock: time.Now(),
	}
}

func (s *TimerState) updateConfig(cfg Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg
	if !cfg.Enabled {
		s.running = false
	}
}

func (s *TimerState) setEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.Enabled = enabled
	if !enabled {
		s.running = false
	}
}

func (s *TimerState) onUnlock() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isLocked = false
	s.lastUnlock = time.Now()
	s.lockedFired = false
	if s.cfg.Enabled {
		s.running = true
	}
}

func (s *TimerState) onLock() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isLocked = true
	s.running = false
}

func (s *TimerState) tick() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.shouldExit || !s.running || s.isLocked || !s.cfg.Enabled {
		return
	}
	if s.lastUnlock.IsZero() {
		return
	}
	elapsed := time.Since(s.lastUnlock)
	if elapsed >= time.Duration(s.cfg.LockMinutes)*time.Minute {
		lockWorkstation()
		s.running = false
		s.lockedFired = true
	}
}

func (s *TimerState) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shouldExit = true
	s.running = false
}

func (s *TimerState) snapshot() Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cfg
}

func (s *TimerState) elapsedMinutesSinceLastUnlock() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastUnlock.IsZero() {
		return 0
	}
	elapsed := time.Since(s.lastUnlock)
	if elapsed < 0 {
		return 0
	}
	return int(elapsed / time.Minute)
}
