//go:build windows

package main

import (
	"log"
	"os"
	"time"
)

func main() {
	if !ensureSingleInstance() {
		return
	}
	defer releaseSingleInstance()

	configPath := defaultConfigPath()
	cfg := loadOrCreateConfig(configPath)

	if err := ensureAutoStart(); err != nil {
		log.Printf("autostart: %v", err)
	}

	state := newTimerState(cfg)

	// Session listener runs in background and drives lock/unlock state.
	go func() {
		if err := listenSessionEvents(state); err != nil {
			log.Printf("session listener error: %v", err)
		}
	}()

	// Timer loop
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			state.tick()
		}
	}()

	// Tray blocks until exit.
	runTray(state, configPath)

	// Clean shutdown
	state.stop()
	os.Exit(0)
}
