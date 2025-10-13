let chart;

async function render(data) {
  const hostDiv = document.getElementById('hosts');
  hostDiv.innerHTML = data.hosts.map(h =>
    `<div><strong>${h.host}</strong> — ${h.up ? '🟢 UP' : '🔴 DOWN'} — loss ${h.loss_pct.toFixed(1)}% — avg ${h.avg_rtt_ms.toFixed(1)} ms</div>`
  ).join('');

  // Simple: show first host’s latency sparkline
  if (data.hosts.length === 0) return;
  const h = data.hosts[0];
  const xs = h.samples.map(s => '');
  const ys = h.samples.map(s => s.ok ? s.rtt_ms || 0 : null);

  const ctx = document.getElementById('latency').getContext('2d');
  if (!chart) {
    chart = new Chart(ctx, {
      type: 'line',
      data: { labels: xs, datasets: [{ label: h.host, data: ys, spanGaps: true }] },
      options: { animation: false, responsive: true, plugins: { legend: { display: true } } }
    });
  } else {
    chart.data.labels = xs;
    chart.data.datasets[0].label = h.host;
    chart.data.datasets[0].data = ys;
    chart.update('none');
  }
}

async function init() {
  const snap = await fetch('/api/snapshot').then(r => r.json());
  render(snap);

  const es = new EventSource('/api/stream');
  es.addEventListener('update', (ev) => render(JSON.parse(ev.data)));
}
init();
