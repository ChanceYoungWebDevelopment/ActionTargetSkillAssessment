package metrics

import (
	"sort"
	"sync"
	"time"
)

type Sample struct {
	T       time.Time
	Success bool
	RTT     time.Duration
}

type HostMetrics struct {
	Host   string
	Win    []Sample
	Head   int
	Count  int

	SumRTT time.Duration
	Succ   int

	ConsecutiveFailures int
	Up       bool
	LastFlip time.Time

	mu sync.RWMutex
}

func NewHostMetrics(host string, window int) *HostMetrics {
	m := &HostMetrics{
		Host:     host,
		Win:      make([]Sample, window),
		Up:       true,
		LastFlip: time.Now(),
		Count:    window, // pretend full window already filled
	}
	now := time.Now()
	step := -time.Second // or use cfg.Interval later
	for i := 0; i < window; i++ {
		m.Win[i] = Sample{
			T:       now.Add(time.Duration(i-window) * step),
			Success: true,
			RTT:     0, // baseline
		}
	}
	m.Succ = window
	return m
}

func (m *HostMetrics) Add(s Sample, downAfter int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Count == len(m.Win) {
		old := m.Win[m.Head]
		if old.Success {
			m.SumRTT -= old.RTT
			m.Succ--
		}
	} else {
		m.Count++
	}
	m.Win[m.Head] = s
	m.Head = (m.Head + 1) % len(m.Win)

	if s.Success {
		m.SumRTT += s.RTT
		m.Succ++
		m.ConsecutiveFailures = 0
	} else {
		m.ConsecutiveFailures++
	}

	prev := m.Up
	m.Up = m.ConsecutiveFailures < downAfter
	if m.Up != prev {
		m.LastFlip = time.Now()
	}
}

type Snapshot struct {
	Host        string         `json:"host"`
	Up          bool           `json:"up"`
	LossPct     float64        `json:"loss_pct"`
	AvgRTTms    float64        `json:"avg_rtt_ms"`
	MedianRTTms float64        `json:"median_rtt_ms"`
	Failures    int            `json:"failures"`
	LastChange  int64          `json:"last_change"`
	Samples     []SamplePublic `json:"samples"`
}

type SamplePublic struct {
	T     int64   `json:"t"`
	OK    bool    `json:"ok"`
	RTTms float64 `json:"rtt_ms,omitempty"`
}

func (m *HostMetrics) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var loss float64
	if m.Count > 0 {
		loss = 100 * float64(m.Count-m.Succ) / float64(m.Count)
	}
	var avg float64
	if m.Succ > 0 {
		avg = float64((m.SumRTT / time.Duration(m.Succ)).Microseconds()) / 1000.0
	}
	// median (optional)
	var rtts []time.Duration
	if m.Succ > 0 {
		rtts = make([]time.Duration, 0, m.Succ)
		for i := 0; i < m.Count; i++ {
			idx := (m.Head - m.Count + i + len(m.Win)) % len(m.Win)
			s := m.Win[idx]
			if s.Success { rtts = append(rtts, s.RTT) }
		}
		sort.Slice(rtts, func(i, j int) bool { return rtts[i] < rtts[j] })
	}
	var median float64
	if len(rtts) > 0 {
		med := rtts[len(rtts)/2]
		median = float64(med.Microseconds()) / 1000.0
	}

	pub := make([]SamplePublic, 0, m.Count)
	for i := 0; i < m.Count; i++ {
		idx := (m.Head - m.Count + i + len(m.Win)) % len(m.Win)
		s := m.Win[idx]
		sp := SamplePublic{T: s.T.Unix(), OK: s.Success}
		if s.Success {
			sp.RTTms = float64(s.RTT.Microseconds()) / 1000.0
		}
		pub = append(pub, sp)
	}

	return Snapshot{
		Host:        m.Host,
		Up:          m.Up,
		LossPct:     loss,
		AvgRTTms:    avg,
		MedianRTTms: median,
		Failures:    m.ConsecutiveFailures,
		LastChange:  m.LastFlip.Unix(),
		Samples:     pub,
	}
}
