# Test 4: TestCheckDependencies_AzureEnvironment

**Location:** `test/01_check_dependencies_test.go:126-223`

**Purpose:** Validate required Azure environment variables are set, and auto-extract them from Azure CLI if not set (since login was already verified).

---

## Environment Variables Validated

| Variable | Purpose | Auto-Extract Command |
|----------|---------|---------------------|
| `AZURE_TENANT_ID` | Azure tenant ID | `az account show --query tenantId -o tsv` |
| `AZURE_SUBSCRIPTION_ID` | Azure subscription ID | `az account show --query id -o tsv` |
| `AZURE_SUBSCRIPTION_NAME` | Azure subscription name (alternative) | `az account show --query name -o tsv` |

**Note:** Either `AZURE_SUBSCRIPTION_ID` or `AZURE_SUBSCRIPTION_NAME` must be set (or auto-extracted).

---

## Detailed Flow

```
1. Check CI environment:
   └─ CI=true OR GITHUB_ACTIONS=true?
      └─ Yes → SKIP test
      └─ No  → Continue

2. Check AZURE_TENANT_ID:
   └─ Subtest: t.Run("AZURE_TENANT_ID", ...)
   └─ Is AZURE_TENANT_ID set?
      ├─ Yes → Log "AZURE_TENANT_ID is set via environment variable"
      └─ No  → Try auto-extract from Azure CLI
               ├─ Success → os.Setenv("AZURE_TENANT_ID", value)
               │            Log "AZURE_TENANT_ID auto-extracted from Azure CLI"
               └─ Failure → FAIL with remediation instructions

3. Check AZURE_SUBSCRIPTION_ID/NAME:
   └─ Subtest: t.Run("AZURE_SUBSCRIPTION", ...)
   └─ Is AZURE_SUBSCRIPTION_ID or AZURE_SUBSCRIPTION_NAME set?
      ├─ Yes → Log which variable is set
      └─ No  → Try auto-extract AZURE_SUBSCRIPTION_ID from Azure CLI
               ├─ Success → os.Setenv("AZURE_SUBSCRIPTION_ID", value)
               │            Log "AZURE_SUBSCRIPTION_ID auto-extracted from Azure CLI"
               └─ Failure → FAIL with remediation instructions
```

---

## Example Output

### Success (Variables Pre-Set)
```
=== RUN   TestCheckDependencies_AzureEnvironment
=== RUN   TestCheckDependencies_AzureEnvironment/AZURE_TENANT_ID
    01_check_dependencies_test.go:143: AZURE_TENANT_ID is set via environment variable
=== RUN   TestCheckDependencies_AzureEnvironment/AZURE_SUBSCRIPTION
    01_check_dependencies_test.go:177: AZURE_SUBSCRIPTION_ID is set via environment variable
--- PASS: TestCheckDependencies_AzureEnvironment (0.00s)
```

### Success (Auto-Extracted)
```
=== RUN   TestCheckDependencies_AzureEnvironment
=== RUN   TestCheckDependencies_AzureEnvironment/AZURE_TENANT_ID
    01_check_dependencies_test.go:169: AZURE_TENANT_ID auto-extracted from Azure CLI: 12345678...cdef
=== RUN   TestCheckDependencies_AzureEnvironment/AZURE_SUBSCRIPTION
    01_check_dependencies_test.go:212: AZURE_SUBSCRIPTION_ID auto-extracted from Azure CLI: abcdef12...5678
--- PASS: TestCheckDependencies_AzureEnvironment (0.80s)
```

### Skipped (CI Environment)
```
=== RUN   TestCheckDependencies_AzureEnvironment
    01_check_dependencies_test.go:133: Skipping Azure environment validation in CI environment
--- SKIP: TestCheckDependencies_AzureEnvironment (0.00s)
```

### Failure
```
=== RUN   TestCheckDependencies_AzureEnvironment
=== RUN   TestCheckDependencies_AzureEnvironment/AZURE_TENANT_ID
    01_check_dependencies_test.go:151: AZURE_TENANT_ID is not set and could not be extracted from Azure CLI.

To fix this, run:
  export AZURE_TENANT_ID=$(az account show --query tenantId -o tsv)

Error: Please run 'az login' to setup account.
--- FAIL: TestCheckDependencies_AzureEnvironment/AZURE_TENANT_ID (0.30s)
```

---

## Key Observations

- Provides seamless UX for users who are logged in with Azure CLI
- Auto-sets environment variables for subsequent tests using `os.Setenv()`
- Only logs partial values (first 8 and last 4 characters) to avoid exposing full IDs
- Runs cleanup function that prints a summary of missing variables if validation failed
- Skipped in CI environments where Azure env vars may not be set

---

## Remediation Instructions

If auto-extraction fails, run these commands:

```bash
export AZURE_TENANT_ID=$(az account show --query tenantId -o tsv)
export AZURE_SUBSCRIPTION_ID=$(az account show --query id -o tsv)
# OR
export AZURE_SUBSCRIPTION_NAME=$(az account show --query name -o tsv)
```
