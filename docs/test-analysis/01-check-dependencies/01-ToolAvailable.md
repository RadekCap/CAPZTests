# Test 1: TestCheckDependencies_ToolAvailable

**Location:** `test/01_check_dependencies_test.go:12-53`

**Purpose:** Verify all required CLI tools are installed and available in PATH, including Azure CLI via the AZ_COMMAND environment variable.

---

## Commands Executed

For each tool, the test uses `CommandExists()` helper which internally runs:

| Tool | Check Method | Fallback |
|------|--------------|----------|
| `docker` | `which docker` | `podman` |
| `kind` | `which kind` | - |
| `oc` | `which oc` | - |
| `helm` | `which helm` | - |
| `git` | `which git` | - |
| `kubectl` | `which kubectl` | - |
| `go` | `which go` | - |
| `az` | `AzCommandExists()` | Uses AZ_COMMAND env var |

**Note:** The `az` CLI is checked separately using `AzCommandExists()` which supports the `AZ_COMMAND` environment variable for custom az paths (e.g., when using a Python virtual environment on systems with incompatible system Python).

---

## Detailed Flow

```
For each tool in [docker, kind, oc, helm, git, kubectl, go]:
│
├─► Run subtest: t.Run(tool, ...)
│
├─► CommandExists(tool)?
│   │
│   ├─ Yes → Log "Tool '<tool>' is available"
│   │
│   └─ No  → Is tool == "docker"?
│            │
│            ├─ Yes → CommandExists("podman")?
│            │        │
│            │        ├─ Yes → Log "docker not found, but podman is available"
│            │        │
│            │        └─ No  → FAIL: "Required tool 'docker' is not installed"
│            │
│            └─ No  → FAIL: "Required tool '<tool>' is not installed"

For "az" (checked separately):
│
├─► Run subtest: t.Run("az", ...)
│
├─► AzCommandExists()?
│   │
│   ├─ Yes → GetAzCommand() == "az"?
│   │        │
│   │        ├─ Yes → Log "Tool 'az' is available"
│   │        │
│   │        └─ No  → Log "Using custom az command: <path>"
│   │
│   └─ No  → FAIL: "Azure CLI is not available"
```

---

## Required Tools List

```go
requiredTools := []string{
    "docker",
    "kind",
    "oc",
    "helm",
    "git",
    "kubectl",
    "go",
}
// Note: "az" is checked separately via AzCommandExists() to support custom az paths
```

---

## Example Output

```
=== RUN   TestCheckDependencies_ToolAvailable
=== RUN   TestCheckDependencies_ToolAvailable/docker
    01_check_dependencies_test.go:35: Tool 'docker' is available
=== RUN   TestCheckDependencies_ToolAvailable/kind
    01_check_dependencies_test.go:35: Tool 'kind' is available
=== RUN   TestCheckDependencies_ToolAvailable/oc
    01_check_dependencies_test.go:35: Tool 'oc' is available
=== RUN   TestCheckDependencies_ToolAvailable/helm
    01_check_dependencies_test.go:35: Tool 'helm' is available
=== RUN   TestCheckDependencies_ToolAvailable/git
    01_check_dependencies_test.go:35: Tool 'git' is available
=== RUN   TestCheckDependencies_ToolAvailable/kubectl
    01_check_dependencies_test.go:35: Tool 'kubectl' is available
=== RUN   TestCheckDependencies_ToolAvailable/go
    01_check_dependencies_test.go:35: Tool 'go' is available
=== RUN   TestCheckDependencies_ToolAvailable/az
    01_check_dependencies_test.go:49: Tool 'az' is available
--- PASS: TestCheckDependencies_ToolAvailable (0.02s)
```

---

## Key Observations

- Uses Go's `t.Run()` for subtests, allowing individual tool checks to pass/fail independently
- Docker has a special fallback to podman for container runtime flexibility
- No version requirements are checked, only presence in PATH
- Azure CLI supports `AZ_COMMAND` environment variable for custom az paths (useful when system Python is incompatible with Azure CLI)
