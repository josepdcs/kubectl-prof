FROM --platform=$BUILDPLATFORM golang:1.25.2-alpine AS agentbuild
ARG TARGETOS
ARG TARGETARCH
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/agent

FROM alpine:3.23 AS builder
RUN apk add --no-cache git  \
    && git clone https://github.com/brendangregg/FlameGraph

FROM alpine:3.23 AS libbpf-builder
# Install build dependencies for libbpf-tools
# Note: Alpine's musl libc allows static linking with -static flag
RUN apk add --no-cache \
    build-base \
    clang \
    llvm \
    llvm-dev \
    elfutils-dev \
    linux-headers \
    libbpf-dev \
    zlib-dev \
    zlib-static \
    zstd-dev \
    zstd-static \
    git \
    bpftool \
    argp-standalone \
    musl-dev

# Clone and build libbpf-tools profile
# Using -mcpu=v2 for better kernel compatibility (supports kernels 5.2+)
# Build as static binary to avoid library dependencies when copying to target container
# Create symlink for llvm-strip if needed (Alpine may have versioned binary)
RUN ln -sf /usr/bin/llvm19-strip /usr/bin/llvm-strip 2>/dev/null || \
    ln -sf $(find /usr/bin -name 'llvm*-strip' | head -1) /usr/bin/llvm-strip 2>/dev/null || true
RUN git clone --depth 1 --recurse-submodules --shallow-submodules https://github.com/iovisor/bcc.git /bcc && \
    cd /bcc/libbpf-tools && \
    sed -i '1i #include <libgen.h>' bpftool/src/gen.c && \
    sed -i 's/BPFCFLAGS := -g -O2 -Wall -Werror=undef/BPFCFLAGS := -g -O2 -Wall -Werror=undef -mcpu=v2/' Makefile && \
    sed -i 's|LIBS = -lelf -lz|LIBS = -lelf -lz -lzstd|g' bpftool/src/Makefile && \
    sed -i 's|$(QUIET_LINK)$(CC) $(BINARY_CFLAGS) $(LDFLAGS) -o $@ $^ $(LDLIBS)|$(QUIET_LINK)$(CC) $(BINARY_CFLAGS) $(LDFLAGS) -o $@ $^ $(LDLIBS) -lzstd|g' /bcc/libbpf-tools/bpftool/src/Makefile && \
    sed -i 's|$(QUIET_LINK)$(CC) $(BINARY_CFLAGS) $(LDFLAGS) -o $@ $^ $(LDLIBS)|$(QUIET_LINK)$(CC) $(BINARY_CFLAGS) $(LDFLAGS) -o $@ $^ $(LDLIBS) -lzstd|g' /bcc/src/cc/libbpf/src/Makefile && \
    find /bcc -name "Makefile*" -exec sed -i 's|-lelf -lz|-lelf -lz -lzstd|g' {} + && \
    make EXTRA_CFLAGS="-I/usr/include" EXTRA_LDFLAGS="-largp -static -lzstd" LIBS="-lelf -lz -lzstd" profile

FROM alpine:3.23
# Install only runtime dependencies (no bcc-tools needed for CO-RE)
RUN apk add --no-cache bash perl procps strace libbpf util-linux \
    && mkdir -p /app/FlameGraph /app/libbpf-profiler
COPY --from=builder /FlameGraph /app/FlameGraph
COPY --from=libbpf-builder /bcc/libbpf-tools/profile /app/libbpf-profiler/profile
RUN chmod +x /app/libbpf-profiler/profile
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/btf/profile-wrapper.sh /app/libbpf-profiler/profile-wrapper.sh
RUN chmod +x /app/libbpf-profiler/profile-wrapper.sh
COPY --from=agentbuild /go/bin/agent /app
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]