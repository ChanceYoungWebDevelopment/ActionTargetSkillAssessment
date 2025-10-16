let latencyChart, lossChart;
let state = { hosts: [], selectedIndex: 0, lastUpdated: null };
let hostOrder = [];

// ---------- Chart.js dark-mode defaults ----------
Chart.defaults.color = getComputedStyle(document.documentElement).getPropertyValue('--text').trim();
Chart.defaults.borderColor = getComputedStyle(document.documentElement).getPropertyValue('--border').trim();
Chart.defaults.font.family = 'ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Arial';

const lineColor = '#4da6ff';

// ---------- Formatting helpers ----------
function fmtPct(n){ return `${n.toFixed(1)}%`; }
function fmtMs(n){ return `${n.toFixed(1)} ms`; }
function nowStamp(){ return new Date().toLocaleString(); }

// ---------- Math helpers ----------
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

// ---------- API origin discovery ----------
const META_ORIGIN = document.querySelector('meta[name="api-origin"]')?.content?.trim();
const CACHE_KEY = 'atping.apiOrigin.v1';

async function fetchWithTimeout(url, ms = 1200) {
  const ctrl = new AbortController();
  const t = setTimeout(() => ctrl.abort(), ms);
  try { return await fetch(url, { signal: ctrl.signal, cache: 'no-store' }); }
  finally { clearTimeout(t); }
}

async function discoverApiOrigin() {
  if (META_ORIGIN) return META_ORIGIN.replace(/\/+$/, '');
  const cached = localStorage.getItem(CACHE_KEY);
  if (cached) return cached;

  const origin = location.origin && location.origin !== 'null' ? location.origin : null;
  const candidates = new Set();
  if (origin?.startsWith('http')) candidates.add(origin);

  const hosts = [];
  if (origin) {
    const u = new URL(origin);
    hosts.push(u.hostname);
  } else {
    hosts.push('localhost','127.0.0.1');
  }
  const ports = [8090,8091,8080,3000];
  for (const h of hosts){ for (const p of ports){ candidates.add(`http://${h}:${p}`); } }
  candidates.add('http://localhost:8090');
  candidates.add('http://127.0.0.1:8090');

  for (const cand of candidates){
    try {
      const r = await fetchWithTimeout(`${cand}/api/snapshot`, 900);
      if (r.ok && r.headers.get('content-type')?.includes('application/json')){
        localStorage.setItem(CACHE_KEY, cand);
        return cand;
      }
    } catch {}
  }
  throw new Error('Could not discover API origin (no /api/snapshot responders).');
}

let API_ORIGIN = null;
const api = (p) => `${API_ORIGIN}${p.startsWith('/') ? p : '/' + p}`;
window.resetApiOrigin = () => localStorage.removeItem(CACHE_KEY);

// ---------- UI renderers ----------
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
          <div class="meta">loss avg ${loss} • avg RTT ${avg}</div>
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
    title.textContent = `${h.host} — ${h.up ? '🟢 UP' : '🔴 DOWN'} — loss avg ${fmtPct(h.loss_pct ?? 0)} — avg RTT ${fmtMs(h.avg_rtt_ms ?? 0)} — Window: ${h.samples.length}`;
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
        backgroundColor: lineColor + '33',
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
            type: 'line',
            label: 'Rolling loss',
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
        plugins: { legend: { display: false } }
      }
    });
  }
}

function renderCharts(){
  const host = state.hosts[state.selectedIndex];
  if (!host || !host.samples) return;
  const { labels, latency, loss01, rollPct } = seriesFromSamples(host.samples);
  ensureCharts();

  latencyChart.data.labels = labels;
  latencyChart.data.datasets[0].data = latency;
  const yMax = niceMaxFromData(latency);
  latencyChart.options.scales.y = {
    beginAtZero: true,
    suggestedMax: yMax,
    ticks: { precision: 0 }
  };
  latencyChart.update('none');

  lossChart.data.labels = labels;
  lossChart.data.datasets[0].data = loss01.map(v => v*100);
  lossChart.update('none');
}

function renderAll(data){
  const ordered = (data.hosts || []).slice().sort((a,b)=>a.host.localeCompare(b.host));
  state.hosts = ordered;
  if (state.selectedIndex >= state.hosts.length) state.selectedIndex = 0;
  state.lastUpdated = nowStamp();

  renderHostList(state.hosts);
  updateSelectedUI();
  renderCharts();
}

// ---------- Chart scaling helper ----------
function niceMaxFromData(values) {
  const nums = values.filter(v => Number.isFinite(v));
  if (!nums.length) return 100;
  const sorted = nums.slice().sort((a,b)=>a-b);
  const p95 = sorted[Math.floor(sorted.length * 0.95)];
  const base = Math.max(5, p95);
  const steps = [5,10,20,25,50,100,200,500,1000];
  const step = steps.find(s => base / s <= 8) || 1000;
  return Math.ceil((base * 1.1) / step) * step;
}

// ---------- Bootstrap ----------
async function init() {
  try {
    API_ORIGIN = await discoverApiOrigin();
    console.log('API origin:', API_ORIGIN);

    const snap = await fetch(api('/api/snapshot')).then(r => r.json());
    renderAll(snap);

    const es = new EventSource(api('/api/stream'));
    const handle = (e) => {
      try { renderAll(JSON.parse(e.data)); }
      catch (err) { console.error('SSE JSON parse failed:', err, e.data); }
    };
    es.onmessage = handle;
    es.addEventListener('update', handle);
    es.onerror = (e) => console.warn('SSE error (browser auto-retries):', e);
  } catch (err) {
    console.error(err);
    const el = document.getElementById('status');
    if (el) el.textContent = 'Unable to connect to API. Check server/port.';
  }
}

init();
