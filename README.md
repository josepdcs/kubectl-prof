# 🔥 Kubectl Prof

[![Build Status](https://img.shields.io/github/actions/workflow/status/josepdcs/kubectl-prof/code-verify.yml?branch=main)](https://github.com/josepdcs/kubectl-prof/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/josepdcs/kubectl-prof)](https://goreportcard.com/report/github.com/josepdcs/kubectl-prof)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue)](LICENSE)
[![Release](https://img.shields.io/github/v/release/josepdcs/kubectl-prof)](https://github.com/josepdcs/kubectl-prof/releases/latest)
[![GitHub stars](https://img.shields.io/github/stars/josepdcs/kubectl-prof?style=social)](https://github.com/josepdcs/kubectl-prof)

> **Profile your Kubernetes applications with zero overhead and zero modifications** 🚀

`kubectl-prof` is a powerful kubectl plugin that enables low-overhead profiling of applications running in Kubernetes environments. Generate [FlameGraphs](http://www.brendangregg.com/flamegraphs.html), [JFR](https://docs.oracle.com/javacomponents/jmc-5-4/jfr-runtime-guide/about.htm#) files, thread dumps, heap dumps, and many other diagnostic outputs without modifying your pods.

✨ **Key Features:**
- 🎯 **Zero modification** - Profile running pods without any changes to your deployment
- 🌐 **Multi-language support** - Java, Go, Python, Ruby, Node.js, Rust, Clang/Clang++
- 📊 **Multiple output formats** - FlameGraphs, JFR, SpeedScope, thread dumps, heap dumps, and more
- ⚡ **Low overhead** - Minimal impact on running applications
- 🔄 **Continuous profiling** - Support for both discrete and continuous profiling modes

> This is an open source fork of [kubectl-flame](https://github.com/yahoo/kubectl-flame) with enhanced features and bug fixes.

## 📋 Table of Contents

- [Requirements](#-requirements)
- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Usage](#-usage)
  - [Java Profiling](#-java-profiling)
  - [Python Profiling](#-python-profiling)
  - [Go Profiling](#-go-profiling)
  - [Node.js Profiling](#-nodejs-profiling)
  - [Ruby Profiling](#-ruby-profiling)
  - [Rust Profiling](#-rust-profiling)
  - [Clang/Clang++ Profiling](#-clangclang-profiling)
  - [Advanced Usage](#-advanced-usage)
- [How It Works](#-how-it-works)
- [Building from Source](#-building-from-source)
- [Contributing](#-contributing)
- [License](#-license)

## 📋 Requirements

### Supported Languages 💻

| Language | Status | Tools Available |
|----------|--------|-----------------|
| ☕ **Java** (JVM) | ✅ Fully Supported | async-profiler, jcmd |
| 🐹 **Go** | ✅ Fully Supported | eBPF profiling |
| 🐍 **Python** | ✅ Fully Supported | py-spy |
| 💎 **Ruby** | ✅ Fully Supported | rbspy |
| 📗 **Node.js** | ✅ Fully Supported | eBPF profiling, perf |
| 🦀 **Rust** | ✅ Fully Supported | cargo-flamegraph |
| ⚙️ **Clang/Clang++** | ✅ Fully Supported | eBPF profiling, perf |

### Container Runtimes 🐳

- **Containerd** - `--runtime=containerd` (default)
- **CRI-O** - `--runtime=crio`

## 🚀 Quick Start

Profile a Java application for 1 minute and save the FlameGraph:

```shell
kubectl prof my-pod -t 1m -l java
```

Profile a Python application and save to a specific location:

```shell
kubectl prof my-pod -t 1m -l python --local-path=/tmp
```

Profile a Rust application with cargo-flamegraph:

```shell
kubectl prof my-pod -t 1m -l rust
```

Profile multiple pods using a label selector:

```shell
kubectl prof --selector app=myapp -t 5m -l java -o jfr
```

## 📖 Usage

### ☕ Java Profiling

#### Basic FlameGraph Generation

Profile a Java application for 5 minutes and generate a FlameGraph:

```shell
kubectl prof my-pod -t 5m -l java -o flamegraph --local-path=/tmp
```

> 💡 **Tip:** If `--local-path` is omitted, the FlameGraph will be saved to the current directory.

#### Alpine-based Containers

For Java applications running in Alpine-based containers, use the `--alpine` flag:

```shell
kubectl prof mypod -t 1m -l java -o flamegraph --alpine
```

> ⚠️ **Note:** The `--alpine` flag is only required for Java applications.

#### JFR Output Generation

**Using `jcmd` (default for JFR):**

```shell
kubectl prof mypod -t 5m -l java -o jfr
```

**Using `async-profiler`:**

```shell
kubectl prof mypod -t 5m -l java -o jfr --tool async-profiler
```

#### Thread Dump

Generate a thread dump using `jcmd`:

```shell
kubectl prof mypod -l java -o threaddump
```

#### Heap Dump

Generate a heap dump in hprof format:

```shell
kubectl prof mypod -l java -o heapdump --tool jcmd
```

#### Heap Histogram

Generate a heap histogram:

```shell
kubectl prof mypod -l java -o heaphistogram --tool jcmd
```

#### Available Event Types for Java

When using `async-profiler`, you can specify different event types:

```shell
# CPU profiling (default: ctimer)
kubectl prof mypod -t 5m -l java -e cpu

# Memory allocation profiling
kubectl prof mypod -t 5m -l java -e alloc

# Lock contention profiling
kubectl prof mypod -t 5m -l java -e lock
```

**Supported events:** `cpu`, `alloc`, `lock`, `cache-misses`, `wall`, `itimer`, `ctimer`

---

### 🐍 Python Profiling

#### FlameGraph Generation

```shell
kubectl prof mypod -t 1m -l python -o flamegraph --local-path=/tmp
```

#### Thread Dump

```shell
kubectl prof mypod -l python -o threaddump --local-path=/tmp
```

#### SpeedScope Format

Generate a [SpeedScope](https://www.speedscope.app/) compatible file:

```shell
kubectl prof mypod -t 1m -l python -o speedscope --local-path=/tmp
```

---

### 🐹 Go Profiling

Profile a Go application for 1 minute:

```shell
kubectl prof mypod -t 1m -l go -o flamegraph
```

---

### 📗 Node.js Profiling

#### FlameGraph Generation

```shell
kubectl prof mypod -t 1m -l node -o flamegraph
```

> 💡 **Tip:** For JavaScript symbols to be resolved, run your Node.js process with the `--perf-basic-prof` flag.

#### Heap Snapshot

Generate a heap snapshot:

```shell
kubectl prof mypod -l node -o heapsnapshot
```

> ⚠️ **Requirements:** Your Node.js app must be run with `--heapsnapshot-signal=SIGUSR2` (default) or `--heapsnapshot-signal=SIGUSR1`.

If using `SIGUSR1`:

```shell
kubectl prof mypod -l node -o heapsnapshot --node-heap-snapshot-signal=10
```

> 📚 **Learn more:** [Node.js Heap Snapshots](https://nodejs.org/en/learn/diagnostics/memory/using-heap-snapshot)

---

### 💎 Ruby Profiling

Profile a Ruby application:

```shell
kubectl prof mypod -t 1m -l ruby -o flamegraph
```

**Available output formats:**
- `flamegraph` - FlameGraph visualization
- `speedscope` - SpeedScope format
- `callgrind` - Callgrind format

---

### 🦀 Rust Profiling

Profile a Rust application using **cargo-flamegraph** (default and recommended):

```shell
kubectl prof mypod -t 1m -l rust -o flamegraph
```

#### 🔥 cargo-flamegraph Benefits

`kubectl-prof` uses [cargo-flamegraph](https://github.com/flamegraph-rs/flamegraph) as the default profiling tool for Rust applications, offering several advantages:

- **📊 Rust-optimized profiling** - Specifically designed for Rust applications with excellent symbol resolution
- **🎨 Beautiful visualizations** - Generates clean, colorized FlameGraphs with Rust-specific color palette
- **⚡ Low overhead** - Minimal performance impact during profiling
- **🔍 Deep insights** - Captures detailed stack traces including inline functions and generics
- **🛠️ Built on perf** - Leverages the powerful Linux `perf` tool under the hood

**Available output format:**
- `flamegraph` - Interactive FlameGraph visualization (SVG format)
---

### ⚙️ Clang/Clang++ Profiling

**Clang:**

```shell
kubectl prof mypod -t 1m -l clang -o flamegraph
```

**Clang++:**

```shell
kubectl prof mypod -t 1m -l clang++ -o flamegraph
```

---

### 🎯 Advanced Usage

#### Specify Container Runtime

```shell
kubectl prof mypod -t 1m -l java --runtime crio
```

**Supported runtimes:** `containerd` (default), `crio`

#### Continuous Profiling

Profile continuously at 60-second intervals for 5 minutes:

```shell
kubectl prof mypod -l java -t 5m --interval 60s
```

> 📝 **Note:** In continuous mode, a new result is produced every interval. Only the last result is available by default.

#### Custom Resource Limits

Set CPU and memory limits for the profiling agent pod:

```shell
kubectl prof mypod -l java -t 5m \
  --cpu-limits=1 \
  --cpu-requests=100m \
  --mem-limits=200Mi \
  --mem-requests=100Mi
```

#### Cross-Namespace Profiling

Profile a pod in a different namespace:

```shell
kubectl prof mypod -n profiling \
  --service-account=profiler \
  --target-namespace=my-apps \
  -l go
```

#### Custom Agent Image

Use a custom profiling agent image:

```shell
kubectl prof mypod -l java -t 5m \
  --image=localhost/my-agent-image-jvm:latest \
  --image-pull-policy=IfNotPresent \
  --runtime containerd
```

#### Profile Multiple Pods with Label Selector

Profile all pods matching a label selector:

```shell
kubectl prof --selector app=myapp -t 5m -l java -o jfr
```

> ⚠️ **ATTENTION:** Use this option with caution as it will profile ALL pods matching the selector.

Control concurrent profiling jobs:

```shell
kubectl prof --selector app=myapp -t 5m -l java -o jfr --pool-size-profiling-jobs 5
```

#### Target Specific Process

By default, `kubectl-prof` attempts to profile all processes in the container. To target a specific process:

**Using PID:**

```shell
kubectl prof mypod -l java --pid 1234
```

**Using process name:**

```shell
kubectl prof mypod -l java --pgrep java-app-process
```

#### Capabilities Configuration

For Java profiling, `kubectl-prof` uses `PERFMON` and `SYSLOG` capabilities by default. To use `SYS_ADMIN`:

```shell
kubectl prof my-pod -t 5m -l java --capabilities=SYS_ADMIN
```

Add multiple capabilities:

```shell
kubectl prof my-pod -t 5m -l java \
  --capabilities=SYS_ADMIN \
  --capabilities=PERFMON
```

#### Node Tolerations

Profile pods on nodes with taints by specifying tolerations:

**Tolerate specific taint:**

```shell
kubectl prof my-pod -t 5m -l java \
  --tolerations=node.kubernetes.io/disk-pressure=true:NoSchedule
```

**Multiple tolerations:**

```shell
kubectl prof my-pod -t 5m -l java \
  --tolerations=node.kubernetes.io/disk-pressure=true:NoSchedule \
  --tolerations=node.kubernetes.io/memory-pressure:NoExecute \
  --tolerations=dedicated=profiling:PreferNoSchedule
```

**Toleration formats:**
- `key=value:effect` - Full specification
- `key:effect` - Any value
- `key` - Defaults to NoSchedule

---

### 📚 Get Help

For a complete list of options:

```shell
kubectl prof --help
```

## 📦 Installation

### Using Krew (Recommended) 🔌

[Krew](https://github.com/kubernetes-sigs/krew) is the plugin manager for kubectl.

1. **Install Krew** (if not already installed)

2. **Add kubectl-prof repository and install:**

```bash
kubectl krew index add kubectl-prof https://github.com/josepdcs/kubectl-prof
kubectl krew search kubectl-prof
kubectl krew install kubectl-prof/prof
kubectl prof --help
```

---

### Pre-built Binaries 📥

Download pre-built binaries from the [releases page](https://github.com/josepdcs/kubectl-prof/releases/latest).

#### Linux x86_64

```shell
wget https://github.com/josepdcs/kubectl-prof/releases/download/1.9.0/kubectl-prof_1.9.0_linux_amd64.tar.gz
tar xvfz kubectl-prof_1.9.0_linux_amd64.tar.gz
sudo install kubectl-prof /usr/local/bin/
```

#### macOS

```shell
wget https://github.com/josepdcs/kubectl-prof/releases/download/1.9.0/kubectl-prof_1.9.0_darwin_amd64.tar.gz
tar xvfz kubectl-prof_1.9.0_darwin_amd64.tar.gz
sudo install kubectl-prof /usr/local/bin/
```

#### Windows

Download the Windows binary from the [releases page](https://github.com/josepdcs/kubectl-prof/releases/latest) and add it to your PATH.

## 🔨 Building from Source

### Prerequisites

- Go 1.21 or higher
- Make
- Docker (for building agent containers)

### Build Steps

1. **Clone and install dependencies:**

```sh
go get -d github.com/josepdcs/kubectl-prof
cd $GOPATH/src/github.com/josepdcs/kubectl-prof
make install-deps
```

2. **Build the binary:**

```sh
make build
```

The binary will be available in `./bin/kubectl-prof`

3. **Build agent containers (optional):**

Modify the `DOCKER_BASE_IMAGE` property in [Makefile](Makefile), then run:

```sh
make build-docker-agents
```

## 🔧 How It Works

`kubectl-prof` launches a Kubernetes Job on the same node as the target pod. The profiling is performed using specialized tools based on the programming language:

### Profiling Tools by Language

#### ☕ Java (JVM)

**[async-profiler](https://github.com/jvm-profiling-tools/async-profiler)** - For FlameGraphs and JFR files
- FlameGraphs: `--tool async-profiler -o flamegraph` (default)
- JFR files: `--tool async-profiler -o jfr`
- Collapsed/Raw: `--tool async-profiler -o collapsed` or `-o raw`
- **Event types:** `cpu`, `alloc`, `lock`, `cache-misses`, `wall`, `itimer`, `ctimer` (default)

**[jcmd](https://download.java.net/java/early_access/panama/docs/specs/man/jcmd.html)** - For JFR, thread dumps, heap dumps
- JFR files: `--tool jcmd -o jfr` (default for jcmd)
- Thread dumps: `--tool jcmd -o threaddump`
- Heap dumps: `--tool jcmd -o heapdump`
- Heap histogram: `--tool jcmd -o heaphistogram`

#### 🐍 Python

**[py-spy](https://github.com/benfred/py-spy)** - Low-overhead Python profiler
- FlameGraphs: `-o flamegraph` (default)
- Thread dumps: `-o threaddump`
- SpeedScope: `-o speedscope`
- Raw output: `-o raw`

#### 🐹 Go

**[eBPF profiling](https://en.wikipedia.org/wiki/Berkeley_Packet_Filter)** - Kernel-level profiling
- FlameGraphs: `-o flamegraph` (default)
- Raw output: `-o raw`

#### 🦀 Rust

**[cargo-flamegraph](https://github.com/flamegraph-rs/flamegraph)** - Rust-optimized profiling tool (default)
- FlameGraphs: `--tool cargo-flamegraph -o flamegraph` (default)
- Rust-specific color palette and symbol resolution
- Low overhead, built on perf

#### 💎 Ruby

**[rbspy](https://rbspy.github.io/)** - Ruby sampling profiler
- FlameGraphs: `-o flamegraph` (default)
- SpeedScope: `-o speedscope`
- Callgrind: `-o callgrind`

#### 📗 Node.js

**[eBPF profiling](https://en.wikipedia.org/wiki/Berkeley_Packet_Filter)** (recommended) and **[perf](https://perf.wiki.kernel.org/index.php/Main_Page)**
- FlameGraphs: `-o flamegraph` (default)
- Raw output: `-o raw`
- Heap snapshot: `-o heapsnapshot`

> 💡 **Tip:** For JavaScript symbol resolution, run Node.js with `--perf-basic-prof` flag  
> 💡 **Tip:** For heap snapshots, run Node.js with `--heapsnapshot-signal` flag

#### ⚙️ Clang/Clang++

**[perf](https://perf.wiki.kernel.org/index.php/Main_Page)** (default) and **[eBPF profiling](https://en.wikipedia.org/wiki/Berkeley_Packet_Filter)**
- FlameGraphs: `-o flamegraph`
- Raw output: `-o raw`

---

### 📊 Raw Output Format

The raw output is a text file containing profiling data that can be:
- Used to generate FlameGraphs manually
- Visualized at [speedscope.app](https://www.speedscope.app/)

---

### 🔄 Profiling Modes

**Discrete Mode** (default)
- Single profiling session
- Result available when profiling completes
- Usage: `-t 5m`

**Continuous Mode**
- Multiple results at regular intervals
- Only the last result is available by default
- Client responsible for storing all results
- Usage: `-t 5m --interval 60s`

---

### 🎯 Process Targeting

By default, `kubectl-prof` profiles **all processes** in the target container matching the specified language.

**Warning example:**
```
⚠ Detected more than one PID to profile: [2508 2509]. 
  It will attempt to profile all of them. 
  Use the --pid flag to profile a specific PID.
```

**Target a specific process:**

- **By PID:** `--pid 1234`
- **By name:** `--pgrep process-name`

---

### 🔐 Capabilities

For Java profiling, `kubectl-prof` uses `PERFMON` and `SYSLOG` capabilities by default.

According to the [Kernel documentation](https://www.kernel.org/doc/html/latest/admin-guide/perf-security.html#perf-events-access-control), these capabilities should be sufficient for collecting performance samples.

To use `SYS_ADMIN` instead:

```shell
kubectl prof my-pod -t 5m -l java --capabilities=SYS_ADMIN
```

Add multiple capabilities:

```shell
kubectl prof my-pod -t 5m -l java \
  --capabilities=SYS_ADMIN \
  --capabilities=PERFMON
```

---

### 🏷️ Node Tolerations

By default, the profiling agent pod is scheduled only on nodes without taints. For nodes with taints, specify tolerations:

**Toleration formats:**
- `key=value:effect` - Full specification
- `key:effect` - Any value
- `key` - Defaults to NoSchedule

**Examples:**

```shell
# Single toleration
kubectl prof my-pod -t 5m -l java \
  --tolerations=node.kubernetes.io/disk-pressure=true:NoSchedule

# Multiple tolerations
kubectl prof my-pod -t 5m -l java \
  --tolerations=node.kubernetes.io/disk-pressure=true:NoSchedule \
  --tolerations=node.kubernetes.io/memory-pressure:NoExecute \
  --tolerations=dedicated=profiling:PreferNoSchedule
```

## 🤝 Contributing

We welcome contributions! Please refer to [Contributing.md](Contributing.md) for information about how to get involved.

**We welcome:**
- 🐛 Bug reports
- 💡 Feature requests
- 📝 Documentation improvements
- 🔧 Pull requests

---

## 👥 Maintainers

- **Josep Damià Carbonell Seguí** - josepdcs@gmail.com

### Special Thanks 🙏

Original author of [kubectl-flame](https://github.com/yahoo/kubectl-flame):
- Eden Federman - efederman@verizonmedia.com
- Verizon Media Code

---

## 📄 License

This project is licensed under the terms of the Apache 2.0 open source license. Please refer to [LICENSE](LICENSE) for the full terms.



