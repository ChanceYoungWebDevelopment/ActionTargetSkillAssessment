package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Hosts        []string
	Interval     time.Duration
	Timeout      time.Duration
	Port         int
	Window       int
	DownAfter    int
}

func (c Config) ListenAddr() string { return fmt.Sprintf(":%d", c.Port) }

func Parse() (Config, error) {
	var hostsRaw string
	var interval, timeout time.Duration
	var port, window, downAfter int

	flag.StringVar(&hostsRaw, "hosts", "1.1.1.1,example.com,192.0.2.1", "comma list or @/path/to/file")
	flag.DurationVar(&interval, "interval", 1*time.Second, "probe interval")
	flag.DurationVar(&timeout, "timeout", 800*time.Millisecond, "per-probe timeout (must be < interval)")
	flag.IntVar(&port, "port", 8090, "http port")
	flag.IntVar(&window, "window", 120, "rolling window size (samples)")
	flag.IntVar(&downAfter, "down-after", 3, "consecutive failures to mark DOWN")

	flag.Parse()

	hosts, err := resolveHosts(hostsRaw)
	if err != nil {
		return Config{}, err
	}
	if interval < 500*time.Millisecond || interval > time.Minute {
		return Config{}, errors.New("interval must be between 500ms and 1m")
	}
	if timeout >= interval {
		return Config{}, errors.New("timeout must be < interval")
	}
	if window < 30 {
		return Config{}, errors.New("window must be >= 30")
	}
	if downAfter < 1 {
		return Config{}, errors.New("down-after must be >= 1")
	}


	return Config{
		Hosts:        hosts,
		Interval:     interval,
		Timeout:      timeout,
		Port:         port,
		Window:       window,
		DownAfter:    downAfter,
	}, nil
}

func resolveHosts(raw string) ([]string, error) {
	if strings.HasPrefix(raw, "@") {
		path := strings.TrimPrefix(raw, "@")
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(b), "\n")
		var out []string
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" || strings.HasPrefix(l, "#") { continue }
			out = append(out, l)
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("no hosts in %s", path)
		}
		return out, nil
	}
	parts := strings.Split(raw, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" { out = append(out, p) }
	}
	if len(out) == 0 {
		return nil, errors.New("no hosts provided")
	}
	return out, nil
}
