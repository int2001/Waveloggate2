package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Profile holds per-profile configuration.
type Profile struct {
	WavelogURL       string  `json:"wavelog_url"`
	WavelogKey       string  `json:"wavelog_key"`
	WavelogID        string  `json:"wavelog_id"`
	WavelogRadioname string  `json:"wavelog_radioname"`
	WavelogPmode     bool    `json:"wavelog_pmode"`
	FlrigHost        string  `json:"flrig_host"`
	FlrigPort        string  `json:"flrig_port"`
	FlrigEna         bool    `json:"flrig_ena"`
	HamlibHost       string  `json:"hamlib_host"`
	HamlibPort       string  `json:"hamlib_port"`
	HamlibEna        bool    `json:"hamlib_ena"`
	IgnorePwr          bool    `json:"ignore_pwr"`
	RotatorEnabled     bool    `json:"rotator_enabled"`
	RotatorHost        string  `json:"rotator_host"`
	RotatorPort        string  `json:"rotator_port"`
	RotatorThresholdAz float64 `json:"rotator_threshold_az"`
	RotatorThresholdEl float64 `json:"rotator_threshold_el"`
	RotatorParkAz      float64 `json:"rotator_park_az"`
	RotatorParkEl      float64 `json:"rotator_park_el"`
}

// Config is the root configuration object.
type Config struct {
	Version      int       `json:"version"`
	Profile      int       `json:"profile"`
	ProfileNames []string  `json:"profileNames"`
	UDPEnabled   bool      `json:"udp_enabled"`
	UDPPort      int       `json:"udp_port"`
	Profiles     []Profile `json:"profiles"`
}

func defaultProfile() Profile {
	return Profile{
		WavelogURL:       "",
		WavelogKey:       "",
		WavelogID:        "0",
		WavelogRadioname: "WLGate",
		WavelogPmode:     true,
		FlrigHost:        "127.0.0.1",
		FlrigPort:        "12345",
		FlrigEna:         false,
		HamlibHost:       "127.0.0.1",
		HamlibPort:       "4532",
		HamlibEna:        false,
		IgnorePwr:          false,
		RotatorHost:        "",
		RotatorPort:        "4533",
		RotatorThresholdAz: 2,
		RotatorThresholdEl: 2,
		RotatorParkAz:      0,
		RotatorParkEl:      0,
	}
}

func Default() Config {
	return Config{
		Version:      3,
		Profile:      0,
		ProfileNames: []string{"Profile 1", "Profile 2"},
		UDPEnabled:   true,
		UDPPort:      2333,
		Profiles:     []Profile{defaultProfile(), defaultProfile()},
	}
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "WavelogGate", "config.json"), nil
}

func Load() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Default(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Default()
			_ = Save(cfg)
			return cfg, nil
		}
		return Default(), err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), err
	}

	cfg = migrate(cfg)
	return cfg, nil
}

func Save(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// migrate ensures the config matches version 3 schema.
func migrate(cfg Config) Config {
	// Ensure at least 2 profiles exist.
	for len(cfg.Profiles) < 2 {
		cfg.Profiles = append(cfg.Profiles, defaultProfile())
	}
	// Ensure profileNames match profile count.
	for len(cfg.ProfileNames) < len(cfg.Profiles) {
		cfg.ProfileNames = append(cfg.ProfileNames, defaultProfileName(len(cfg.ProfileNames)))
	}
	// Version upgrades.
	if cfg.Version < 3 {
		cfg.Version = 3
		if cfg.UDPPort == 0 {
			cfg.UDPPort = 2333
		}
		cfg.UDPEnabled = true
	}
	return cfg
}

func defaultProfileName(idx int) string {
	return fmt.Sprintf("Profile %d", idx+1)
}

// ActiveProfile returns the currently active profile.
func (c *Config) ActiveProfile() Profile {
	if c.Profile >= 0 && c.Profile < len(c.Profiles) {
		return c.Profiles[c.Profile]
	}
	return defaultProfile()
}
