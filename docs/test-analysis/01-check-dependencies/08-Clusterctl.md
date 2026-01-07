# Test 8: TestCheckDependencies_Clusterctl_IsAvailable

**Location:** `test/01_check_dependencies_test.go:258-297`

**Purpose:** Check if clusterctl is available in PATH. This is an informational check that warns users but does not fail if clusterctl is missing.

---

## Command Executed

| Command | Purpose |
|---------|---------|
| `clusterctl version` | Get clusterctl version (if available) |

---

## Detailed Flow

```
1. Check if clusterctl is in PATH:
   └─ CommandExists("clusterctl")?
      │
      ├─ Yes → Run: clusterctl version
      │        └─► Success?
      │            ├─ Yes → Log version
      │            └─ No  → Log "found but version check failed"
      │
      └─ No  → Log informational message with install instructions
               (Does NOT fail the test)
```

---

## Platform-Specific Install Instructions

### macOS
```
To install clusterctl on macOS:
  brew install clusterctl

Or manually:
  curl -L https://github.com/kubernetes-sigs/cluster-api/releases/latest/download/clusterctl-darwin-arm64 -o /usr/local/bin/clusterctl
  chmod +x /usr/local/bin/clusterctl
```

### Linux
```
To install clusterctl on Linux:
  curl -L https://github.com/kubernetes-sigs/cluster-api/releases/latest/download/clusterctl-linux-amd64 -o /usr/local/bin/clusterctl
  chmod +x /usr/local/bin/clusterctl
```

---

## Example Output

### Success (Found in PATH)
```
=== RUN   TestCheckDependencies_Clusterctl_IsAvailable
    01_check_dependencies_test.go:268: clusterctl version: clusterctl version: &version.Info{Major:"1", Minor:"6", GitVersion:"v1.6.0", ...}
--- PASS: TestCheckDependencies_Clusterctl_IsAvailable (0.05s)
```

### Not Found (Informational Only)
```
=== RUN   TestCheckDependencies_Clusterctl_IsAvailable
    01_check_dependencies_test.go:291: clusterctl not found in system PATH.

clusterctl is required for cluster monitoring (TestDeployment_MonitorCluster) and
kubeconfig retrieval (TestVerification_GetKubeconfig).

It will be looked for in cluster-api-installer's bin directory during test execution.
If not available there either, those tests will be skipped.

To install clusterctl on macOS:
  brew install clusterctl

Or manually:
  curl -L https://github.com/kubernetes-sigs/cluster-api/releases/latest/download/clusterctl-darwin-arm64 -o /usr/local/bin/clusterctl
  chmod +x /usr/local/bin/clusterctl
--- PASS: TestCheckDependencies_Clusterctl_IsAvailable (0.00s)
```

---

## Key Observations

- This test is **informational only** - it does not fail if clusterctl is missing
- clusterctl is used in later phases:
  - `TestDeployment_MonitorCluster` for `clusterctl describe cluster`
  - `TestVerification_RetrieveKubeconfig` for `clusterctl get kubeconfig`
- If not in system PATH, tests will look for it in `cluster-api-installer/bin/` directory
- If not available anywhere, dependent tests will be skipped with a clear message

---

## When Is clusterctl Actually Required?

| Test | Phase | Uses clusterctl for |
|------|-------|---------------------|
| `TestDeployment_MonitorCluster` | 5 | `clusterctl describe cluster` |
| `TestVerification_RetrieveKubeconfig` | 6 | `clusterctl get kubeconfig` (fallback method) |

Both tests have fallback methods or will skip gracefully if clusterctl is unavailable.
