package metrics

import (
    "testing"
    "time"
)

func TestAddAndSnapshot(t *testing.T) {
    m := NewHostMetrics("example.com", 5)

    // Simulate a few successful probes
    for i := 0; i < 3; i++ {
        m.Add(Sample{T: time.Now(), Success: true, RTT: 10 * time.Millisecond}, 3)
    }

    snap := m.Snapshot()
    if snap.LossPct != 0 {
        t.Errorf("expected 0%% loss, got %.1f", snap.LossPct)
    }
    if snap.AvgRTTms < 9.9 || snap.AvgRTTms > 10.1 {
        t.Errorf("unexpected avg RTT: %.2f ms", snap.AvgRTTms)
    }
}
