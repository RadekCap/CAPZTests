# Test 2: TestCheckDependencies_DockerDaemonRunning

**Location:** `test/01_check_dependencies_test.go:55-106`

**Purpose:** Verify the Docker daemon is running and accessible before running tests that depend on container operations.

---

## Command Executed

| Command | Purpose |
|---------|---------|
| `docker info --format {{.ServerVersion}}` | Check if Docker daemon is responding |

---

## Detailed Flow

```
1. Check if docker is available:
   └─ CommandExists("docker")?
      ├─ No  → CommandExists("podman")?
      │        ├─ Yes → SKIP: "Using podman instead of docker"
      │        └─ No  → SKIP: "Docker not installed"
      └─ Yes → Continue

2. Check CI environment:
   └─ CI=true OR GITHUB_ACTIONS=true?
      └─ Yes → SKIP: "Skipping Docker daemon check in CI environment"
      └─ No  → Continue

3. Run: docker info --format {{.ServerVersion}}
   └─ Success → Log "Docker daemon is running, server version: <version>"
   └─ Failure → FAIL with platform-specific help message
```

---

## Platform-Specific Help Messages

### macOS
```
To start Docker on macOS, run one of:
  open -a 'Rancher Desktop'
  open -a 'Docker Desktop'
  open -a Docker

Then wait a few seconds for the daemon to start.
```

### Linux
```
To start Docker on Linux, run:
  sudo systemctl start docker

Or check if the Docker socket exists:
  ls -la /var/run/docker.sock
```

---

## Environment Variables Checked

| Variable | Value | Effect |
|----------|-------|--------|
| `CI` | `true` | Skip test |
| `GITHUB_ACTIONS` | `true` | Skip test |

---

## Example Output

### Success
```
=== RUN   TestCheckDependencies_DockerDaemonRunning
    01_check_dependencies_test.go:104: Docker daemon is running, server version: 24.0.7
--- PASS: TestCheckDependencies_DockerDaemonRunning (0.15s)
```

### Skipped (CI Environment)
```
=== RUN   TestCheckDependencies_DockerDaemonRunning
    01_check_dependencies_test.go:71: Skipping Docker daemon check in CI environment
--- SKIP: TestCheckDependencies_DockerDaemonRunning (0.00s)
```

### Skipped (Using Podman)
```
=== RUN   TestCheckDependencies_DockerDaemonRunning
    01_check_dependencies_test.go:62: Using podman instead of docker, skipping Docker daemon check
--- SKIP: TestCheckDependencies_DockerDaemonRunning (0.00s)
```

### Failure (Daemon Not Running)
```
=== RUN   TestCheckDependencies_DockerDaemonRunning
    01_check_dependencies_test.go:96: Docker daemon is not running or not accessible.

To start Docker on macOS, run one of:
  open -a 'Rancher Desktop'
  open -a 'Docker Desktop'
  open -a Docker

Then wait a few seconds for the daemon to start.

Error: Cannot connect to the Docker daemon at unix:///var/run/docker.sock
--- FAIL: TestCheckDependencies_DockerDaemonRunning (0.30s)
```

---

## Key Observations

- This test catches Docker daemon issues early, before Kind Cluster tests fail with confusing errors
- Provides clear, platform-specific instructions for starting the Docker daemon
- Skipped in CI environments where Docker may not be available
- Supports podman as an alternative container runtime
