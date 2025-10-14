package monitor

import (
	"context"
	"log"
	"time"

	"github.com/ChanceYoungWebDevelopment/ActionTargetSkillAssessment/internal/config"
	"github.com/ChanceYoungWebDevelopment/ActionTargetSkillAssessment/internal/metrics"
	"github.com/go-ping/ping"
)

type Manager struct {
	m       map[string]*metrics.HostMetrics
	window  int
	downAfter int
}

func NewManager(window, downAfter int) *Manager {
	return &Manager{m: make(map[string]*metrics.HostMetrics), window: window, downAfter: downAfter}
}

func (mgr *Manager) Start(ctx context.Context, cfg config.Config) error {
	for _, h := range cfg.Hosts {
		h := h
		m := metrics.NewHostMetrics(h, mgr.window)
		mgr.m[h] = m
		go mgr.worker(ctx, h, m, cfg)
	}
	return nil
}

func (mgr *Manager) worker(ctx context.Context, host string, m *metrics.HostMetrics, cfg config.Config) {
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rtt, ok := probeOnce(host, cfg.Timeout, cfg.Privileged)
			if !ok {
				m.Add(metrics.Sample{T: time.Now(), Success: false}, mgr.downAfter)
			} else {
				m.Add(metrics.Sample{T: time.Now(), Success: true, RTT: rtt}, mgr.downAfter)
			}
		}
	}
}

var initErrCount = make(map[string]int)

func probeOnce(target string, timeout time.Duration, privileged bool) (time.Duration, bool) {
p, err := ping.NewPinger(target)
if err != nil{
	if initErrCount[target] < 3 {
		log.Printf("[%s] probe error (pinger init): %v", target, err)
		initErrCount[target]++
	}

    return 0, false
}
if privileged {
    p.SetPrivileged(true) // use raw ICMP
} else {
	p.SetPrivileged(false)
	p.SetNetwork("udp4")
}

p.Count = 1
p.Timeout = timeout

if err := p.Run(); err != nil {
    log.Printf("probe run error: %v", err)
    return 0, false
}
initErrCount[target] = 0
st := p.Statistics()
if st.PacketsRecv == 1 {
    return st.AvgRtt, true
}

return 0, false

}

type Snapshot struct {
	GeneratedAt int64               `json:"generated_at"`
	Hosts       []metrics.Snapshot  `json:"hosts"`
}

func (mgr *Manager) SnapshotAll() Snapshot {
	out := make([]metrics.Snapshot, 0, len(mgr.m))
	for _, m := range mgr.m {
		out = append(out, m.Snapshot())
	}
	return Snapshot{
		GeneratedAt: time.Now().Unix(),
		Hosts:       out,
	}
}
