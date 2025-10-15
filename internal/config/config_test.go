package config

import (
    "testing"
    "time"
	"os"
)

func TestParseDefaults(t *testing.T) {
    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"cmd"}

    cfg, err := Parse()
    if err != nil {
        t.Fatalf("Parse() failed: %v", err)
    }

    if cfg.Interval < 500*time.Millisecond {
        t.Errorf("interval too low: %v", cfg.Interval)
    }
    if len(cfg.Hosts) == 0 {
        t.Error("expected default hosts")
    }
}
