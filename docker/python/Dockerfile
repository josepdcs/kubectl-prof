FROM --platform=$BUILDPLATFORM golang:1.25.2-alpine AS agentbuild
ARG TARGETOS
ARG TARGETARCH
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/agent

FROM python:3.13-slim-bullseye AS pyspybuild
RUN pip3 install py-spy==0.4.1 \
    && apt-get update && apt-get install -y git \
    && git clone https://github.com/brendangregg/FlameGraph

# Build memray with all dependencies for each supported Python version.
# Dependencies like rich are needed because memray's __init__.py imports them,
# and the PAYLOAD executed inside the target process does `import memray`.
FROM python:3.10-slim-bullseye AS memray310
RUN pip install --no-cache-dir --target=/opt/memray/3.10 memray==1.19.1

FROM python:3.11-slim-bullseye AS memray311
RUN pip install --no-cache-dir --target=/opt/memray/3.11 memray==1.19.1

FROM python:3.12-slim-bullseye AS memray312
RUN pip install --no-cache-dir --target=/opt/memray/3.12 memray==1.19.1

FROM python:3.13-slim-bullseye AS memray313
RUN pip install --no-cache-dir --target=/opt/memray/3.13 memray==1.19.1

FROM python:3.13-slim-bullseye
RUN apt-get update && apt-get install -y --no-install-recommends perl procps strace gdb util-linux \
    && rm -rf /var/lib/apt/lists/* \
    && pip3 install --no-cache-dir memray==1.19.1 \
    && mkdir -p /app/FlameGraph \
    && ln -s /usr/local/bin/memray /app/memray

# Copy version-specific memray packages (with deps) for staging into target containers
COPY --from=memray310 /opt/memray/3.10 /opt/memray/3.10
COPY --from=memray311 /opt/memray/3.11 /opt/memray/3.11
COPY --from=memray312 /opt/memray/3.12 /opt/memray/3.12
COPY --from=memray313 /opt/memray/3.13 /opt/memray/3.13

# Patch memray payload to include sys.path insertion for staged package.
# The PAYLOAD string in attach.py is executed inside the target process via PyEval_EvalCode().
# Prepending sys.path.insert ensures the target can find the staged memray package at
# /tmp/.kubectl-prof-memray/ even when memray is not installed in the target.
# In memray 1.19.1, the PAYLOAD imports are: import sys / from contextlib import suppress / import memray
COPY <<'PATCH' /tmp/patch_memray.py
import pathlib
attach = pathlib.Path(
    '/usr/local/lib/python3.13/site-packages/memray/commands/attach.py'
)
code = attach.read_text()
old = 'from contextlib import suppress\n\nimport memray'
new = (
    'from contextlib import suppress\n\n'
    'sys.path.insert(0, "/tmp/.kubectl-prof-memray")\n'
    'import memray'
)
code = code.replace(old, new, 1)
assert 'sys.path.insert' in code, (
    'Payload patch failed: could not find contextlib import'
    ' + import memray sequence in attach.py PAYLOAD'
)
attach.write_text(code)
PATCH
RUN python3 /tmp/patch_memray.py && rm /tmp/patch_memray.py

COPY --from=pyspybuild /FlameGraph /app/FlameGraph
COPY --from=agentbuild /go/bin/agent /app/agent
COPY --from=pyspybuild /usr/local/bin/py-spy /app/py-spy
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]
