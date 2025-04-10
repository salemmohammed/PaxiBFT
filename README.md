# PaxiBFT

**PaxiBFT** is a modular and extensible framework written in Go for implementing, benchmarking, and evaluating Byzantine Fault Tolerant (BFT) consensus protocols. It provides a unified environment for testing and comparing BFT protocols under standardized conditions.

## Overview

This framework helps researchers and engineers to:

- Implement new BFT protocols or extend existing ones.
- Benchmark different protocols using the same environment.
- Analyze performance metrics like latency, throughput, and commit rate.

## Implemented Protocols

- PBFT
- Paxos
- Tendermint
- Streamlet
- HotStuff

Each protocol is implemented as a pluggable module with a standard interface.

## Installation

Requires Go 1.17+.

```bash
git clone https://github.com/salemmohammed/PaxiBFT.git
cd PaxiBFT
go mod tidy
```

## Usage

Run a protocol with default parameters:

```bash
go run main.go -protocol=pbft -nodes=4 -f=1
```

**Flags:**
- `-protocol`: One of `pbft`, `paxos`, `hotstuff`, `streamlet`, `tendermint`
- `-nodes`: Number of nodes (e.g., 4)
- `-f`: Number of tolerated Byzantine faults

## Configuration

You can configure experiments via a YAML file in the `config/` directory:

```yaml
protocol: pbft
nodes: 4
faults: 1
message_size: 256
block_size: 10
duration: 60  # seconds
```

## Output Format

PaxiBFT produces logs with the following metrics (in `logs/` folder):

- `throughput` (tx/sec)
- `latency_avg_ms`
- `view_changes`
- `commit_rate`
- `bandwidth_usage`

Example output (`.json` or `.csv`):

```json
{
  "protocol": "PBFT",
  "nodes": 4,
  "throughput": 950,
  "latency_avg_ms": 15.2,
  "view_changes": 0,
  "duration_s": 60
}
```

## Example

To run Tendermint for 60 seconds with 4 nodes:

```bash
go run main.go -protocol=tendermint -nodes=4 -duration=60
```

Check the `logs/` directory for results.

## Publications

This framework is featured in:

- **Bottlenecks in Blockchain Consensus Protocols**  
  [IEEE Xplore](https://ieeexplore.ieee.org/document/9524210)

## Contributing

We welcome contributions! Please open an issue or submit a pull request to get started.

