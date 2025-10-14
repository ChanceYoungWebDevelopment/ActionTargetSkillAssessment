# ActionTarget Skill Assessment

A Go-based Linux console application that pings specified hosts, tracks uptime metrics, and serves real-time data via a local web dashboard.

---

## Features
- Command-line flags for `--hosts`, `--port`, and `--interval`
- Real-time latency and packet-loss monitoring
- Embedded Chart.js web dashboard (dark mode)
- Optional systemd service for daemon operation

---

## Quick Start
1. Clone the repository:
   ```bash
   git clone https://github.com/ChanceYoungWebDevelopment/ActionTargetSkillAssessment.git
   cd ActionTargetSkillAssessment
2. Run Locally with:
    ```bash
    make run

## Installation as a Service
```bash
make build
sudo make install
sudo systemctl enable at-ping
sudo systemctl start at-ping

