# Kubectl Prof

This is a kubectl plugin that allows you to profile applications with low-overhead in Kubernetes environments by
generating
[FlameGraphs](http://www.brendangregg.com/flamegraphs.html) and many other outputs
as [JFR](https://docs.oracle.com/javacomponents/jmc-5-4/jfr-runtime-guide/about.htm#),
thread dump, heap dump and class histogram for Java applications by
using [jcmd](https://download.java.net/java/early_access/panama/docs/specs/man/jcmd.html). For Python applications,
thread dump output and [speed scope](https://github.com/jlfwong/speedscope) format file are also supported.
See [Usage](#usage) section.
More functionalities will be added in the future.

Running `kubectl-prof` does **not** require any modification to existing pods.

This is an open source fork of [kubectl-flame](https://github.com/yahoo/kubectl-flame) with several new features and bug
fixes.

## Table of Contents

- [Requirements](#requirements)
- [Usage](#usage)
- [Installation](#installation)
- [How it works](#how-it-works)
- [Contribute](#contribute)
- [Maintainers](#maintainers)
- [License](#license)

## Requirements

* Supported languages: Go, Java (any JVM based language), Python, Ruby, NodeJS, Clang and Clang++.
* Kubernetes that use some of the following container runtimes:
    * **Containerd** by using flag `--runtime=containerd` (default)
    * **CRI-O** by using flag `--runtime=crio`

## Usage

### Profiling Java Pod

In order to profile a Java application in pod `mypod` for 1 minute and save the flamegraph into `/tmp` run:

```shell
kubectl prof my-pod -t 5m -l java -o flamegraph --local-path=/tmp
```

*NOTICE*:

* if `--local-path` is omitted, flamegraph result will be saved into current directory

### Profiling Alpine based container

Profiling Java application in alpine based containers require using `--alpine` flag:

```shell
kubectl prof mypod -t 1m --lang java -o flamegraph --alpine 
```

*NOTICE*: this is only required for Java apps, the `--alpine` flag is unnecessary for other languages.

### Profiling Java Pod and generate JFR output by using `jcmd` as default tool

Profiling Java Pod and generate JFR output require using `-o/--output jfr` option:

```shell
kubectl prof mypod -t 5m -l java -o jfr 
```

### Profiling Java Pod and generate JFR output but by using `async-profiler`

In this case, profiling Java Pod and generate JFR output require using `-o/--output jfr` and `--tool async-profiler`
options:

```shell
kubectl prof mypod -t 5m -l java -o jfr --tool jcmd
```

### Profiling Java Pod and generate thread dump output by using `jcmd` as default tool

In this case, profiling Java Pod and generate the thread dump output require using `-o/--output threaddump` options:

```shell
kubectl prof mypod -l java -o threaddump
```

### Profiling Java Pod and generate heap dump output (hprof format) by using `jcmd` as default tool

In this case, profiling Java Pod and generate the heap dump output require using `-o/--output heapdump` options:

```shell
kubectl prof mypod -l java -o heapdump --tool jcmd
```

### Profiling Java Pod and generate heap histogram output (hprof format) by using `jcmd` as default tool

In this case, profiling Java Pod and generate the heap histogram output require using `-o/--output heaphistogram`
options:

```shell
kubectl prof mypod -l java -o heaphistogram --tool jcmd
```

### Profiling specifying the container runtime

Supported container runtimes values are: `crio`, `containerd`.

```shell
kubectl prof mypod -t 1m --lang java --runtime crio
```

### Profiling Python Pod

In order to profile a Python application in pod `mypod` for 1 minute and save the flamegraph into `/tmp` run:

```shell
kubectl prof mypod -t 1m --lang python -o flamegraph --local-path=/tmp
```

### Profiling Python Pod and generate thread dump output

In this case, profiling Python Pod and generate the thread dump output require using `-o/--output threaddump` option:

```shell
kubectl prof mypod -t 1m --lang python --local-path=/tmp -o threaddump 
```

### Profiling Python Pod and generate speed scope output format file

In this case, profiling Python Pod and generate the thread dump output require using `-o/--output speedscope` option:

```shell
kubectl prof mypod -t 1m --lang python --local-path=/tmp -o speedscope 
```

### Profiling Golang Pod

In order to profile a Golang application in pod `mypod` for 1 minute run:

```shell
kubectl prof mypod -t 1m --lang go -o flamegraph
```

### Profiling Node Pod

In order to profile a Python application in pod `mypod` for 1 minute run:

```shell
kubectl prof mypod -t 1m --lang node -o flamegraph
```

### Profiling Ruby Pod

In order to profile a Ruby application in pod `mypod` for 1 minute run:

```shell
kubectl prof mypod -t 1m --lang ruby -o flamegraph
```

### Profiling Clang Pod

In order to profile a Clang application in pod `mypod` for 1 minute run:

```shell
kubectl prof mypod -t 1m --lang clang -o flamegraph
```

### Profiling Clang++ Pod

In order to profile a Clang++ application in pod `mypod` for 1 minute run:

```shell
kubectl prof mypod -t 1m --lang clang++ -o flamegraph
```

### Profiling with several options:

#### Profiling a pod for 5 minutes in intervals of 60 seconds for java language by giving the cpu limits, the container runtime, the agent image and the image pull policy

```shell
kubectl prof mypod -l java -o flamegraph -t 5m --interval 60s --cpu-limits=1 -r containerd --image=localhost/my-agent-image-jvm:latest --image-pull-policy=IfNotPresent
```

### Profiling in contprof namespace a pod running in contprof-apps namespace by using the profiler service account for go language

```shell
kubectl prof mypod -n contprof --service-account=profiler --target-namespace=contprof-stupid-apps -l go
```

### Profiling by setting custom resource requests and limits for the agent pod (default: neither requests nor limits are set) for python language

```shell
kubectl prof mypod --cpu-requests 100m --cpu-limits 200m --mem-requests 100Mi --mem-limits 200Mi -l python
```

### For more detailed options run:

```shell
kubectl prof --help
```

## Installation

### Using krew
Install [Krew](https://github.com/kubernetes-sigs/krew)

Install repository and plugin:

```bash
kubectl krew index add kubectl-prof https://github.com/josepdcs/kubectl-prof
kubectl krew search kubectl-prof
kubectl krew install kubectl-prof/prof
kubectl prof --help
```

### Pre-built binaries

See the [release](https://github.com/josepdcs/kubectl-prof/releases/tag/1.2.1) page for the full list of pre-built
assets. And download the binary according yours architecture.

### Installing for Linux x86_64

```shell
curl -sL https://github.com/josepdcs/kubectl-prof/releases/download/1.2.1/kubectl-prof_1.2.1_linux_x86_64.tar.gz
tar xvfz kubectl-prof_1.2.1_linux_x86_64.tar.gz && sudo install kubectl-prof /usr/local/bin/
```

## Building

### Install source code and golang dependencies

```sh
$ go get -d github.com/josepdcs/kubectl-prof
$ cd $GOPATH/src/github.com/josepdcs/kubectl-prof
$ make install-deps
```

### Build binary

```sh
$ make
```

### Build Agents Containers

Modify [Makefile](Makefile), property DOCKER_BASE_IMAGE, and run:

```sh
$ make agents
```

## How it works

`kubectl-prof` launch a Kubernetes Job on the same node as the target pod. Under the hood `kubectl-prof`can use the
following tools according the programming language:

* For Java:
    * [async-profiler](https://github.com/jvm-profiling-tools/async-profiler) in order to generate flame graphs or JFR
      files and the rest of output type supported for this tool.
      * For generating flame graphs use the option: `--tool async-profiler` and `-o flamegraph`.
      * For generating JFR files use the option: `--tool async-profiler` and `-o jfr`.
      * For generating collapsed/raw use the option: `--tool async-profiler` and `-o collapsed` or `-o raw`.
      * Note: Default output is flame graphs if no option `-o/--output` is given.
    * [jcmd](https://download.java.net/java/early_access/panama/docs/specs/man/jcmd.html) in order to generate: JFR
      files, thread dumps, heap dumps and heap histogram.
        * For generating JFR files use the options: `--tool jcmd` and `-o jfr`.
        * For generating thread dumps use the options: `--tool jcmd` and `-o threaddump`.
        * For generating heap dumps use the options: `--tool jcmd` and `-o heapdump`.
        * For generating heap histogram use the options: `--tool jcmd` and `-o histogram`.
        * Note: Default output is JFR if no option `-o/--output` is given.
    * Note: Default tool is [async-profiler](https://github.com/jvm-profiling-tools/async-profiler) if no
      option `--tool` is given and default output is flame graphs if no option `-o/--output` is also given.
* For Golang: [ebpf profiling](https://en.wikipedia.org/wiki/Berkeley_Packet_Filter).
  * For generating flame graphs use the option: `-o flamegraph`.
  * For generating raw use the option: `-o raw`.
  * Note: Default output is flame graphs if no option `-o/--output` is given.
* For Python: [py-spy](https://github.com/benfred/py-spy).
  * For generating flame graphs use the option: `-o flamegraph`.
  * For generating thread dumps use the option: `-o threaddump`.
  * For generating speed scope use the option : `-o speedscope`.
  * For generating raw use the option: `-o raw`. 
  * Note: Default output is flame graphs if no option `-o/--output` is given.
* For Ruby: [rbspy](https://rbspy.github.io/).
  * For generating flame graphs use the option: `-o flamegraph`.
  * For generating speed scope use the option : `-o speedscope`.
  * For generating callgrind use the option: `-o callgrind`.
  * Note: Default output is flame graphs if no option `-o/--output` is given.
* For Node.js: [ebpf profiling](https://en.wikipedia.org/wiki/Berkeley_Packet_Filter) and [perf](https://perf.wiki.kernel.org/index.php/Main_Page) but last one is not recommended.
  * For generating flame graphs use the option: `-o flamegraph`.
  * For generating raw use the option: `-o raw`.
  * Note: Default output is flame graphs if no option `-o/--output` is given.
  * In order for Javascript Symbols to be resolved, node process needs to be run with `--prof-basic-prof` flag.
* For Clang and Clang++: [perf](https://perf.wiki.kernel.org/index.php/Main_Page) is the default profiler
  but [ebpf profiling](https://en.wikipedia.org/wiki/Berkeley_Packet_Filter) is also supported.

The raw output is a text file with the raw data from the profiler. It could be used to generate flame graphs, or you can use https://www.speedscope.app/ to visualize the data.

`kubectl-prof` also supports to work in modes discrete and continuous:

* In discrete mode: only one profiling result is requested. Once this result is obtained, the profiling process
  finishes. This is the default behaviour when only using `-t time` option.
* In continuous mode: can produce more than one result. Given a session duration and an interval, a result is produced
  every interval until the profiling session finishes. Only the last produced result is available. It is client
  responsibility to store all the session results.
    * For using this option you must use the `--interval time` option in addition to `-t time`.

In addition, `kubectl-prof` will attempt to profile all the processes detected in the container. 
It will try to profile them all based on the provided language. When this happens, the tool will display a warning similar to:


    ⚠ Detected more than one PID to profile: [2508 2509]. It will be attempt to profile all of them. Use the --pid flag specifying the corresponding PID if you only want to profile one of them.

But if you want to profile a specific process, you have two options:
* Provide the specific PID using the `--pid PID` flag if you know the PID (the previous warning can help you identify the PID you want to profile). 
* Provide a process name using the `--pgrep process-matching-name` flag.

## Contribute

Please refer to [the contributing.md file](Contributing.md) for information about how to get involved. We welcome
issues, questions, and pull requests

## Maintainers

- Josep Damià Carbonell Seguí: josepdcs@gmail.com

### Special thanks to the original Author of [kubectl-flame](https://github.com/yahoo/kubectl-flame)

- Eden Federman: efederman@verizonmedia.com
- Verizon Media Code

## License

This project is licensed under the terms of the [Apache 2.0](LICENSE-Apache-2.0) open source license. Please refer
to [LICENSE](LICENSE) for the full terms.

## Project status

| Service | Status |
|-----|:---|
| [Github Actions](https://github.com/actions/) |  ![Build Status][actions-image]  |
| [GoReport](https://goreportcard.com/) |  [![Go Report Card][goreportcard-image]][goreportcard-url] |

[actions-image]: https://github.com/josepdcs/kubectl-prof/actions

[goreportcard-image]: https://goreportcard.com/badge/github.com/josepdcs/kubectl-prof

[goreportcard-url]: https://goreportcard.com/report/github.com/josepdcs/kubectl-prof


