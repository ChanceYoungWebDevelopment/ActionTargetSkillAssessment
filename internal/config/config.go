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
	Privileged   bool
	Window       int
	DownAfter    int
	PushInterval time.Duration
	WebDir       string
}

func (c Config) ListenAddr() string { return fmt.Sprintf(":%d", c.Port) }

func Parse() (Config, error) {
	var hostsRaw string
	var interval, timeout, push time.Duration
	var port, window, downAfter int
	var privileged bool
	var webdir string

	flag.StringVar(&hostsRaw, "hosts", "8.8.8.8", "comma list or @/path/to/file")
	flag.DurationVar(&interval, "interval", 2*time.Second, "probe interval")
	flag.DurationVar(&timeout, "timeout", 800*time.Millisecond, "per-probe timeout (must be < interval)")
	flag.IntVar(&port, "port", 8090, "http port")
	flag.BoolVar(&privileged, "privileged", false, "use raw ICMP (needs CAP_NET_RAW)")
	flag.IntVar(&window, "window", 120, "rolling window size (samples)")
	flag.IntVar(&downAfter, "down-after", 3, "consecutive failures to mark DOWN")
	flag.DurationVar(&push, "push-interval", 1*time.Second, "UI/SSE push interval")
	flag.StringVar(&webdir, "web-dir", "", "static assets directory")

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
	if push <= 0 {
		push = time.Second
	}

	return Config{
		Hosts:        hosts,
		Interval:     interval,
		Timeout:      timeout,
		Port:         port,
		Privileged:   privileged,
		Window:       window,
		DownAfter:    downAfter,
		//PushInterval: push,
		//WebDir:       webdir,
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
