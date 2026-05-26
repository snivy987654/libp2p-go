<div align="center">


<br/>

<img width="1024" height="572" alt="image" src="https://github.com/user-attachments/assets/51f52427-dd01-423e-9779-6cdc235074b7" />


<br/><br/>

# libp2p-go — Pre-compiled Binaries

**Production-ready, statically-compiled libp2p node binaries for Linux.**  
Drop-in peer-to-peer networking — no Go toolchain required.

<br/>

[![Go Version](https://img.shields.io/badge/Go-1.25.7-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![libp2p](https://img.shields.io/badge/go--libp2p-v0.48.0-7B3FE4?style=flat-square)](https://github.com/libp2p/go-libp2p)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](#license)
[![Platforms](https://img.shields.io/badge/Platform-Linux-orange?style=flat-square&logo=linux)]()

<br/>

```
linux/amd64  ·  linux/arm64
```

<br/>

</div>

---

## Table of Contents

- [Overview](#overview)
- [Binary Releases](#binary-releases)
- [Quick Start](#quick-start)
- [Supported Transports](#supported-transports)
- [Supported Protocols](#supported-protocols)
- [Dependencies & Modules](#dependencies--modules)
- [Verification](#verification)
- [System Requirements](#system-requirements)
- [Build Information](#build-information)
- [License](#license)

---

## Overview

This repository distributes pre-compiled binaries of a **go-libp2p** application — a full-featured peer-to-peer networking node built on the [libp2p](https://libp2p.io/) protocol stack implemented in Go.

These binaries are compiled for production deployment and include the complete libp2p protocol suite: multi-transport connectivity, stream multiplexing, encrypted channels, peer discovery, content routing via Kademlia DHT, GossipSub pubsub messaging, relay, hole-punching, and more — all in a single self-contained executable.

> **No Go installation required.** Download, make executable, run.

---

## Binary Releases

| File | Architecture | OS | Linking | Size | SHA-256 |
|------|-------------|-----|---------|------|---------|
| `libp2p-linux-amd64` | `x86_64` | Linux | Dynamic (glibc) | 24.0 MB | `047c8758...` |
| `libp2p-linux-arm64` | `AArch64` | Linux | Static | 22.5 MB | `2369daec...` |

<details>
<summary><b>Full SHA-256 Checksums</b></summary>

```
047c87580e0d2f082b9ee45a0d99f34b42ad5439e09c778fc2618f9070b9b987  libp2p-linux-amd64
2369daec399948ac7c9b71ae94a38bc861a2b75572443199fde186529019e5da  libp2p-linux-arm64
```

Verify your download:

```bash
sha256sum -c <<'EOF'
047c87580e0d2f082b9ee45a0d99f34b42ad5439e09c778fc2618f9070b9b987  libp2p-linux-amd64
2369daec399948ac7c9b71ae94a38bc861a2b75572443199fde186529019e5da  libp2p-linux-arm64
EOF
```

</details>

---

## Quick Start

### Linux (x86_64 / amd64)

```bash
# Download
curl -L -o libp2p-node https://github.com/<your-user>/<your-repo>/releases/latest/download/libp2p-linux-amd64

# Verify checksum
echo "047c87580e0d2f082b9ee45a0d99f34b42ad5439e09c778fc2618f9070b9b987  libp2p-node" | sha256sum -c

# Make executable
chmod +x libp2p-node

# Run
./libp2p-node
```

### Linux (ARM64 / AArch64)

For Raspberry Pi 4/5, AWS Graviton, Ampere, Apple Silicon VMs, and any other 64-bit ARM Linux system:

```bash
# Download
curl -L -o libp2p-node https://github.com/<your-user>/<your-repo>/releases/latest/download/libp2p-linux-arm64

# Verify checksum
echo "2369daec399948ac7c9b71ae94a38bc861a2b75572443199fde186529019e5da  libp2p-node" | sha256sum -c

# Make executable
chmod +x libp2p-node

# Run
./libp2p-node
```

> **ARM64 note:** The ARM64 binary is fully statically linked and has **no external dependencies** — it runs on any 64-bit ARM Linux system regardless of libc version or distribution.

---

## Supported Transports

The binaries include the full go-libp2p transport stack:

| Transport | Protocol | Notes |
|-----------|----------|-------|
| **TCP** | `/ip4/.../tcp/...` `/ip6/.../tcp/...` | Default transport; port reuse enabled |
| **QUIC v1** | `/ip4/.../udp/.../quic-v1` | HTTP/3-based, 0-RTT capable |
| **WebTransport** | `/ip4/.../udp/.../quic-v1/webtransport` | QUIC-based, browser-compatible |
| **WebRTC Direct** | `/ip4/.../udp/.../webrtc-direct` | Browser-to-server without TURN |
| **WebSocket** | `/ip4/.../tcp/.../ws` | For browser-compatible nodes |
| **TCP Demultiplexer** | `tcp-demultiplex` | Shared-port multi-protocol |

---

## Supported Protocols

### Security & Encryption

| Protocol | Description |
|----------|-------------|
| **Noise** | XX handshake pattern (default) |
| **TLS 1.3** | x509 certificate-based channel security |

### Stream Multiplexing

| Protocol | Description |
|----------|-------------|
| **yamux v5** | Default multiplexer (`go-yamux/v5 v5.0.1`) |

### Peer Discovery

| Protocol | Description |
|----------|-------------|
| **mDNS** | Local network discovery (IPv4 + IPv6 multicast) |
| **Kademlia DHT** | `go-libp2p-kad-dht v0.39.1` — global peer routing |
| **AutoNAT v1 + v2** | NAT type detection and reachability probing |

### Connectivity & NAT Traversal

| Protocol | Description |
|----------|-------------|
| **Circuit Relay v2** | Relayed connectivity through relay nodes |
| **DCUtR / Hole Punching** | Direct connection upgrade through relay |
| **AutoRelay** | Automatic relay selection and management |

### Application Protocols

| Protocol | Description |
|----------|-------------|
| **GossipSub** | `go-libp2p-pubsub v0.16.0` — epidemic broadcast |
| **Kademlia DHT** | Content routing and provider records |
| **Identify** | Peer metadata exchange (addresses, protocols, agent) |
| **Identify Push** | Proactive address/protocol update notifications |
| **Ping** | Round-trip latency measurement |

---

## Dependencies & Modules

Core libp2p modules bundled in these binaries:

| Module | Version |
|--------|---------|
| `github.com/libp2p/go-libp2p` | `v0.48.0` |
| `github.com/libp2p/go-libp2p-kad-dht` | `v0.39.1` |
| `github.com/libp2p/go-libp2p-pubsub` | `v0.16.0` |
| `github.com/libp2p/go-libp2p-kbucket` | `v0.8.0` |
| `github.com/libp2p/go-libp2p-record` | `v0.3.1` |
| `github.com/libp2p/go-libp2p-routing-helpers` | `v0.7.5` |
| `github.com/libp2p/go-yamux/v5` | `v5.0.1` |
| `github.com/libp2p/go-buffer-pool` | `v0.1.0` |
| `github.com/libp2p/go-cidranger` | `v1.1.0` |
| `github.com/libp2p/go-flow-metrics` | `v0.3.0` |
| `github.com/libp2p/go-libp2p-asn-util` | `v0.4.1` |
| `github.com/libp2p/go-msgio` | `v0.3.0` |
| `github.com/libp2p/go-netroute` | `v0.4.0` |
| `github.com/libp2p/go-reuseport` | `v0.4.0` |

<details>
<summary><b>Notable third-party dependencies</b></summary>

| Module | Purpose |
|--------|---------|
| `github.com/quic-go/quic-go` | QUIC transport implementation |
| `github.com/quic-go/webtransport-go` | WebTransport over QUIC |
| `github.com/pion/webrtc` | WebRTC Direct transport |
| `github.com/pion/dtls` | DTLS for WebRTC |
| `github.com/pion/sctp` | SCTP data channels |
| `github.com/pion/ice` | ICE for NAT traversal |
| `github.com/multiformats/go-multiaddr` | Multiaddress parsing |
| `github.com/ipfs/go-cid` | Content ID handling |
| `go.opentelemetry.io/otel` | Observability / tracing |
| `github.com/prometheus/client_golang` | Prometheus metrics |

</details>

---

## Verification

Always verify the integrity of downloaded binaries before running them.

### SHA-256 (recommended)

```bash
# amd64
sha256sum libp2p-linux-amd64
# expected: 047c87580e0d2f082b9ee45a0d99f34b42ad5439e09c778fc2618f9070b9b987

# arm64
sha256sum libp2p-linux-arm64
# expected: 2369daec399948ac7c9b71ae94a38bc861a2b75572443199fde186529019e5da
```

### Confirm binary identity

```bash
file libp2p-linux-amd64
# ELF 64-bit LSB executable, x86-64, version 1 (SYSV), dynamically linked

file libp2p-linux-arm64
# ELF 64-bit LSB executable, ARM aarch64, version 1 (SYSV), statically linked
```

### Check embedded Go build info

```bash
strings libp2p-linux-amd64 | grep "^go1\."
# go1.25.7

strings libp2p-linux-amd64 | grep "go-libp2p"
# dep    github.com/libp2p/go-libp2p    v0.48.0    ...
```

---

## System Requirements

### `libp2p-linux-amd64`

| Requirement | Value |
|------------|-------|
| OS | Linux (kernel 3.2+) |
| Architecture | x86_64 (64-bit) |
| libc | glibc ≥ 2.17 (most distros since 2012) |
| Linking | Dynamic — requires `libc.so.6` |
| Tested on | Ubuntu 20.04+, Debian 11+, RHEL 8+, Arch |

> If you encounter `version GLIBC_X.XX not found`, use the ARM64 static binary or compile from source.

### `libp2p-linux-arm64`

| Requirement | Value |
|------------|-------|
| OS | Linux |
| Architecture | AArch64 / ARM64 (64-bit) |
| libc | None — fully statically linked |
| Tested on | Raspberry Pi 4/5 (64-bit OS), AWS Graviton 2/3, Alpine Linux ARM64 |

---

## Build Information

These binaries were compiled with:

```
Go toolchain : go1.25.7
go-libp2p    : v0.48.0
Build target : linux/amd64, linux/arm64
Strip        : yes (debug symbols removed)
CGO (amd64)  : enabled
CGO (arm64)  : disabled (static build)
```

### Reproducing the build

```bash
# Clone and build from source
git clone https://github.com/libp2p/go-libp2p.git
cd go-libp2p

# linux/amd64
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o libp2p-linux-amd64 .

# linux/arm64 (fully static)
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o libp2p-linux-arm64 .
```

---

## Observability

The binaries expose runtime metrics and tracing out of the box:

- **Prometheus metrics** — peer counts, stream counts, bandwidth, DHT operations, pubsub delivery, relay connections, GC stats, memory classes
- **OpenTelemetry tracing** — configurable via `OTEL_*` environment variables
- **Structured logging** — controlled via `GOLOG_LOG_LEVEL` and `GOLOG_OUTPUT`

```bash
# Set log level (debug, info, warn, error)
GOLOG_LOG_LEVEL=info ./libp2p-node

# JSON log output
GOLOG_LOG_FMT=json ./libp2p-node

# Per-subsystem log level
GOLOG_LOG_LEVEL="dht=debug,pubsub=info,*=warn" ./libp2p-node
```

---

## Security Notes

- All peer connections are **encrypted by default** (Noise XX or TLS 1.3). Plaintext connections are not accepted.
- Peer identity is enforced via **Ed25519 / ECDSA / RSA** cryptographic keys.
- FIPS 140-3 constraints are enforced in the TLS and crypto layers (RSA keys ≥ 2048 bits, ECDH curves restricted to approved list).
- Private network support is available via a pre-shared key protector.

---

## License

This project is distributed under the [MIT License](LICENSE).

go-libp2p and its dependencies are copyright their respective authors and are used under the terms of the MIT License. See [github.com/libp2p/go-libp2p](https://github.com/libp2p/go-libp2p) for upstream licensing details.

---

<div align="center">

Built with [go-libp2p v0.48.0](https://github.com/libp2p/go-libp2p) · Compiled with Go 1.25.7

</div>
