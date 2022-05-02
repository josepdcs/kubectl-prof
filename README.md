# Kubectl Prof

This is a kubectl plugin that allows you to profile production applications with low-overhead by generating
[FlameGraphs](http://www.brendangregg.com/flamegraphs.html) and many other outputs as [JFR](https://docs.oracle.com/javacomponents/jmc-5-4/jfr-runtime-guide/about.htm#),
thread dump, heap dump and class histogram for Java applications by using [jcmd](https://download.java.net/java/early_access/panama/docs/specs/man/jcmd.html). For Python applications, thread dump output and [speed scope](https://github.com/jlfwong/speedscope) format file are also supported. See [Usage](#usage) section.
More functionalities will be added in the future.

Running `kubectl-prof` does **not** require any modification to existing pods.

This is an open source fork of [kubectl-flame](https://github.com/yahoo/kubectl-flame) with several new features and bug fixes.

## Table of Contents

- [Requirements](#requirements)
- [Usage](#usage)
- [Installation](#installation)
- [How it works](#how-it-works)
- [Contribute](#contribute)
- [Maintainers](#maintainers)
- [License](#license)

## Requirements

* Supported languages: Go, Java (any JVM based language), Python, Ruby, and NodeJS
* Kubernetes that use some of the following container runtimes:
    * CRI-O (default)
    * Containerd
    * Docker (support will be removed since is deprecated as container runtime from Kubernetes v1.20 and will be removed in [Kubernetes v1.24](https://kubernetes.io/blog/2022/01/07/kubernetes-is-moving-on-from-dockershim/))

## Usage

### Profiling Java Pod

In order to profile a Java application in pod `mypod` for 1 minute and save the flamegraph as `/tmp/flamegraph.html` run:

```shell
kubectl prof mypod -t 1m --lang java -f /tmp/flamegraph.html
```

*NOTICE*: for java case, last version of [async-profiler](https://github.com/jvm-profiling-tools/async-profiler) produces html output, while rest of tools still producing svg outputs

### Profiling Alpine based container

Profiling Java application in alpine based containers require using `--alpine` flag:

```shell
kubectl prof mypod -t 1m -f /tmp/flamegraph.html --lang java --alpine
```

*NOTICE*: this is only required for Java apps, the `--alpine` flag is unnecessary for Go profiling.

### Profiling Java Pod and generate JFR output

Profiling Java Pod and generate JFR output require using `-o/--output jfr` option:

```shell
kubectl prof mypod -f /tmp/flight.jfr -t 5m -l java -o jfr 
```

### Profiling Java Pod and generate JFR output but by using jcmd

In this case, profiling Java Pod and generate JFR output require using `-o/--output jfr` and `--tool jcmd` options:

```shell
kubectl prof mypod -f /tmp/flight.jfr -t 5m -l java -o jfr --tool jcmd
```

### Profiling Java Pod and generate thread dump output but by using jcmd

In this case, profiling Java Pod and generate the thread dump output require using `-o/--output threaddump` and `--tool jcmd` options:

```shell
kubectl prof mypod -f /tmp/threaddump.txt -l java -o threaddump --tool jcmd
```

### Profiling Java Pod and generate heap dump output (hprof format) but by using jcmd

In this case, profiling Java Pod and generate the heap dump output require using `-o/--output heapdump` and `--tool jcmd` options:

```shell
kubectl prof mypod -f /tmp/heapdump.hprof -l java -o heapdump --tool jcmd
```

### Profiling Java Pod and generate heap histogram output (hprof format) but by using jcmd

In this case, profiling Java Pod and generate the heap histogram output require using `-o/--output heaphistogram` and `--tool jcmd` options:

```shell
kubectl prof mypod -f /tmp/heaphistogram.txt -l java -o heaphistogram --tool jcmd
```

### Profiling specifying the container runtime

Supported container runtimes values are: `crio`, `containerd` and `docker`

```shell
kubectl prof mypod -t 1m -f /tmp/flamegraph.html --lang java --runtime crio
```

### Profiling Python Pod

In order to profile a Python application in pod `mypod` for 1 minute and save the flamegraph as `/tmp/flamegraph.svg` run:

```shell
kubectl prof mypod -t 1m --lang python -f /tmp/flamegraph.svg
```

### Profiling Python Pod and generate thread dump output

In this case, profiling Java Pod and generate the thread dump output require using `-o/--output threaddump` option:

```shell
kubectl prof mypod -t 1m --lang python -f /tmp/threaddump.txt -o threaddump 
```

### Profiling Python Pod and generate speed scope output format file

In this case, profiling Java Pod and generate the thread dump output require using `-o/--output speedscope` option:

```shell
kubectl prof mypod -t 1m --lang python -f /tmp/speedscope.json -o speedscope 
```

### Profiling Golang Pod

In order to profile a Python application in pod `mypod` for 1 minute and save the flamegraph as `/tmp/flamegraph.svg` run:

```shell
kubectl prof mypod -t 1m --lang go -f /tmp/flamegraph.svg
```

### Profiling Node Pod

In order to profile a Python application in pod `mypod` for 1 minute and save the flamegraph as `/tmp/flamegraph.svg` run:

```shell
kubectl prof mypod -t 1m --lang node -f /tmp/flamegraph.svg
```

### Profiling Ruby Pod

In order to profile a Ruby application in pod `mypod` for 1 minute and save the flamegraph as `/tmp/flamegraph.svg` run:

```shell
kubectl prof mypod -t 1m --lang ruby -f /tmp/flamegraph.svg
```

### Profiling sidecar container

Pods that contains more than one container require specifying the target container as an argument:

```shell
kubectl prof mypod -t 1m --lang go -f /tmp/flamegraph.svg mycontainer
```

### Profiling Golang multi-process container

Profiling Go application in pods that contains more than one process require specifying the target process name
via `--pgrep` flag:

```shell
kubectl prof mypod -t 1m --lang go -f /tmp/flamegraph.svg --pgrep go-app
```

### Additional info. For Docker runtime

* Java profiling assumes that the process name is `java`.
* Python profiling assumes that the process name is `python`.
* Ruby profiling assumes that the process name is `ruby`.

Use `--pgrep` flag if your process name is different.

## Installation

### Pre-built binaries

See the [release](https://github.com/josepdcs/kubectl-prof/releases/tag/v0.6.1) page for the full list of pre-built assets. And download the binary according yours architecture.

### Installing for Linux x86_64
```shell
curl -sL https://github.com/josepdcs/kubectl-prof/releases/download/v0.6.1/kubectl-prof_v0.6.1_linux_x86_64.tar.gz -o kubectl-prof.tar.gz
tar xvfz kubectl-prof.tar.gz && sudo install kubectl-prof /usr/local/bin/
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

`kubectl-prof` launch a Kubernetes Job on the same node as the target pod. Under the hood `kubectl-prof`can use the following tools according the programming language:
* For Java: 
  * [async-profiler](https://github.com/jvm-profiling-tools/async-profiler) in order to generate flame graphs or JFR files.
  * [jcmd](https://download.java.net/java/early_access/panama/docs/specs/man/jcmd.html) in order to generate: JFR files, thread dumps, heap dumps and heap histogram.
    * For generating JFR files use the options: `--tool jcmd` and `-o jfr`. 
    * For generating thread dumps use the options: `--tool jcmd` and `-o threaddump`.
    * For generating heap dumps use the options: `--tool jcmd` and `-o heapdump`.
    * For generating heap histogram use the options: `--tool jcmd` and `-o histogram`.
  * Note: Default tool is [async-profiler](https://github.com/jvm-profiling-tools/async-profiler) if no option `--tool` is given and default output is flame graphs if no option `-o/--output` is also given.
* For Golang: [ebpf profiling](https://en.wikipedia.org/wiki/Berkeley_Packet_Filter). 
* For Python: [py-spy](https://github.com/benfred/py-spy). 
  * For generating thread dumps use the option: `-o threaddump`.
  * For generating speed scope use the option : `-o speedscope`.
* For Ruby: [rbspy](https://rbspy.github.io/). 
* For NodeJS: [ebpf profiling](https://en.wikipedia.org/wiki/Berkeley_Packet_Filter) and [perf](https://perf.wiki.kernel.org/index.php/Main_Page) but last one is not recommended. 
  * In order for Javascript Symbols to be resolved, node process needs to be run with `--prof-basic-prof` flag.

## Contribute

Please refer to [the contributing.md file](Contributing.md) for information about how to get involved. We welcome
issues, questions, and pull requests

## Maintainers

- Josep Damià Carbonell Seguí: josepdcs@gmail.com, josepdcs@ext.inditex.com

### Special thanks to the original Author

- Eden Federman: efederman@verizonmedia.com
- Verizon Media Code

## License

This project is licensed under the terms of the [Apache 2.0](LICENSE-Apache-2.0) open source license. Please refer
to [LICENSE](LICENSE) for the full terms.
