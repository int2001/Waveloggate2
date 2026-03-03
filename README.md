# WaveLogGate

WaveLogGate is a desktop gateway application that connects amateur radio logging software (WSJT-X, FLDigi) and radio control hardware (FLRig, Hamlib) to the [WaveLog](https://github.com/wavelog/wavelog) web logging platform.

Built with Go + [Wails v2](https://wails.io) + Svelte. Ships as a single self-contained binary — no runtime dependencies.

---

## User Manual

### Network ports

| Port | Protocol | Direction | Purpose |
|------|----------|-----------|---------|
| 2333 (configurable) | UDP | Inbound | QSO log packets from WSJT-X / FLDigi |
| 54321 | HTTP | Inbound | QSY requests from WaveLog (`GET /{freq}/{mode}`) |
| 54322 | WebSocket | Outbound | Live radio status broadcast |

If any of these ports is already in use, a message is shown in the Status tab. Stop the conflicting application and restart WaveLogGate.

---

### Configuration tab

#### WaveLog

| Field | Description |
|-------|-------------|
| URL | Full WaveLog URL including `index.php`, e.g. `https://log.example.com/index.php` |
| API Key | WaveLog API key (found in WaveLog → Settings) |
| Station | Station profile dropdown — populated automatically from WaveLog after entering URL and key |
| Radio name | Name sent with radio status updates (default: `WLGate`) |

Press **↻** to reload the station list without leaving the field.

#### Radio Control

Select the backend that matches your setup:

| Type | Description |
|------|-------------|
| None | No radio control — WaveLogGate only forwards UDP log entries |
| FLRig | Connects to a running FLRig instance via XML-RPC |
| Hamlib | Connects to a running `rigctld` daemon via TCP |

Enter the **Host** and **Port** for the chosen backend. Defaults are `127.0.0.1:12345` (FLRig) and `127.0.0.1:4532` (Hamlib).

**Set MODE on QSY** — when WaveLog sends a QSY request, also change the radio mode (LSB below 8 MHz, USB above).

**Ignore Power** (Hamlib only) — skip reading TX power, useful for rigs where Hamlib reports power unreliably.

#### Buttons

| Button | Action |
|--------|--------|
| 💾 Save | Save the current profile settings to disk |
| Profiles | Open the profile manager (create / rename / delete / switch) |
| Test | Send a demo QSO to WaveLog's dry-run endpoint to verify connectivity |
| ⚙ Advanced | Configure UDP port and enable/disable the UDP listener |
| Quit | Exit the application |

---

### Profiles

WaveLogGate supports multiple named configuration profiles. A minimum of two profiles must exist at all times.

- **Switch** — activates the selected profile; radio poller and WaveLog client switch immediately.
- **Rename** — change the display name of any profile.
- **Add** — create a new profile with default (empty) settings.
- **Delete** — remove a profile (disabled when only two remain or for the active profile).

Unsaved field changes are lost when switching profiles — save first if needed.

---

### Status tab

- **TRX display** — shows the current frequency and mode polled from the radio (updates every second).
- **Status messages** — UDP listener startup confirmation, errors, etc.
- **QSO result** — shows a green alert on successful WaveLog submission, red on failure, with callsign / band / mode details.

---

### UDP Logger Setup (WSJT-X / FLDigi)

#### WSJT-X

1. Open **WSJT-X → File → Settings → Reporting**
2. Enable **Secondary UDP Server**
3. Set **Server name**: `localhost` (or the WaveLogGate machine IP)
4. Set **Server port**: `2333`

> Use **Secondary UDP Server** only — the primary server sends binary protocol packets that WaveLogGate does not handle.

#### FLDigi

1. Open **FLDigi → Configure → User Interface → Logging**
2. Enable **UDP** log output
3. Set host to `localhost` and port to `2333`

---

### Radio Control Setup

#### FLRig

1. Install and launch [FLRig](http://www.w1hkj.com/), configure it for your radio.
2. FLRig's XML-RPC server runs on port **12345** by default — no additional setup needed.
3. In WaveLogGate, set Radio type to **FLRig**, host `127.0.0.1`, port `12345`, and save.

#### Hamlib (rigctld)

Start `rigctld` for your radio, for example:

```bash
# Icom IC-7300 on USB serial
rigctld -m 3073 -r /dev/ttyUSB0 -s 115200 -t 4532

# Kenwood TS-2000
rigctld -m 2 -r /dev/ttyUSB0 -s 4800 -t 4532
```

Find your radio's model number with `rigctl -l`. In WaveLogGate, set Radio type to **Hamlib**, host `127.0.0.1`, port `4532`, and save.

---

### WebSocket broadcast

Any client can connect to `ws://localhost:54322` to receive live radio status:

```json
{
  "type": "radio_status",
  "frequency": 14225000,
  "mode": "USB",
  "power": 100,
  "radio": "WLGate",
  "timestamp": 1700000000000
}
```

A `{"type":"welcome","message":"..."}` message is sent on connect, followed immediately by the last known radio status.

---

### Troubleshooting

**Port conflict** — another application is using port 2333, 54321, or 54322. Find it with `lsof -i :<port>` (macOS/Linux) or `netstat -ano | findstr :<port>` (Windows) and stop it.

**Station dropdown empty** — check that the WaveLog URL (including `index.php`) and API key are correct, then press ↻.

**Test returns "wrong URL"** — the URL points to a page that returns HTML instead of JSON. Ensure the path ends with `index.php`.

**No QSOs appearing** — in WSJT-X, make sure you're using the **Secondary** UDP server, not the primary one.

**macOS quarantine (Apple Silicon)** — if the app is blocked after download, run:
```bash
xattr -d com.apple.quarantine /Applications/WaveLogGate.app
```

---

## Building from Source

### Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.23+ | https://go.dev/dl/ |
| Wails CLI | v2.x | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| Bun | any | https://bun.sh |

### Development (live reload)

```bash
cd WaveLogGate-Go
wails dev
```

This starts a live-reload server — Go and Svelte changes are picked up automatically.

### Production build

```bash
wails build
```

The binary is placed in `build/bin/`. On macOS it produces a `.app` bundle, on Windows an `.exe`, on Linux a standalone binary.

### Build flags

| Flag | Effect |
|------|--------|
| `-clean` | Clean build cache before building |
| `-platform windows/amd64` | Cross-compile for Windows |
| `-nsis` | Generate Windows NSIS installer (requires NSIS) |
| `-upx` | Compress binary with UPX |

Example:
```bash
wails build -clean -platform darwin/arm64
```
