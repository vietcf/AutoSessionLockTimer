//go:build windows

package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/getlantern/systray"
)

func runTray(state *TimerState, configPath string) {
	systray.Run(func() {
		systray.SetTitle("AutoLock")
		systray.SetIcon(buildICO())
		updateTrayTooltip(state)

		mStart := systray.AddMenuItem("Start", "Enable autolock")
		mStop := systray.AddMenuItem("Stop", "Disable autolock")
		mConfigure := systray.AddMenuItem("Configure Lock Time", "Set lock time")
		systray.AddSeparator()
		mExit := systray.AddMenuItem("Exit", "Exit AutoLock")

		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				updateTrayTooltip(state)
			}
		}()

		go func() {
			for {
				select {
				case <-mStart.ClickedCh:
					cfg := state.snapshot()
					cfg.Enabled = true
					state.setEnabled(true)
					if err := saveConfig(configPath, cfg); err != nil {
						log.Printf("save config: %v", err)
					}
					updateTrayTooltip(state)
				case <-mStop.ClickedCh:
					cfg := state.snapshot()
					cfg.Enabled = false
					state.setEnabled(false)
					if err := saveConfig(configPath, cfg); err != nil {
						log.Printf("save config: %v", err)
					}
					updateTrayTooltip(state)
				case <-mConfigure.ClickedCh:
					cfg := state.snapshot()
					newMinutes, ok, err := showLockTimeDialog(cfg.LockMinutes)
					if err != nil {
						log.Printf("open configure dialog: %v", err)
						continue
					}
					if !ok {
						continue
					}
					cfg.LockMinutes = newMinutes
					state.updateConfig(cfg)
					if err := saveConfig(configPath, cfg); err != nil {
						log.Printf("save config: %v", err)
					}
					updateTrayTooltip(state)
				case <-mExit.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()
	}, func() {})
}

func ensureConfigExists(path string) error {
	_, err := loadConfig(path)
	if err == nil {
		return nil
	}
	return saveConfig(path, defaultConfig())
}

func updateTrayTooltip(state *TimerState) {
	mins := state.elapsedMinutesSinceLastUnlock()
	systray.SetTooltip(fmt.Sprintf("Auto Lock Session Timer - %d minutes since last unlock", mins))
}

func showLockTimeDialog(current int) (minutes int, saved bool, err error) {
	script := fmt.Sprintf(`Option Explicit
Dim answer, n, ok
Do
  answer = InputBox("Lock after (minutes):", "Configure Lock Time", "%d")
  If answer = "" Then
    WScript.Quit 0
  End If

  ok = False
  If IsNumeric(answer) Then
    n = CLng(answer)
    If n > 0 Then ok = True
  End If

  If ok Then
    WScript.Echo CStr(n)
    WScript.Quit 0
  End If

  MsgBox "Please enter a positive integer.", vbExclamation, "Invalid value"
Loop
`, current)

	tmp, err := os.CreateTemp("", "autolock-dialog-*.vbs")
	if err != nil {
		return 0, false, err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(script); err != nil {
		_ = tmp.Close()
		return 0, false, err
	}
	if err := tmp.Close(); err != nil {
		return 0, false, err
	}

	cmd := exec.Command("cscript.exe", "//nologo", tmp.Name())
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return 0, false, err
	}
	result := strings.TrimSpace(string(out))
	if result == "" {
		return 0, false, nil
	}
	v, err := strconv.Atoi(result)
	if err != nil || v <= 0 {
		return 0, false, fmt.Errorf("invalid dialog output: %q", result)
	}
	return v, true, nil
}

func buildICO() []byte {
	const (
		w       = 16
		h       = 16
		dibSize = 40
		xorSize = w * h * 4
		andRow  = ((w + 31) / 32) * 4
		andSize = andRow * h
		imgSize = dibSize + xorSize + andSize
		dataOff = 6 + 16
	)

	buf := make([]byte, dataOff+imgSize)

	// ICONDIR
	buf[0] = 0
	buf[1] = 0
	buf[2] = 1
	buf[3] = 0
	buf[4] = 1
	buf[5] = 0

	// ICONDIRENTRY
	buf[6] = w
	buf[7] = h
	buf[8] = 0
	buf[9] = 0
	binary.LittleEndian.PutUint16(buf[10:12], 1)  // planes
	binary.LittleEndian.PutUint16(buf[12:14], 32) // bpp
	binary.LittleEndian.PutUint32(buf[14:18], uint32(imgSize))
	binary.LittleEndian.PutUint32(buf[18:22], uint32(dataOff))

	// BITMAPINFOHEADER
	off := dataOff
	binary.LittleEndian.PutUint32(buf[off+0:off+4], dibSize)
	binary.LittleEndian.PutUint32(buf[off+4:off+8], w)
	binary.LittleEndian.PutUint32(buf[off+8:off+12], h*2)
	binary.LittleEndian.PutUint16(buf[off+12:off+14], 1)
	binary.LittleEndian.PutUint16(buf[off+14:off+16], 32)
	binary.LittleEndian.PutUint32(buf[off+20:off+24], uint32(xorSize+andSize))

	// XOR bitmap (BGRA), bottom-up.
	pixelOff := off + dibSize
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := pixelOff + ((h-1-y)*w+x)*4
			buf[i+0] = 215 // B
			buf[i+1] = 120 // G
			buf[i+2] = 0   // R
			buf[i+3] = 255 // A
		}
	}

	// AND mask (all zeros => opaque)
	return buf
}
