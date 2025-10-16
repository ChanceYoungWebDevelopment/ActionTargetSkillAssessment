
# ActionTarget Skill Assessment

  ## Program Overview

A Go-based Linux console application that continuously pings specified hosts, tracks uptime metrics, and serves a real-time web dashboard using embedded static assets.

### Features
- Command-line flags for `--hosts`, `--port`, and `--interval`

- Real-time latency and packet-loss monitoring

- Embedded Chart.js web dashboard (dark mode)

- Optional systemd service for daemon operation

### Requirements
- Linux system (tested on Debian 12 Bookworm)
- Go 1.22+ installed
- Network access to target hosts
- **ICMP (ping) support required**
  - The program uses ICMP Echo Requests to measure latency and packet loss.
  - This requires the `cap_net_raw` capability or root privileges.
  - The provided install script for the service automatically applies:
    ```bash
    sudo setcap cap_net_raw+ep /usr/local/bin/at-ping
    ```
   - This needs to be run after any `make build` command

- Optional: systemd (for running as a service)


### Compatibility Notes
This project was developed and tested on Debian 12 (Bookworm).
It should run on most Linux distributions with glibc-based environments.
Minimal distributions such as Alpine (musl-based) may require additional dependencies 

## Usage and Configuration

### Quick Start

1. Clone the repository:

```bash
git clone https://github.com/ChanceYoungWebDevelopment/ActionTargetSkillAssessment.git

cd ActionTargetSkillAssessment
```

2. Run Locally with:
```bash
make run
```

### Required flags
- hosts -> allows for specifying the target hosts. *Note: you can set the list of hosts with a txt file by passing @/path/to/your/txt*. Defaults to: example.com
- port -> specifies the port used by the UI dashboard. Defaults to 8090.
- interval -> specifies the interval used by the monitors for each host. *This will apply to all host probes.* Defaults to 1 second. 

### Optional Configuration Flags
- timeout -> specifies the per-probe timeout interval. Defaults to 800ms.
- window -> specifies the size of the ring buffer or data window. Defaults to 120.
- down-after -> specifies how many failed attempts count as a target being "down". Defaults to 3. 

### Installation as a Service
I have included a makefile and some installation scripts which allow you to set this up as a systemd service by running the following:
```bash
make build
sudo make install
sudo systemctl enable at-ping
sudo systemctl start at-ping
```
Once installed, visit your dashboard at:
```http://<your-server-ip>:8090```

**PLEASE** review the **Requirements** section before running this.

## Design and Implementation Notes

### Development Environment
Development was done on a **DigitalOcean Debian 12 (Bookworm)** droplet to mirror a real Linux environment.  
Early issues with CPU throttling in VS Code’s remote connection were resolved by upgrading to a 2-CPU droplet.  
A **code-server** instance was configured for remote browser-based editing to maintain progress while traveling.

### Language and Framework Selection
This was my **first Go project**, chosen to meet the requirement of building a non-Python, non-Java, non-JavaScript solution.  
Go’s native concurrency model, strong networking libraries, and easy binary deployment made it a natural fit for a lightweight, reliable monitoring daemon.

### Core Architecture
- **Config System:**  
  A central `Config` struct handles CLI flag parsing and runtime parameters (`--hosts`, `--port`, `--interval`, etc.).  
  This design allows the same binary to be used both interactively and as a `systemd` service.
  
- **Monitor Manager:**  
  The `Manager` component launches a lightweight **goroutine** for each host, using the Go-Ping library to send periodic ICMP echo requests according to the interval passed via the  `Config` object. 
  
  Each host’s results are stored in a **ring buffer** ( a fixed-size rolling window, also customizable with a `Config` flag) for efficient time-series tracking of latency and packet loss.

- **Metrics Model:**  
  Each sample records success/failure, timestamp, and round-trip time (RTT).  
  Aggregated data provides average latency, packet-loss percentage, median RTT, and consecutive failure counts.

### Web Dashboard
Visualization is handled by a **Chart.js-based dashboard** served from an embedded static file system using the *embed* module provided by Go.
 
The dashboard streams live data through **Server-Sent Events (SSE)** for real-time updates without client polling.  

Dark-mode styling and minimal layout were added for clarity and professional presentation.

### Design Rationale

-   **Ring Buffer:** Efficiently maintains a sliding window of recent metrics without unbounded memory growth.
    
-   **Goroutines:** Each host’s monitoring loop runs in its own goroutine, allowing multiple hosts to be pinged concurrently without blocking each other.
    
-   **Go-Ping Library:** Provides reliable probing and a data structure for collecting desired metrics
    
-   **Embedded Assets:** Simplify portability and deployment, especially for service-based environments.
    
-   **Systemd + Capabilities:** Align with modern Linux service practices for persistent and secure operation.

© 2025 Chance Young
> Written with [StackEdit](https://stackedit.io/).