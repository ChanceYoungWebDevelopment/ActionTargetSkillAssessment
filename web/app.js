let latencyChart, lossChart;
let state = { hosts: [], selectedIndex: 0, lastUpdated: null };
let hostOrder = [];

// Configure Chart.js to respect dark mode from CSS
Chart.defaults.color = getComputedStyle(document.documentElement).getPropertyValue('--text').trim();
Chart.defaults.borderColor = getComputedStyle(document.documentElement).getPropertyValue('--border').trim();
Chart.defaults.font.family = 'ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Arial';

const lineColor = '#4da6ff';

function fmtPct(n){ return `${n.toFixed(1)}%`; }
function fmtMs(n){ return `${n.toFixed(1)} ms`; }
function nowStamp(){ return new Date().toLocaleString(); }

function movingAverage(arr, window=10){
  const out = new Array(arr.length).fill(null);
  let sum = 0, count = 0;
  for (let i=0;i<arr.length;i++){
    sum += arr[i]; count++;
    if (i >= window){ sum -= arr[i-window]; count--; }
    out[i] = (count>0) ? (sum / count) : null;
  }
  return out;
}

function seriesFromSamples(samples){
  const labels = samples.map((_, i) => String(i+1));
  const latency = samples.map(s => (s && s.ok) ? (s.rtt_ms || 0) : null);
  const loss01  = samples.map(s => (s && s.ok) ? 0 : 1);
  const rollPct = movingAverage(loss01, 10).map(v => v==null ? null : v*100);
  return { labels, latency, loss01, rollPct };
}

function renderHostList(hosts){
  const list = document.getElementById('hosts');
  list.innerHTML = hosts.map((h, i) => {
    const statusDot = `<span class="dot ${h.up ? 'up' : 'down'}"></span>`;
    const loss = fmtPct(h.loss_pct ?? 0);
    const avg  = fmtMs(h.avg_rtt_ms ?? 0);
    return `
      <button class="host ${i===state.selectedIndex?'active':''}" data-index="${i}" title="Click to view charts">
        <div>
          <div class="name">${statusDot}${h.host}</div>
          <div class="meta">loss ${loss} • avg ${avg}</div>
        </div>
      </button>
    `;
  }).join('');

  list.querySelectorAll('.host').forEach(btn => {
    btn.addEventListener('click', (e) => {
      const idx = Number(e.currentTarget.dataset.index);
      if (idx !== state.selectedIndex){
        state.selectedIndex = idx;
        updateSelectedUI();
        renderCharts();
      }
    });
  });
}

function updateSelectedUI(){
  document.querySelectorAll('.host').forEach((el, i) => {
    el.classList.toggle('active', i === state.selectedIndex);
  });
  const h = state.hosts[state.selectedIndex];
  const title = document.getElementById('host-title');
  if (h){
    title.textContent = `${h.host} — ${h.up ? '🟢 UP' : '🔴 DOWN'} — loss ${fmtPct(h.loss_pct ?? 0)} — avg ${fmtMs(h.avg_rtt_ms ?? 0)}`;
  } else {
    title.textContent = 'Select a host';
  }

  const pill = document.getElementById('summary-pill');
  const upCt = state.hosts.filter(x => x.up).length;
  pill.textContent = `${upCt}/${state.hosts.length} up • ${state.lastUpdated ? 'updated ' + state.lastUpdated : ''}`;
}

function ensureCharts(){
  if (!latencyChart){
    const ctx = document.getElementById('latency').getContext('2d');
    latencyChart = new Chart(ctx, {
      type: 'line',
data: { labels: [], datasets: [{
  label: 'RTT (ms)',
  data: [],
  borderColor: lineColor,
  backgroundColor: lineColor + '33', // translucent fill under line
  spanGaps: true,
  pointRadius: 0,
  borderWidth: 2,
  tension: 0.25
}]},
      options: {
        animation: false,
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          x: { grid: { display: false }, ticks: { display: false } },
          y: { beginAtZero: true }
        },
        plugins: { legend: { display: false } }
      }
    });
  }
  if (!lossChart){
    const ctx2 = document.getElementById('loss').getContext('2d');
    lossChart = new Chart(ctx2, {
      data: {
        labels: [],
        datasets: [
          {
            type: 'bar',
            label: 'Per-sample loss (0/1)',
            data: [],
            borderWidth: 0,
            borderSkipped: false
          },
          {
            type: 'line',
            label: 'Rolling loss (10-sample %) ',
            data: [],
            spanGaps: true,
            pointRadius: 0,
            tension: 0.25,
            borderColor: '#00e0ff',
            backgroundColor: '#00e0ff33',
            borderWidth: 2,
          }
        ]
      },
      options: {
        animation: false,
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          x: { grid: { display: false }, ticks: { display: false } },
          y: { beginAtZero: true, suggestedMax: 100 }
        },
        plugins: { legend: { display: true } }
      }
    });
  }
}

function renderCharts(){
  const host = state.hosts[state.selectedIndex];
  if (!host || !host.samples) return;

  const { labels, latency, loss01, rollPct } = seriesFromSamples(host.samples);

  ensureCharts();

  // Latency chart
  latencyChart.data.labels = labels;
  latencyChart.data.datasets[0].data = latency;
  const yMax = niceMaxFromData(latency);
latencyChart.options.scales.y = {
  beginAtZero: true,           // keeps zero on chart; remove if you want tight fit
  suggestedMax: yMax,          // dynamic ceiling from data
  ticks: { precision: 0 }
};
  latencyChart.update('none');


  // Loss chart
  lossChart.data.labels = labels;
  lossChart.data.datasets[0].data = loss01.map(v => v*100); // show as 0 or 100 for visibility
  lossChart.data.datasets[1].data = rollPct;                // smoothed %
  lossChart.update('none');

  document.getElementById('last-updated').textContent = `Last updated: ${state.lastUpdated || nowStamp()}`;
}

function renderAll(data){
  const ordered = (data.hosts || []).slice().sort((a,b) => a.host.localeCompare(b.host));
  state.hosts = ordered
  // keep selection stable; if the selected index is out-of-range, fallback to 0
  if (state.selectedIndex >= state.hosts.length) state.selectedIndex = 0;
  state.lastUpdated = nowStamp();

  renderHostList(state.hosts);
  updateSelectedUI();
  renderCharts();
}

async function init(){
  try {
    const snap = await fetch('/api/snapshot').then(r => r.json());
    renderAll(snap);
  } catch (e){
    console.error('Failed to load snapshot', e);
  }

  // Live updates
  try {
    const es = new EventSource('/api/stream');
    es.addEventListener('update', (ev) => {
      const payload = JSON.parse(ev.data);
      renderAll(payload);
    });
    es.onerror = (e) => console.warn('EventSource error', e);
  } catch (e){
    console.warn('Stream unavailable', e);
  }
}

function niceMaxFromData(values) {
  const nums = values.filter(v => Number.isFinite(v));
  if (!nums.length) return 100;
  // ignore big spikes: use ~95th percentile
  const sorted = nums.slice().sort((a,b)=>a-b);
  const p95 = sorted[Math.floor(sorted.length * 0.95)];
  const base = Math.max(5, p95);
  // round up to a “nice” step
  const steps = [5, 10, 20, 25, 50, 100, 200, 500, 1000];
  const step = steps.find(s => base / s <= 8) || 1000;
  return Math.ceil((base * 1.1) / step) * step; // +10% headroom
}

init();
