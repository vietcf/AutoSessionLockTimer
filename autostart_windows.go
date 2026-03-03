//go:build windows

package main

import (
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

func ensureAutoStart() error {
	exe, err := currentExePath()
	if err != nil {
		return err
	}
	value := quoteIfNeeded(exe)
	key, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetStringValue("AutoLockSessionTimer", value)
}

func currentExePath() (string, error) {
	return os.Executable()
}

func quoteIfNeeded(path string) string {
	if strings.HasPrefix(path, "\"") && strings.HasSuffix(path, "\"") {
		return path
	}
	if strings.Contains(path, " ") {
		return "\"" + path + "\""
	}
	return path
}
