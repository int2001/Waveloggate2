package wavelog

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"waveloggate/internal/config"
)

// QSOResult holds the result of a QSO submission.
type QSOResult struct {
	Success  bool
	Call     string
	Band     string
	Mode     string
	RstSent  string
	RstRcvd  string
	TimeOn   string
	Reason   string
}

// RadioData holds the data sent to WaveLog's /api/radio endpoint.
type RadioData struct {
	Radio       string
	Key         string
	Frequency   int64
	Mode        string
	Power       float64
	FrequencyRx int64
	ModeRx      string
	Split       bool
}

// Station represents a WaveLog station profile.
type Station struct {
	Name     string `json:"station_profile_name"`
	Callsign string `json:"station_callsign"`
	ID       string `json:"station_id"`
}

// Client communicates with the WaveLog API.
type Client struct {
	cfg        *config.Profile
	httpClient *http.Client
	appVersion string
}

// New creates a new WaveLog client.
func New(cfg *config.Profile, appVersion string) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
	}
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout:   5 * time.Second,
			Transport: transport,
		},
		appVersion: appVersion,
	}
}

// UpdateProfile updates the config profile used by the client.
func (c *Client) UpdateProfile(cfg *config.Profile) {
	c.cfg = cfg
}

type qsoPayload struct {
	Key              string `json:"key"`
	StationProfileID string `json:"station_profile_id"`
	Type             string `json:"type"`
	String           string `json:"string"`
}

type apiResponse struct {
	Status   string   `json:"status"`
	Reason   string   `json:"reason"`
	Messages []string `json:"messages"`
	Call     string   `json:"call"`
	Band     string   `json:"band"`
	Mode     string   `json:"mode"`
	RstSent  string   `json:"rst_sent"`
	RstRcvd  string   `json:"rst_rcvd"`
	TimeOn   string   `json:"time_on"`
}

// SendQSO posts an ADIF string to WaveLog. dryRun uses /api/qso/true.
func (c *Client) SendQSO(adifStr string, dryRun bool) (*QSOResult, error) {
	endpoint := strings.TrimRight(c.cfg.WavelogURL, "/") + "/api/qso"
	if dryRun {
		endpoint += "/true"
	}

	payload := qsoPayload{
		Key:              c.cfg.WavelogKey,
		StationProfileID: c.cfg.WavelogID,
		Type:             "adif",
		String:           adifStr,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("internet problem")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "SW2WL_v"+c.appVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
			return &QSOResult{Success: false, Reason: "timeout"}, nil
		}
		return &QSOResult{Success: false, Reason: "internet problem"}, nil
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	bodyStr := string(data)

	// Detect HTML response (wrong URL).
	if strings.Contains(bodyStr, "<html") || strings.Contains(bodyStr, "<!DOCTYPE") {
		return &QSOResult{Success: false, Reason: "wrong URL"}, nil
	}

	var ar apiResponse
	if err := json.Unmarshal(data, &ar); err != nil {
		return &QSOResult{Success: false, Reason: "invalid response"}, nil
	}

	if len(ar.Messages) > 0 {
		return &QSOResult{Success: false, Reason: strings.Join(ar.Messages, ", ")}, nil
	}

	statusLower := strings.ToLower(ar.Status)
	reasonLower := strings.ToLower(ar.Reason)
	isDryrunOK := dryRun && (strings.Contains(statusLower, "dryrun") || strings.Contains(reasonLower, "dryrun"))

	if ar.Status == "created" || isDryrunOK {
		return &QSOResult{
			Success: true,
			Call:    ar.Call,
			Band:    ar.Band,
			Mode:    ar.Mode,
			RstSent: ar.RstSent,
			RstRcvd: ar.RstRcvd,
			TimeOn:  ar.TimeOn,
		}, nil
	}

	reason := ar.Reason
	if reason == "" {
		reason = ar.Status
	}
	return &QSOResult{Success: false, Reason: reason}, nil
}

type radioPayload struct {
	Radio       string  `json:"radio"`
	Key         string  `json:"key"`
	Frequency   int64   `json:"frequency"`
	Mode        string  `json:"mode"`
	Power       float64 `json:"power,omitempty"`
	FrequencyRx int64   `json:"frequency_rx,omitempty"`
	ModeRx      string  `json:"mode_rx,omitempty"`
}

// UpdateRadioStatus posts radio status to WaveLog's /api/radio.
func (c *Client) UpdateRadioStatus(data RadioData) error {
	endpoint := strings.TrimRight(c.cfg.WavelogURL, "/") + "/api/radio"

	freq := data.Frequency
	freqRx := data.FrequencyRx
	mode := data.Mode
	modeRx := data.ModeRx

	// If split, swap TX/RX.
	if data.Split {
		freq, freqRx = freqRx, freq
		mode, modeRx = modeRx, mode
	}

	p := radioPayload{
		Radio:       c.cfg.WavelogRadioname,
		Key:         c.cfg.WavelogKey,
		Frequency:   freq,
		Mode:        mode,
		FrequencyRx: freqRx,
		ModeRx:      modeRx,
	}
	if data.Power > 0 {
		p.Power = data.Power
	}

	body, err := json.Marshal(p)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// GetStations fetches the station profile list from WaveLog.
func (c *Client) GetStations() ([]Station, error) {
	endpoint := strings.TrimRight(c.cfg.WavelogURL, "/") + "/api/station_info/" + c.cfg.WavelogKey

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "SW2WL_v"+c.appVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var stations []Station
	if err := json.Unmarshal(data, &stations); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	return stations, nil
}
