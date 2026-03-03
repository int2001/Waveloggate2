package rotator

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"waveloggate/internal/config"
)

const (
	busyWatchdog    = 5 * time.Second
	pollInterval    = 2 * time.Second
	pollSuppression = 3 * time.Second
	connTimeout     = 3 * time.Second
	wsRateLimit     = 150 * time.Millisecond
)

type FollowMode = string

const (
	FollowOff FollowMode = "off"
	FollowHF  FollowMode = "hf"
	FollowSAT FollowMode = "sat"
)

// Position holds azimuth and elevation angles.
type Position struct {
	Az, El float64
}

type wsCmd struct {
	az, el float64
	typ    string
}

// Client is a single-goroutine rotctld TCP client with a follow state machine.
type Client struct {
	mu         sync.Mutex
	cfg        config.Profile
	followMode FollowMode

	conn       net.Conn
	connTarget string // "host:port" last connected to
	buf        string

	currentCmd  string    // "set" | "get" | ""
	busyTimer   *time.Timer
	pendingSet  *Position // latest P command not yet sent
	pollPending bool
	hasSentP    bool // gate: don't poll until first P sent
	lastPTime   time.Time
	stopping    bool      // S sent, waiting for RPRT
	stopAfter   *Position // P to issue after S's RPRT

	lastCmdPos Position
	currentPos Position

	pendingPark bool // when true, pendingSet is a park command — bypass threshold

	lastWsCmd time.Time
	pendingWs *wsCmd
	wsTimer   *time.Timer

	OnPosition func(az, el float64)            // → Wails event rotator:position
	OnStatus   func(connected bool)             // → Wails event rotator:status
	OnBearing  func(typ string, az, el float64) // → Wails event rotator:bearing

	cmdCh  chan struct{}
	stopCh chan struct{}
}

// New creates a new Client with the given profile.
func New(cfg config.Profile) *Client {
	return &Client{
		cfg:        cfg,
		followMode: FollowOff,
		cmdCh:      make(chan struct{}, 1),
		stopCh:     make(chan struct{}),
	}
}

// Start launches the background goroutine.
func (c *Client) Start() {
	go c.run()
}

// Stop shuts down the client.
func (c *Client) Stop() {
	close(c.stopCh)
}

// UpdateProfile updates the configuration. Triggers a reconnect if host/port changed.
func (c *Client) UpdateProfile(cfg config.Profile) {
	c.mu.Lock()
	c.cfg = cfg
	c.mu.Unlock()
	c.signal()
}

// SetFollow sets the follow mode.
func (c *Client) SetFollow(mode FollowMode) {
	c.mu.Lock()
	c.followMode = mode
	if mode == FollowOff {
		c.pendingSet = nil
		if c.conn != nil && c.currentCmd == "" {
			fmt.Fprintf(c.conn, "S\n")
		}
	}
	c.mu.Unlock()
	c.signal()
}

// Park sets follow to off and queues a move to the park position, bypassing threshold.
func (c *Client) Park() {
	c.mu.Lock()
	c.followMode = FollowOff
	c.pendingSet = &Position{Az: c.cfg.RotatorParkAz, El: c.cfg.RotatorParkEl}
	c.pendingPark = true
	c.mu.Unlock()
	c.signal()
}

// HandleWSCommand handles an incoming bearing update from WS (rate-limited).
// Always fires OnBearing immediately for live display.
func (c *Client) HandleWSCommand(az, el float64, typ string) {
	// Always update bearing display.
	c.mu.Lock()
	onBearing := c.OnBearing
	followMode := c.followMode
	c.mu.Unlock()

	if onBearing != nil {
		onBearing(typ, az, el)
	}

	// Only queue rotator move if follow mode matches.
	c.mu.Lock()
	if (typ == "hf" && followMode != FollowHF) || (typ == "sat" && followMode != FollowSAT) {
		c.mu.Unlock()
		return
	}

	now := time.Now()
	if now.Sub(c.lastWsCmd) >= wsRateLimit {
		c.lastWsCmd = now
		c.pendingSet = &Position{Az: az, El: el}
		if c.wsTimer != nil {
			c.wsTimer.Stop()
			c.wsTimer = nil
		}
		c.mu.Unlock()
		c.signal()
		return
	}

	// Rate-limit: schedule deferred send.
	c.pendingWs = &wsCmd{az: az, el: el, typ: typ}
	remaining := wsRateLimit - now.Sub(c.lastWsCmd)
	if c.wsTimer == nil {
		c.wsTimer = time.AfterFunc(remaining, func() {
			c.mu.Lock()
			pw := c.pendingWs
			c.pendingWs = nil
			c.wsTimer = nil
			if pw != nil {
				c.lastWsCmd = time.Now()
				c.pendingSet = &Position{Az: pw.az, El: pw.el}
			}
			c.mu.Unlock()
			c.signal()
		})
	}
	c.mu.Unlock()
}

// GetFollowMode returns the current follow mode.
func (c *Client) GetFollowMode() FollowMode {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.followMode
}

// CurrentPosition returns the last known position.
func (c *Client) CurrentPosition() Position {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.currentPos
}

// IsConnected returns true if currently connected.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil
}

// signal wakes up the run loop without blocking.
func (c *Client) signal() {
	select {
	case c.cmdCh <- struct{}{}:
	default:
	}
}

// run is the single goroutine that owns the TCP socket.
func (c *Client) run() {
	tick := time.NewTicker(pollInterval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			c.mu.Lock()
			c.pollPending = true
			c.mu.Unlock()
			c.ensureConnected()
			c.processQueue()
		case <-c.cmdCh:
			c.ensureConnected()
			c.processQueue()
		case <-c.stopCh:
			c.mu.Lock()
			c.closeSocket()
			c.mu.Unlock()
			return
		}
	}
}

// ensureConnected dials rotctld if not already connected.
// Must be called from the run goroutine (not holding mu).
func (c *Client) ensureConnected() {
	c.mu.Lock()
	host := c.cfg.RotatorHost
	port := c.cfg.RotatorPort
	target := host + ":" + port
	conn := c.conn
	connTarget := c.connTarget
	c.mu.Unlock()

	if host == "" {
		// No host configured — disconnect if currently connected.
		c.mu.Lock()
		wasConnected := c.conn != nil
		if wasConnected {
			c.closeSocket()
		}
		onStatus := c.OnStatus
		c.mu.Unlock()
		if wasConnected && onStatus != nil {
			go onStatus(false)
		}
		return
	}

	if conn != nil && connTarget == target {
		return // already connected to correct target
	}

	// Target changed or not connected — close old socket.
	c.mu.Lock()
	c.closeSocket()
	c.mu.Unlock()

	nc, err := net.DialTimeout("tcp", target, connTimeout)
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		if c.OnStatus != nil {
			go c.OnStatus(false)
		}
		return
	}

	c.conn = nc
	c.connTarget = target
	c.buf = ""
	c.currentCmd = ""
	c.pendingSet = nil
	c.pollPending = false
	c.hasSentP = false
	c.stopping = false
	c.stopAfter = nil
	if c.busyTimer != nil {
		c.busyTimer.Stop()
		c.busyTimer = nil
	}

	if c.OnStatus != nil {
		go c.OnStatus(true)
	}

	// Start reader goroutine.
	go c.readLoop(nc)
}

// readLoop reads from the TCP socket and feeds data into onData.
// Exits when the connection is closed.
func (c *Client) readLoop(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		c.mu.Lock()
		// Only process if this connection is still current.
		if c.conn != conn {
			c.mu.Unlock()
			return
		}
		c.onLine(line)
		c.mu.Unlock()
	}
	// Connection closed.
	c.mu.Lock()
	if c.conn == conn {
		c.closeSocket()
		if c.OnStatus != nil {
			go c.OnStatus(false)
		}
	}
	c.mu.Unlock()
	c.signal()
}

// onLine processes one line received from rotctld. Must be called with mu held.
func (c *Client) onLine(line string) {
	c.buf += line + "\n"

	switch c.currentCmd {
	case "set":
		// Wait for RPRT N
		if strings.HasPrefix(line, "RPRT ") {
			c.clearBusy()
			if c.stopping {
				c.stopping = false
				if c.stopAfter != nil {
					pos := c.stopAfter
					c.stopAfter = nil
					c.sendP(pos)
				}
			}
			c.pollPending = false
			c.buf = ""
			c.signal()
		}

	case "get":
		// Accumulate lines; parse when we have ≥2 numeric lines.
		lines := strings.Split(strings.TrimSpace(c.buf), "\n")
		var nums []float64
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" {
				continue
			}
			v, err := strconv.ParseFloat(l, 64)
			if err == nil {
				nums = append(nums, v)
			}
		}
		if len(nums) >= 2 {
			c.currentPos = Position{Az: nums[0], El: nums[1]}
			if c.OnPosition != nil {
				az, el := c.currentPos.Az, c.currentPos.El
				go c.OnPosition(az, el)
			}
			c.clearBusy()
			c.buf = ""
			c.signal()
		}
	}
}

// processQueue decides what command to send next. Must be called from run goroutine.
func (c *Client) processQueue() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil || c.currentCmd != "" {
		return
	}

	if c.pendingSet != nil {
		pos := c.pendingSet
		isPark := c.pendingPark
		c.pendingSet = nil
		c.pendingPark = false

		// Check threshold (skip for explicit park commands).
		if !isPark {
			diffAz := math.Abs(pos.Az - c.lastCmdPos.Az)
			if diffAz > 180 {
				diffAz = 360 - diffAz
			}
			diffEl := math.Abs(pos.El - c.lastCmdPos.El)
			if diffAz < c.cfg.RotatorThresholdAz && diffEl < c.cfg.RotatorThresholdEl {
				return
			}
		}

		// Check direction reversal: if we're moving and need to reverse,
		// send S first.
		if c.stopping {
			// Already stopping, queue new target.
			c.stopAfter = pos
			return
		}

		// Detect direction reversal by checking if new az is on opposite side.
		needStop := false
		if c.hasSentP {
			prevDir := c.lastCmdPos.Az - c.currentPos.Az
			newDir := pos.Az - c.currentPos.Az
			if prevDir*newDir < 0 && math.Abs(prevDir) > c.cfg.RotatorThresholdAz {
				needStop = true
			}
		}

		if needStop {
			fmt.Fprintf(c.conn, "S\n")
			c.currentCmd = "set"
			c.stopping = true
			c.stopAfter = pos
			c.armBusy()
			return
		}

		c.sendP(pos)
		return
	}

	if c.pollPending && c.hasSentP && time.Since(c.lastPTime) > pollSuppression {
		c.pollPending = false
		fmt.Fprintf(c.conn, "p\n")
		c.currentCmd = "get"
		c.armBusy()
	}
}

// sendP sends a P az el command. Must be called with mu held.
func (c *Client) sendP(pos *Position) {
	if c.conn == nil {
		return
	}
	fmt.Fprintf(c.conn, "P %.1f %.1f\n", pos.Az, pos.El)
	c.lastCmdPos = *pos
	c.lastPTime = time.Now()
	c.hasSentP = true
	c.currentCmd = "set"
	c.armBusy()
}

// armBusy arms the busy watchdog timer. Must be called with mu held.
func (c *Client) armBusy() {
	if c.busyTimer != nil {
		c.busyTimer.Stop()
	}
	c.busyTimer = time.AfterFunc(busyWatchdog, func() {
		c.mu.Lock()
		c.currentCmd = ""
		c.buf = ""
		c.mu.Unlock()
		c.signal()
	})
}

// clearBusy stops the watchdog and clears currentCmd. Must be called with mu held.
func (c *Client) clearBusy() {
	if c.busyTimer != nil {
		c.busyTimer.Stop()
		c.busyTimer = nil
	}
	c.currentCmd = ""
}

// closeSocket closes the TCP connection. Must be called with mu held.
func (c *Client) closeSocket() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connTarget = ""
	c.currentCmd = ""
	c.buf = ""
	if c.busyTimer != nil {
		c.busyTimer.Stop()
		c.busyTimer = nil
	}
}
