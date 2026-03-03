package radio

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"waveloggate/internal/config"
	"waveloggate/internal/wavelog"
)

// StatusCallback is called whenever radio status changes or needs reporting.
type StatusCallback func(status RigStatus)

// Poller polls a RadioClient every second and reports changes.
type Poller struct {
	mu          sync.Mutex
	client      RadioClient
	cfg         *config.Profile
	wlClient    *wavelog.Client
	onStatus    StatusCallback
	lastStatus  RigStatus
	lastUpdated time.Time
	cancel      context.CancelFunc
}

// NewPoller creates a Poller. onStatus is called whenever status changes (or 30 min force update).
func NewPoller(cfg *config.Profile, wlClient *wavelog.Client, onStatus StatusCallback) *Poller {
	return &Poller{
		cfg:      cfg,
		wlClient: wlClient,
		onStatus: onStatus,
	}
}

// UpdateConfig updates the profile and radio client.
func (p *Poller) UpdateConfig(cfg *config.Profile) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cfg = cfg
	p.client = buildClient(cfg)
}

// Start begins polling in a background goroutine.
func (p *Poller) Start(ctx context.Context) {
	p.mu.Lock()
	p.client = buildClient(p.cfg)
	p.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			p.poll()
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}
		}
	}()
}

// Stop stops the poller.
func (p *Poller) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}

// SetFreqMode issues a QSY command through the current radio client.
func (p *Poller) SetFreqMode(hz int64, mode string) error {
	p.mu.Lock()
	client := p.client
	cfg := p.cfg
	p.mu.Unlock()

	if client == nil {
		return fmt.Errorf("no radio client configured")
	}

	modes, _ := client.GetModes()
	targetMode := SelectMode(mode, hz, modes)

	setMode := ""
	if cfg.WavelogPmode && targetMode != "" {
		setMode = targetMode
	}

	return client.SetFreqMode(hz, setMode)
}

// CurrentStatus returns the last known radio status.
func (p *Poller) CurrentStatus() RigStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastStatus
}

func (p *Poller) poll() {
	p.mu.Lock()
	client := p.client
	cfg := p.cfg
	p.mu.Unlock()

	if client == nil {
		return
	}

	status, err := client.GetStatus()
	if err != nil {
		return
	}

	// Optionally zero out power.
	if cfg.IgnorePwr {
		status.Power = 0
	}

	p.mu.Lock()
	changed := !statusEqual(status, p.lastStatus)
	forceUpdate := time.Since(p.lastUpdated) > 30*time.Minute
	if changed || forceUpdate {
		p.lastStatus = status
		p.lastUpdated = time.Now()
		p.mu.Unlock()

		if p.onStatus != nil {
			p.onStatus(status)
		}

		// Send to Wavelog.
		if p.wlClient != nil {
			data := wavelog.RadioData{
				Radio:     cfg.WavelogRadioname,
				Key:       cfg.WavelogKey,
				Frequency: int64(math.Round(status.FreqA)),
				Mode:      status.Mode,
				Power:     status.Power,
				Split:     status.Split,
			}
			if status.Split {
				data.FrequencyRx = int64(math.Round(status.FreqB))
				data.ModeRx = status.ModeB
			}
			_ = p.wlClient.UpdateRadioStatus(data)
		}
	} else {
		p.mu.Unlock()
	}
}

func statusEqual(a, b RigStatus) bool {
	return a.FreqA == b.FreqA &&
		a.FreqB == b.FreqB &&
		a.Mode == b.Mode &&
		a.ModeB == b.ModeB &&
		a.Power == b.Power &&
		a.Split == b.Split &&
		a.PTT == b.PTT
}

func buildClient(cfg *config.Profile) RadioClient {
	if cfg == nil {
		return nil
	}
	switch {
	case cfg.FlrigEna:
		return NewFLRig(cfg.FlrigHost, cfg.FlrigPort)
	case cfg.HamlibEna:
		return NewHamlib(cfg.HamlibHost, cfg.HamlibPort)
	default:
		return nil
	}
}
