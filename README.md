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
- 🌐 **Multi-language support** - Java, Go, Python, Ruby, Node.js, Rust, Clang/Clang++, PHP, **.NET**
- 📊 **Multiple output formats** - FlameGraphs, JFR, SpeedScope, thread dumps, heap dumps, GC dumps, memory dumps, and more
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
  - [PHP Profiling](#-php-profiling)
  - [.NET Profiling](#-net-profiling)
  - [Advanced Usage](#-advanced-usage)
- [How It Works](#-how-it-works)
- [Building from Source](#-building-from-source)
- [Contributing](#-contributing)
- [License](#-license)

## 📋 Requirements

### Supported Languages 💻

| Language | Status | Tools Available                                       |
|----------|--------|-------------------------------------------------------|
| ☕ **Java** (JVM) | ✅ Fully Supported | async-profiler, jcmd                                  |
| 🐹 **Go** | ✅ Fully Supported | eBPF profiling                                        |
| 🐍 **Python** | ✅ Fully Supported | py-spy, memray                                  |
| 💎 **Ruby** | ✅ Fully Supported | rbspy                                                 |
| 📗 **Node.js** | ✅ Fully Supported | eBPF profiling, perf                                  |
| 🦀 **Rust** | ✅ Fully Supported | cargo-flamegraph                                      |
| ⚙️ **Clang/Clang++** | ✅ Fully Supported | eBPF profiling, perf                                  |
| 🐘 **PHP** | ✅ Fully Supported | phpspy                                                |
| 🟣 **.NET** (Core/5+) | ✅ Fully Supported | dotnet-trace, dotnet-gcdump, dotnet-counters, dotnet-dump |

### Container Runtimes 🐳

- **Containerd** - `--runtime=containerd` (default)
- **CRI-O** - `--runtime=crio`

### eBPF Profiling Tools 🔧

For eBPF profiling (Go, Node.js, Clang/Clang++), two tools are available:

#### BPF (default) - BCC-based profiler
- **Requirements:** Kernel headers or kheaders module (`/lib/modules`)
- **Usage:** Automatically used by default (no `--tool` flag needed)
- **Compatibility:** Works on most systems with kernel headers installed

#### BTF - CO-RE eBPF profiler (NEW - Experimental)
- **Requirements:** 
  - Linux kernel 5.2+ with BTF enabled (check `/sys/kernel/btf/vmlinux`)
  - BPF CPU v2 support (kernel 5.2+)
- **Usage:** Add `--tool btf` flag to your command
- **Benefits:**
  - ✅ No kernel headers required - works on DigitalOcean and other cloud providers without kheaders
  - ✅ Uses [CO-RE](https://nakryiko.com/posts/bpf-core-reference-guide/) (Compile Once - Run Everywhere) technology
  - ✅ Portable across different kernel versions without recompilation
  - ✅ Smaller Docker image size
- **Note:** Most modern distributions (Ubuntu 20.04+, RHEL 8+, etc.) include BTF by default and meet the kernel requirements

**Example using BTF:**
```shell
kubectl prof my-pod -t 1m -l go --tool btf
```

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

Profile a PHP application and generate a FlameGraph:

```shell
kubectl prof my-pod -t 1m -l php
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

Heap dumps can be large files. Use `--output-split-size` to split the result into smaller chunks for easier transfer (default: `50M`):

```shell
# Split into 100 MB chunks
kubectl prof mypod -l java -o heapdump --tool jcmd --output-split-size=100M

# Split into 1 GB chunks
kubectl prof mypod -l java -o heapdump --tool jcmd --output-split-size=1G
```

> 💡 **Tip:** The value follows the format accepted by the `split` Unix command (e.g. `50M`, `200M`, `1G`).

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

#### Additional Arguments for async-profiler

You can pass additional command-line arguments to `async-profiler` using the `--async-profiler-args` flag. This is useful for enabling specific profiling modes or customizing profiler behavior:

```shell
# Wall-clock profiling in per-thread mode (most useful for wall-clock profiling)
kubectl prof mypod -t 5m -l java -e wall --async-profiler-args -t

# Multiple additional arguments
kubectl prof mypod -t 5m -l java -e alloc --async-profiler-args -t --async-profiler-args --alloc=2m

# Combine with other options
kubectl prof mypod -t 5m -l java -e wall -o flamegraph --async-profiler-args -t
```

**Common use cases:**
- `-t` - Per-thread mode (recommended for wall-clock profiling)
- `--alloc=SIZE` - Set allocation profiling interval
- `--lock=DURATION` - Set lock profiling threshold
- `--cstack=MODE` - Control how native frames are captured

> 💡 **Tip:** Refer to the [async-profiler documentation](https://github.com/async-profiler/async-profiler) for a complete list of available arguments and their descriptions.

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

#### Memory Profiling with Memray

[Memray](https://github.com/bloomberg/memray) is a memory profiler for Python that tracks every allocation and deallocation made by your code. Unlike py-spy (which profiles CPU usage), memray reveals where your application allocates memory, helping you find memory leaks, reduce peak memory usage, and understand allocation patterns.

Memray attaches to running Python processes via GDB injection -- your application keeps running with zero downtime. No restart, no code changes, no instrumentation required.

> **Note:** You must specify `--tool memray` explicitly. The default Python profiling tool remains py-spy.

**Requirements:**
- **Capabilities:** `SYS_PTRACE` and `SYS_ADMIN` are required (for ptrace-based attach and nsenter into the target container's namespaces). Both are added automatically when `--tool memray` is used -- no extra flags needed.
- **Python versions:** 3.10, 3.11, 3.12, 3.13 (glibc-based images only)
- **Not supported:** Alpine/musl-based target containers, statically-linked Python builds

**Output types:**

| Output | Flag | Format | Description |
|--------|------|--------|-------------|
| Memory flamegraph | `-o flamegraph` | HTML | Interactive flamegraph showing allocation call stacks and sizes |
| Allocation summary | `-o summary` | Text | Tabular summary of the largest allocators by function |

**Memory flamegraph (HTML):**

```shell
kubectl prof mypod -t 1m -l python --tool memray -o flamegraph --local-path=/tmp
```

The output is a self-contained HTML file you can open in any browser. Wider frames indicate functions responsible for more memory allocations.

**Allocation summary (text):**

```shell
kubectl prof mypod -t 1m -l python --tool memray -o summary --local-path=/tmp
```

The output is a text file listing the top allocators by total bytes allocated.

**Long profiling sessions and the heartbeat interval:**

When profiling for longer durations (e.g. 5-10 minutes), network proxies or load balancers in front of your Kubernetes API server may terminate idle connections. Memray emits periodic heartbeat events to keep the log stream alive. The default interval is 30 seconds. You can adjust it with `--heartbeat-interval`:

```shell
kubectl prof mypod -t 10m -l python --tool memray -o flamegraph --heartbeat-interval=15s
```

**Targeting a specific process:**

If your pod runs multiple Python processes, use `--pid` or `--pgrep` to target a specific one:

```shell
kubectl prof mypod -t 2m -l python --tool memray -o flamegraph --pid 1234
kubectl prof mypod -t 2m -l python --tool memray -o flamegraph --pgrep my-worker
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

Heap snapshots can grow large for memory-heavy applications. Use `--output-split-size` to split the result into smaller chunks (default: `50M`):

```shell
kubectl prof mypod -l node -o heapsnapshot --output-split-size=200M
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

### 🐘 PHP Profiling

Profile a PHP 7+ application using [phpspy](https://github.com/adsr/phpspy), a low-overhead sampling profiler:

#### FlameGraph Generation

```shell
kubectl prof mypod -t 1m -l php -o flamegraph --local-path=/tmp
```

#### Raw Output

Generate raw stack-trace data that can be post-processed into a FlameGraph:

```shell
kubectl prof mypod -t 1m -l php -o raw --local-path=/tmp
```

**Available output formats:**
- `flamegraph` - Interactive FlameGraph visualization (SVG format)
- `raw` - Raw stack traces in folded format

> ⚠️ **Requirements:** The `SYS_PTRACE` capability is required. It is added automatically by `kubectl-prof`.

> 💡 **Tip:** phpspy works with PHP 7+ processes and requires no modifications to your application or PHP configuration.

---

### 🟣 .NET Profiling

`kubectl-prof` supports four specialised tools from the [.NET diagnostics suite](https://github.com/dotnet/diagnostics/blob/main/documentation) for profiling **.NET Core / .NET 5+** applications running in Kubernetes.

> ⚠️ **Requirements:** The target container must be running a .NET Core / .NET 5+ application with the .NET diagnostic socket enabled (default behaviour).

---

#### 🔥 CPU Traces — `dotnet-trace`

[`dotnet-trace`](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-trace) captures CPU samples and runtime events through the EventPipe mechanism. It is the default tool for .NET when no `--tool` flag is specified.

**SpeedScope format (default):**

```shell
kubectl prof mypod -t 30s -l dotnet -o speedscope --local-path=/tmp
```

The output is a `.speedscope.json` file that can be loaded directly at **[speedscope.app](https://www.speedscope.app/)** for interactive flame-graph analysis.

**Raw nettrace format:**

```shell
kubectl prof mypod -t 1m -l dotnet -o raw --local-path=/tmp
```

The output is a `.nettrace` binary file that can be opened with:
- [PerfView](https://github.com/microsoft/perfview) on Windows
- [Visual Studio](https://visualstudio.microsoft.com/) on Windows
- [`dotnet-trace convert`](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-trace#dotnet-trace-convert) CLI to convert it to other formats

**Using `--tool` flag explicitly:**

```shell
kubectl prof mypod -t 30s -l dotnet --tool dotnet-trace -o speedscope
```

| Flag | Output file | Visualiser |
|------|-------------|------------|
| `-o speedscope` | `.speedscope.json` | [speedscope.app](https://www.speedscope.app/) |
| `-o raw` | `.nettrace` | PerfView, Visual Studio, `dotnet-trace convert` |

---

#### 🗑️ GC Heap Dump — `dotnet-gcdump`

[`dotnet-gcdump`](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-gcdump) captures a snapshot of the managed (GC) heap. It is a lightweight alternative to a full memory dump — only managed objects are captured, so the file is much smaller than a `.dmp`.

```shell
kubectl prof mypod -l dotnet --tool dotnet-gcdump -o gcdump --local-path=/tmp
```

For large heaps, use `--output-split-size` to split the result into smaller chunks (default: `50M`):

```shell
kubectl prof mypod -l dotnet --tool dotnet-gcdump -o gcdump --output-split-size=200M --local-path=/tmp
```

> 💡 **Tip:** `dotnet-gcdump` is the recommended starting point for memory analysis. Use `dotnet-dump` only when you need native frames or a complete memory picture.

The output is a `.gcdump` file that can be opened with:
- [Visual Studio](https://visualstudio.microsoft.com/) — Heap Snapshot view
- [PerfView](https://github.com/microsoft/perfview) — GCDump viewer
- [dotnet-gcdump report](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-gcdump#dotnet-gcdump-report) CLI for a quick text summary

**Quick CLI report from the dump file:**

```shell
dotnet-gcdump report ./agent-gcdump-<pid>-1.gcdump
```

---

#### 📊 Performance Counters — `dotnet-counters`

[`dotnet-counters`](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-counters) collects runtime and application performance metrics (CPU usage, GC collections, exception rates, thread-pool queue length, etc.) over a configurable duration and writes them to a JSON file.

```shell
kubectl prof mypod -t 30s -l dotnet --tool dotnet-counters -o counters --local-path=/tmp
```

The output is a `.json` file structured as a time series of counter values. It can be:
- **Inspected directly** — plain JSON, human-readable
- **Visualised with [PerfView](https://github.com/microsoft/perfview)** — open the JSON report
- **Post-processed** with any standard JSON tooling (`jq`, Python, etc.)

**Example: print a quick summary with `jq`:**

```shell
jq '.events[] | {name: .name, value: .value}' ./agent-counters-<pid>-1.json
```

**Counters captured by default** (from the `dotnet-common` + `dotnet-sampled-thread-time` profiles):

| Counter | Description |
|---------|-------------|
| `cpu-usage` | Total CPU usage (%) |
| `working-set` | Working set memory (MB) |
| `gc-heap-size` | GC heap size (MB) |
| `gen-0-gc-count` | Gen 0 GC collections / interval |
| `gen-1-gc-count` | Gen 1 GC collections / interval |
| `gen-2-gc-count` | Gen 2 GC collections / interval |
| `exception-count` | Exceptions thrown / interval |
| `threadpool-queue-length` | Thread-pool work-item queue length |
| `active-timer-count` | Active `System.Threading.Timer` instances |

---

#### 💾 Full Memory Dump — `dotnet-dump`

[`dotnet-dump`](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-dump) captures a **point-in-time full memory dump** (`.dmp`) of the process, including both managed and native frames. This is the most comprehensive diagnostic artefact — use it for crash analysis, deadlock investigation, or when `dotnet-gcdump` does not capture enough context.

> ⚠️ **Note:** `dotnet-dump` does **not** accept a `--duration` flag — it captures the dump immediately when invoked. The `-t` flag is ignored for this tool.

```shell
kubectl prof mypod -l dotnet --tool dotnet-dump -o dump --local-path=/tmp
```

Full memory dumps can be very large (several GB for production processes). Use `--output-split-size` to split the result into smaller chunks for easier transfer (default: `50M`):

```shell
kubectl prof mypod -l dotnet --tool dotnet-dump -o dump --output-split-size=500M --local-path=/tmp
```

The output is a `.dmp` file (ELF core dump format on Linux) that can be analysed with:

- **[`dotnet-dump analyze`](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-dump#analyze-a-dump)** — cross-platform interactive SOS shell:
  ```shell
  dotnet-dump analyze ./agent-dump-<pid>-1.dmp
  ```
  Useful SOS commands inside the session:
  ```
  > clrstack          # managed call stacks for all threads
  > dumpheap -stat    # managed heap statistics
  > gcroot <address>  # find GC roots for an object
  > threads           # list all threads
  > pe                # print last exception on each thread
  ```

- **[Visual Studio](https://visualstudio.microsoft.com/)** on Windows — open the `.dmp` file for mixed managed/native debugging
- **[WinDbg](https://learn.microsoft.com/windows-hardware/drivers/debugger/debugger-download-tools)** with the [SOS extension](https://learn.microsoft.com/dotnet/core/diagnostics/sos-debugging-extension) on Windows
- **[LLDB](https://lldb.llvm.org/)** with the SOS plugin on Linux/macOS:
  ```shell
  lldb --core ./agent-dump-<pid>-1.dmp
  ```

---

#### 🗂️ .NET Tools Summary

| Tool flag | `-o` / Output type | Output file | Default? | Visualiser / Tool |
|---|---|---|---|---|
| `dotnet-trace` (default) | `speedscope` | `.speedscope.json` | ✅ | [speedscope.app](https://www.speedscope.app/) |
| `dotnet-trace` | `raw` | `.nettrace` | | PerfView, Visual Studio, `dotnet-trace convert` |
| `dotnet-gcdump` | `gcdump` | `.gcdump` | | Visual Studio, PerfView, `dotnet-gcdump report` |
| `dotnet-counters` | `counters` | `.json` | | PerfView, `jq`, Python |
| `dotnet-dump` | `dump` | `.dmp` | | `dotnet-dump analyze`, Visual Studio, WinDbg, LLDB |

---

#### 🔗 Further Reading

- [.NET diagnostics documentation](https://github.com/dotnet/diagnostics/blob/main/documentation)
- [`dotnet-trace` reference](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-trace)
- [`dotnet-gcdump` reference](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-gcdump)
- [`dotnet-counters` reference](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-counters)
- [`dotnet-dump` reference](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-dump)
- [Well-known counters in .NET](https://learn.microsoft.com/dotnet/core/diagnostics/available-counters)

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
wget https://github.com/josepdcs/kubectl-prof/releases/download/2.0.0/kubectl-prof_2.0.0_linux_amd64.tar.gz
tar xvfz kubectl-prof_2.0.0_linux_amd64.tar.gz
sudo install kubectl-prof /usr/local/bin/
```

#### macOS

```shell
wget https://github.com/josepdcs/kubectl-prof/releases/download/2.0.0/kubectl-prof_2.0.0_darwin_amd64.tar.gz
tar xvfz kubectl-prof_2.0.0_darwin_amd64.tar.gz
sudo install kubectl-prof /usr/local/bin/
```

#### Windows

Download the Windows binary from the [releases page](https://github.com/josepdcs/kubectl-prof/releases/latest) and add it to your PATH.

## 🔨 Building from Source

### Prerequisites

- Go 1.25 or higher
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

**[memray](https://github.com/bloomberg/memray)** - Python memory profiler (`--tool memray`)
- Memory flamegraph (HTML): `-o flamegraph`
- Allocation summary (text): `-o summary`
- Attaches to running processes via GDB injection (zero downtime)
- Requires `SYS_PTRACE` + `SYS_ADMIN` capabilities (added automatically)
- Supported target Python versions: 3.10, 3.11, 3.12, 3.13 (glibc-based only)


#### 🐹 Go

**eBPF Profiling** - Two options available:

1. **BPF (default)** - BCC-based profiler
   - Uses BCC tools with runtime compilation
   - Requires kernel headers (`/lib/modules`)
   - Usage: No `--tool` flag needed (default)

2. **BTF** - [CO-RE eBPF profiler](https://nakryiko.com/posts/bpf-core-reference-guide/)
   - Uses libbpf-tools with CO-RE support
   - **No kernel headers required** - only needs BTF (available on modern kernels)
   - Usage: Add `--tool btf` flag
   - Example: `kubectl prof my-pod -t 1m -l go --tool btf`

**Output formats (both tools):**
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

#### 🐘 PHP

**[phpspy](https://github.com/adsr/phpspy)** - Low-overhead sampling profiler for PHP 7+
- FlameGraphs: `-o flamegraph` (default)
- Raw output: `-o raw`

**Output formats:**
- `flamegraph` - Interactive FlameGraph visualization (SVG format)
- `raw` - Raw stack traces in folded format

#### 🟣 .NET (Core / .NET 5+)

Four tools from the [.NET diagnostics suite](https://github.com/dotnet/diagnostics/blob/main/documentation) are available, each targeting a different diagnostic scenario:

**[dotnet-trace](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-trace)** — CPU and runtime event tracing (default tool for .NET)
- SpeedScope: `--tool dotnet-trace -o speedscope` (default) → `.speedscope.json`
- Raw nettrace: `--tool dotnet-trace -o raw` → `.nettrace`
- Uses EventPipe; zero JVM-agent overhead

**[dotnet-gcdump](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-gcdump)** — Lightweight GC heap snapshot
- GC heap dump: `--tool dotnet-gcdump -o gcdump` → `.gcdump`
- Captures managed objects only; much smaller than a full dump

**[dotnet-counters](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-counters)** — Real-time performance counter collection
- Counters: `--tool dotnet-counters -o counters` → `.json`
- Captures CPU, GC, thread-pool, exception rate and other runtime metrics

**[dotnet-dump](https://learn.microsoft.com/dotnet/core/diagnostics/dotnet-dump)** — Full process memory dump
- Full dump: `--tool dotnet-dump -o dump` → `.dmp`
- Point-in-time; includes both managed and native frames
- Analysable with `dotnet-dump analyze`, Visual Studio, WinDbg, LLDB+SOS

#### 📗 Node.js

**eBPF Profiling** - Two options available (recommended):

1. **BPF (default)** - BCC-based profiler
   - Requires kernel headers (`/lib/modules`)
   - Usage: No `--tool` flag needed (default)

2. **BTF** - [CO-RE eBPF profiler](https://nakryiko.com/posts/bpf-core-reference-guide/)
   - **No kernel headers required** - only needs BTF
   - Usage: Add `--tool btf` flag
   - Example: `kubectl prof my-pod -t 1m -l node --tool btf`

**Alternative: [perf](https://perf.wiki.kernel.org/index.php/Main_Page)**
- Available for fallback if eBPF profiling unavailable

**Output formats:**
- FlameGraphs: `-o flamegraph` (default)
- Raw output: `-o raw`
- Heap snapshot: `-o heapsnapshot`

> 💡 **Tip:** For JavaScript symbol resolution, run Node.js with `--perf-basic-prof` flag  
> 💡 **Tip:** For heap snapshots, run Node.js with `--heapsnapshot-signal` flag

#### ⚙️ Clang/Clang++

**eBPF Profiling** - Two options available (recommended):

1. **BPF (default)** - BCC-based profiler
   - Requires kernel headers (`/lib/modules`)
   - Usage: No `--tool` flag needed (default)

2. **BTF** - [CO-RE eBPF profiler](https://nakryiko.com/posts/bpf-core-reference-guide/)
   - **No kernel headers required** - only needs BTF
   - Usage: Add `--tool btf` flag
   - Example: `kubectl prof my-pod -t 1m -l clang --tool btf`

**Alternative: [perf](https://perf.wiki.kernel.org/index.php/Main_Page)**
- Available for fallback if eBPF profiling unavailable

**Output formats:**
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



