package hamlib

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"waveloggate/internal/config"
	"waveloggate/internal/debug"
)

// State represents the current lifecycle state of the managed rigctld process.
type State int

const (
	StateStopped  State = iota
	StateStarting       // process launched, waiting for TCP readiness
	StateRunning        // process is up and port is bound
	StateError          // process exited unexpectedly or failed to start
)

// Manager manages the lifecycle of a rigctld child process.
type Manager struct {
	mu        sync.Mutex
	cmd       *exec.Cmd
	state     State
	lastMsg   string
	cancelMon context.CancelFunc
	cfg       config.Profile

	// OnStatus is called on every state transition.
	// running=true means rigctld is accepting connections.
	OnStatus func(running bool, message string)
}

// New creates a new Manager. onStatus is called on every state change (may be nil).
func New(onStatus func(running bool, message string)) *Manager {
	return &Manager{OnStatus: onStatus}
}

// IsRunning returns true if rigctld is currently accepting connections.
func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state == StateRunning
}

// StatusString returns a human-readable status suitable for the frontend.
func (m *Manager) StatusString() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch m.state {
	case StateStopped:
		return "Stopped"
	case StateStarting:
		return "Starting…"
	case StateRunning:
		return "Running"
	case StateError:
		if m.lastMsg != "" {
			return "Error: " + m.lastMsg
		}
		return "Error"
	default:
		return "Unknown"
	}
}

// Validate checks that the profile has all fields required for a managed launch.
func Validate(cfg config.Profile) error {
	if cfg.HamlibModel <= 0 {
		return fmt.Errorf("no radio model selected")
	}
	if strings.TrimSpace(cfg.HamlibDevice) == "" {
		return fmt.Errorf("serial port (device) must not be empty")
	}
	if cfg.HamlibBaud <= 0 {
		return fmt.Errorf("baud rate must be greater than 0")
	}
	if p, err := strconv.Atoi(cfg.HamlibPort); err != nil || p < 1 || p > 65535 {
		return fmt.Errorf("invalid rigctld port %q", cfg.HamlibPort)
	}
	return nil
}

// Start validates the profile, finds rigctld, and launches it.
// The function returns immediately; readiness and process monitoring happen
// in a background goroutine. State changes are reported via OnStatus.
//
// IMPORTANT: no blocking OS operations are performed while m.mu is held.
// On Windows, cmd.Start() can be delayed by Defender/SmartScreen; holding
// the mutex during that delay would block GetHamlibStatus() and freeze the UI.
func (m *Manager) Start(cfg config.Profile) error {
	// Validate before touching any state.
	if err := Validate(cfg); err != nil {
		return err
	}

	// Cancel the existing monitor and grab the old process — all under lock,
	// but no blocking work happens here.
	m.mu.Lock()
	if m.cancelMon != nil {
		m.cancelMon()
		m.cancelMon = nil
	}
	oldCmd := m.cmd
	m.cmd = nil
	m.cfg = cfg
	m.mu.Unlock()

	// Terminate the old process OUTSIDE the lock (Wait can block up to 3 s).
	stopCmd(oldCmd)

	// Find the rigctld binary (filesystem stat — no lock needed).
	rigctldPath, err := RigctldPath()
	if err != nil {
		m.mu.Lock()
		m.setState(StateError, err.Error())
		m.mu.Unlock()
		return err
	}

	args := buildArgs(cfg)
	debug.Log("[HAMLIB] launching: %s %s", rigctldPath, strings.Join(args, " "))

	cmd := exec.Command(rigctldPath, args...)

	// Collect stderr lines for diagnostics.
	stderr, err := cmd.StderrPipe()
	if err != nil {
		m.mu.Lock()
		m.setState(StateError, err.Error())
		m.mu.Unlock()
		return fmt.Errorf("cannot attach stderr: %w", err)
	}

	// Prevent the child from inheriting the parent's console / WebView2 message
	// pump on Windows (see cmd_windows.go).
	setCmdAttrs(cmd)

	// cmd.Start() can be slow on Windows (Defender/SmartScreen) — keep mutex free.
	if err := cmd.Start(); err != nil {
		m.mu.Lock()
		m.setState(StateError, err.Error())
		m.mu.Unlock()
		return fmt.Errorf("cannot start rigctld: %w", err)
	}

	monCtx, cancel := context.WithCancel(context.Background())

	// Re-acquire lock only to update stored state.
	m.mu.Lock()
	m.cmd = cmd
	m.cancelMon = cancel
	m.setState(StateStarting, "Starting…")
	m.mu.Unlock()

	// Capture stderr in a separate goroutine.
	stderrLines := make(chan string, 64)
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			debug.Log("[HAMLIB stderr] %s", line)
			select {
			case stderrLines <- line:
			default:
			}
		}
		close(stderrLines)
	}()

	// Monitor goroutine: wait for readiness then watch for process exit.
	go m.monitor(monCtx, cmd, cfg, stderrLines)
	return nil
}

// Stop terminates the managed rigctld process if it is running.
// The lock is released before waiting for process exit to avoid blocking callers.
func (m *Manager) Stop() {
	m.mu.Lock()
	if m.cancelMon != nil {
		m.cancelMon()
		m.cancelMon = nil
	}
	cmd := m.cmd
	m.cmd = nil
	wasRunning := m.state != StateStopped
	m.state = StateStopped
	m.lastMsg = ""
	m.mu.Unlock()

	if wasRunning {
		m.notify(false, "")
	}

	// Terminate and wait OUTSIDE the lock (can block up to 3 s).
	stopCmd(cmd)
}

// Restart stops and then starts rigctld with the current profile config.
func (m *Manager) Restart() error {
	m.mu.Lock()
	cfg := m.cfg
	m.mu.Unlock()
	m.Stop()
	return m.Start(cfg)
}

// stopCmd terminates cmd and waits for it to exit (max 3 s). Safe to call with nil.
func stopCmd(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = terminateProcess(cmd.Process)
	done := make(chan struct{})
	go func() { cmd.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		_ = cmd.Process.Kill()
	}
}

// monitor waits for TCP readiness then watches for process exit.
func (m *Manager) monitor(ctx context.Context, cmd *exec.Cmd, cfg config.Profile, stderrLines <-chan string) {
	addr := net.JoinHostPort(cfg.HamlibHost, cfg.HamlibPort)
	if cfg.HamlibHost == "" {
		addr = net.JoinHostPort("127.0.0.1", cfg.HamlibPort)
	}

	// Wait for rigctld to start listening (5-second window).
	ready := waitForPort(ctx, addr, 5*time.Second)
	if !ready {
		// Collect any error lines already buffered.
		var stderrBuf []string
		drainLoop:
		for {
			select {
			case line, ok := <-stderrLines:
				if !ok {
					break drainLoop
				}
				stderrBuf = append(stderrBuf, line)
			default:
				break drainLoop
			}
		}
		msg := "rigctld did not start in time — check your serial port and baud rate"
		if len(stderrBuf) > 0 {
			last := stderrBuf[len(stderrBuf)-1]
			msg = interpretStderr(last, cfg)
		}
		m.mu.Lock()
		m.setState(StateError, msg)
		m.mu.Unlock()
		cmd.Process.Kill()
		return
	}

	m.mu.Lock()
	m.setState(StateRunning, "Running")
	m.mu.Unlock()

	// Drain remaining stderr lines.
	var lastStderr string
	exitCh := make(chan error, 1)
	go func() { exitCh <- cmd.Wait() }()

	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-stderrLines:
			if ok {
				lastStderr = line
			}
		case exitErr := <-exitCh:
			select {
			case <-ctx.Done():
				// Expected stop — don't report error.
				return
			default:
			}
			msg := "rigctld exited unexpectedly"
			if lastStderr != "" {
				msg = interpretStderr(lastStderr, cfg)
			} else if exitErr != nil {
				msg = exitErr.Error()
			}
			m.mu.Lock()
			m.state = StateError
			m.lastMsg = msg
			m.notify(false, msg)
			m.mu.Unlock()
			return
		}
	}
}

// setState sets state+message and calls notify. Must be called with m.mu held.
func (m *Manager) setState(s State, msg string) {
	m.state = s
	m.lastMsg = msg
	m.notify(s == StateRunning, msg)
}

// notify calls OnStatus without holding mu (to avoid deadlock if the callback
// tries to call back into the manager). Caller must hold mu when calling.
func (m *Manager) notify(running bool, msg string) {
	fn := m.OnStatus
	if fn != nil {
		go fn(running, msg)
	}
}

// buildArgs constructs the rigctld argument list from a profile.
func buildArgs(cfg config.Profile) []string {
	args := []string{
		"-m", strconv.Itoa(cfg.HamlibModel),
		"-r", cfg.HamlibDevice,
		"-s", strconv.Itoa(cfg.HamlibBaud),
		"-t", cfg.HamlibPort,
	}

	// Bind address (if non-default).
	if cfg.HamlibHost != "" && cfg.HamlibHost != "0.0.0.0" {
		args = append(args, "-T", cfg.HamlibHost)
	}

	// Parity.
	if cfg.HamlibParity != "" && cfg.HamlibParity != "none" {
		args = append(args, "-P", cfg.HamlibParity)
	}

	// Stop bits.
	if cfg.HamlibStopBits > 0 {
		args = append(args, "-S", strconv.Itoa(cfg.HamlibStopBits))
	}

	// Handshake (mapped to set-conf).
	switch cfg.HamlibHandshake {
	case "rtscts":
		args = append(args, "--set-conf=rts_state=ON,cts_state=ON")
	case "xonxoff":
		args = append(args, "--set-conf=xon_xoff=1")
	}

	return args
}

// waitForPort tries to TCP-connect to addr until it succeeds or the deadline
// is reached. Returns true if the port became available before the deadline.
func waitForPort(ctx context.Context, addr string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return false
		}
		conn, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(200 * time.Millisecond):
		}
	}
	return false
}

// interpretStderr maps common rigctld error messages to user-friendly strings.
func interpretStderr(line string, cfg config.Profile) string {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "no such device") || strings.Contains(lower, "no such file"):
		return fmt.Sprintf("Serial port %q not found — check the device name", cfg.HamlibDevice)
	case strings.Contains(lower, "permission denied"):
		return fmt.Sprintf("Permission denied on %q — try: sudo chmod a+rw %s", cfg.HamlibDevice, cfg.HamlibDevice)
	case strings.Contains(lower, "address already in use") || strings.Contains(lower, "bind"):
		return fmt.Sprintf("Port %s is already in use — another rigctld may be running", cfg.HamlibPort)
	case strings.Contains(lower, "rig not found") || strings.Contains(lower, "no rig found"):
		return fmt.Sprintf("Hamlib model %d not recognised by rigctld — check your radio model selection", cfg.HamlibModel)
	default:
		return "rigctld: " + line
	}
}
