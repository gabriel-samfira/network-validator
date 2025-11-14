# Network Monitoring System - Deployment Guide

A distributed network monitoring system with aggregator and agent modes for monitoring server connectivity and bonds.

## Quick Start

### 1. Build the Application

```bash
go build
```

### 2. Generate Configuration Files

```bash
# For aggregator server
./validate -generate-config aggregator

# For agent servers
./validate -generate-config agent
```

### 3. Start Aggregator

```bash
./validate -config config.toml
```

Open `http://localhost:8080` to view the dashboard.

### 4. Start Agents on Other Servers

Edit the generated `config.toml` to point to your aggregator:

```toml
mode = "agent"

[agent]
aggregator_url = "http://your-aggregator-server:8080"
register_interval = 300
```

Then run:

```bash
./validate -config config.toml
```

## Configuration Files

Example configurations are provided:
- `config.aggregator.toml` - Aggregator mode
- `config.agent.toml` - Agent mode

## API Endpoints

### Aggregator
- `GET /` - Web dashboard
- `POST /api/server` - Agent registration
- `GET /api/servers` - List all registered servers
- `GET /api/test-results` - View connectivity test results
- `POST /api/test-results` - Submit test results

### Agent
- `GET /api/sysinfo` - System information
- `POST /api/run-tests` - Run connectivity tests

## Testing Connectivity

```bash
# Get test configuration from aggregator
curl http://aggregator:8080/api/run-tests > tests.json

# Send to agent to run tests
curl -X POST http://agent:8080/api/run-tests \
  -H "Content-Type: application/json" \
  -d @tests.json
```

Results will appear in the aggregator dashboard.
