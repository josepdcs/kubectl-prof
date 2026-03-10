# Add Memray Python Memory Profiler

## Overview

Add [memray](https://github.com/bloomberg/memray) as a second profiling tool for Python alongside the existing `pyspy`. Memray is a memory profiler for Python that can attach to running processes and capture heap allocations.

It integrates into the existing profiler architecture using the same Docker image as pyspy (`python` tag).

**Supported output types:** `flamegraph` (HTML), `summary` (text), `tree` (text)

**Profiling flow:**
1. `memray attach --aggregate -o <raw.bin> --duration <seconds> <pid>` — captures allocations
2. Report generation from the binary:
   - FlameGraph: `memray flamegraph <raw.bin> -o <output.html>`
   - Summary/Tree: `memray summary|tree <raw.bin>` → stdout → file

## Context (from discovery)

- `api/profiling_tools.go` — ProfilingTool constants and language→tool mappings
- `api/output_types.go` — OutputType constants and tool→outputtype mappings
- `internal/agent/profiler/common/common.go` — `GetFileExtension` per tool
- `internal/agent/profiler/python.go` — existing PythonProfiler (pyspy) implementation to mirror
- `internal/agent/profiler/python_mock.go` — mock pattern to follow
- `internal/agent/profiler/factory.go` — profiler factory switch
- `internal/cli/kubernetes/job/python.go` — K8s job creator (reused for memray)
- `internal/cli/kubernetes/job/job_factory.go` — job creator factory
- `docker/python/Dockerfile` — adds memray alongside py-spy in pyspybuild stage

## Development Approach

- **Testing approach:** Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**
- Run `go test ./...` after each task

## Testing Strategy

- **Unit tests:** follow given/when/then table-driven pattern used throughout codebase
- Mock `MemrayManager` interface the same way `mockPythonManager` mocks `PythonManager`
- No e2e tests needed (no UI changes)

## Progress Tracking

- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix

## Implementation Steps

### Task 1: API — add Memray tool constant and mappings

**Files:**
- Modify: `api/profiling_tools.go`
- Modify: `api/profiling_tools_test.go`

- [x] add `Memray ProfilingTool = "memray"` constant in `profiling_tools.go`
- [x] add `Memray` to `profilingTools` slice
- [x] add `Memray` to `GetProfilingToolsByProgrammingLanguage[Python]`: `{Pyspy, Memray}`
- [x] update `GetProfilingTool` for Python: `Summary`/`Tree` → `Memray`; keep `FlameGraph` → `Pyspy` (backward compat)
- [x] add `TestIsSupportedProfilingTool` case for `"memray"` returning `true`
- [x] add `TestIsValidProfilingTool` case for `Memray + Python` returning `true`
- [x] add `TestGetProfilingToolsByProgrammingLanguage` case updating Python expected to `{Pyspy, Memray}`
- [x] add `TestGetProfilingTool` cases: `Python + Summary → Memray`, `Python + Tree → Memray`
- [x] run tests — must pass before task 2

### Task 2: API — add Memray output type mapping and file extensions

**Files:**
- Modify: `api/output_types.go`
- Modify: `internal/agent/profiler/common/common.go`
- Modify: `internal/agent/profiler/common/common_test.go`

- [x] add `Memray: {FlameGraph, Summary, Tree}` to `GetOutputTypesByProfilingTool` in `output_types.go`
- [x] add `case api.Memray` to `GetFileExtension` in `common.go`:
  - `FlameGraph` → `.html`
  - `Summary`, `Tree` → `.txt`
- [x] add test cases to `common_test.go` for `Memray + FlameGraph`, `Memray + Summary`, `Memray + Tree`
- [x] run tests — must pass before task 3

### Task 3: Agent profiler — MemrayProfiler implementation

**Files:**
- Create: `internal/agent/profiler/python_memory.go`
- Create: `internal/agent/profiler/python_memory_mock.go`
- Create: `internal/agent/profiler/python_memory_test.go`

- [x] create `python_memory.go` with:
  - `memrayLocation = "/app/memray"` constant
  - `memrayDelayBetweenJobs` constant (2s like pyspy)
  - `memrayCommand` func: `memray attach --aggregate -o <raw.bin> --duration <seconds> <pid>`
  - `memrayReportCommand` func: generates report from raw.bin for each output type
  - `MemrayProfiler` struct embedding `MemrayManager` interface
  - `MemrayManager` interface with `invoke` and `handleReport` methods
  - `memrayManager` concrete struct with `commander` and `publisher`
  - `NewMemrayProfiler` constructor
  - `SetUp`, `Invoke`, `CleanUp` methods (same pattern as `PythonProfiler`)
  - `invoke`: runs attach cmd → runs report cmd → publishes result
  - `handleReport`: for `FlameGraph` calls memray flamegraph; for `Summary`/`Tree` writes stdout to file
- [x] create `python_memory_mock.go` with `mockMemrayManager` following `mockPythonManager` pattern
- [x] create `python_memory_test.go` with table-driven tests for:
  - `TestMemrayProfiler_SetUp` (setup, setup with given PID, fail when PID not found)
  - `TestMemrayProfiler_Invoke` (invoke success, invoke fail)
  - `TestMemrayProfiler_CleanUp`
  - `Test_memrayManager_invoke` (FlameGraph, Summary, Tree, command fail, publish fail)
  - `Test_memrayManager_handleReport` (FlameGraph, Summary, Tree, error cases)
- [x] run tests — must pass before task 4

### Task 4: Agent profiler factory — register Memray

**Files:**
- Modify: `internal/agent/profiler/factory.go`
- Modify: `internal/agent/profiler/factory_test.go`

- [ ] add `case api.Memray: return NewMemrayProfiler(...)` to `Get` switch in `factory.go`
- [ ] add test case for `api.Memray` in `factory_test.go`
- [ ] run tests — must pass before task 5

### Task 5: K8s job factory — route Python+Memray to pythonCreator

**Files:**
- Modify: `internal/cli/kubernetes/job/job_factory.go`
- Modify: `internal/cli/kubernetes/job/job_factory_test.go`

- [ ] update `case api.Python` in `job_factory.go` to handle `api.Memray` tool (returns `&pythonCreator{}`, same as current default)
- [ ] add test case for `Python + Memray` in `job_factory_test.go`
- [ ] run tests — must pass before task 6

### Task 6: Dockerfile — add memray to Python image

**Files:**
- Modify: `docker/python/Dockerfile`

- [ ] add `pip3 install memray` to the `pyspybuild` stage (alongside py-spy install)
- [ ] copy `memray` binary to final image: `COPY --from=pyspybuild /usr/local/bin/memray /app/memray`

### Task 7: Verify acceptance criteria

- [ ] verify all output types (FlameGraph, Summary, Tree) are properly registered
- [ ] verify `IsSupportedProfilingTool("memray")` returns true
- [ ] verify `IsValidProfilingTool(Memray, Python)` returns true
- [ ] run full test suite: `go test ./...`
- [ ] verify test coverage for new files

### Task 8: [Final] Update documentation

- [ ] update README.md if it lists supported profiling tools/languages
- [ ] move this plan to `docs/plans/completed/`

## Technical Details

**memray attach command:**
```
memray attach --aggregate -o /tmp/prof-memray-raw-<pid>-<iter>.bin --duration <seconds> <pid>
```

**Report generation commands:**
```
# FlameGraph (HTML output):
memray flamegraph /tmp/prof-memray-raw-<pid>-<iter>.bin -o /tmp/prof-flamegraph-<pid>-<iter>.html

# Summary (text to stdout):
memray summary /tmp/prof-memray-raw-<pid>-<iter>.bin

# Tree (text to stdout):
memray tree /tmp/prof-memray-raw-<pid>-<iter>.bin
```

**File extensions (`GetFileExtension`):**
- `Memray + FlameGraph` → `.html`
- `Memray + Summary` → `.txt`
- `Memray + Tree` → `.txt`

**Raw binary intermediate file:**
Use `api.Raw` output type for the intermediate `.bin` file name, same pattern as pyspy uses `api.Raw` for the intermediate `.txt` before flamegraph conversion.

## Post-Completion

**Manual verification:**
- Deploy updated Python Docker image to a test cluster
- Profile a running Python pod with `--tool memray --output flamegraph`
- Verify HTML flamegraph is returned and renderable in browser
- Profile with `--output summary` and `--output tree`, verify text output

**External:**
- Docker image must be rebuilt and pushed with memray installed
