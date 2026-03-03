# AutoLock Session Timer

AutoLock Session Timer is a lightweight Windows tray app that locks the current user session after a configured number of minutes since the last session logon/unlock.

## Features

- Runs in system tray (no main window).
- Lock timer is based on session events, not keyboard/mouse idle.
- Single instance only (double-click will not start another instance).
- Configure lock minutes from tray menu.
- Live tray tooltip:
  - `Auto Lock Session Timer - X minutes since last logon`
- Optional auto-start at Windows sign-in (HKCU Run key).

## Requirements

- Windows 10 or Windows 11
- Go 1.21+ (for building from source)

## Build

```powershell
go build -ldflags="-H=windowsgui" -o AutoLock.exe
```

## Run

```powershell
.\AutoLock.exe
```

Or use the helper script:

```powershell
.\build_run.bat
```

## Tray Menu

- `Start`: enable auto lock.
- `Stop`: disable auto lock.
- `Configure Lock Time`: open input dialog and set lock minutes.
  - Save/confirm applies the new value immediately and writes to config.
  - Cancel keeps current value.
- `Exit`: quit app.

## Config

Config file path is in the same folder as `AutoLock.exe`:

- `config.json`

Example:

```json
{
  "lock_minutes": 15,
  "enabled": true
}
```

Notes:

- If `config.json` does not exist, app runs with defaults (`lock_minutes=15`, `enabled=true`).
- `config.json` is created/updated when you save settings from tray.

## Runtime Logic

- On session unlock/logon: reset timer to 0 and start counting (if enabled).
- On session lock: stop counting.
- When elapsed time reaches `lock_minutes`: call `LockWorkStation()`.

## Single Instance

App uses a named Windows mutex (`Local\\AutoLockSessionTimer`) to ensure only one process is active.

## Installer

Inno Setup script is included:

- `installer\autolock.iss`

Build `AutoLock.exe` first, then compile the `.iss` script to create an installer/uninstaller package.

## Project Files

- `main_windows.go`: app startup and goroutines.
- `session_windows.go`: Windows session event listener.
- `timer_windows.go`: timer state and lock trigger.
- `tray_windows.go`: tray UI, config dialog, tooltip updates.
- `singleinstance_windows.go`: mutex guard.
- `config_windows.go`: load/save config.

## Troubleshooting

- Tray icon not visible:
  - Check hidden tray icons in Windows taskbar.
- Config changes not applied:
  - Use tray `Configure Lock Time` and confirm value > 0.
- Multiple instances appear:
  - End all `AutoLock.exe` in Task Manager, then start once.
