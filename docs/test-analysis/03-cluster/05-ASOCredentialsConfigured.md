# Test 5: TestKindCluster_ASOCredentialsConfigured

**Location:** `test/03_cluster_test.go:284-382`

**Purpose:** Validate that ASO controller has Azure credentials configured correctly. Currently **SKIPPED** because it requires service principal credentials.

---

## Current Status

**SKIPPED**: This test is currently skipped with the message:
```
Skipped: ASO credentials check requires service principal (AZURE_CLIENT_ID/SECRET) - not yet configured
```

---

## When Enabled, This Test Validates:

| Secret Field | Required | Purpose |
|--------------|----------|---------|
| `AZURE_TENANT_ID` | Yes | Azure tenant identifier |
| `AZURE_SUBSCRIPTION_ID` | Yes | Azure subscription identifier |
| `AZURE_CLIENT_ID` | Yes (one of) | Service principal client ID |
| `AZURE_CLIENT_SECRET` | Yes (one of) | Service principal client secret |

---

## Detailed Flow (When Enabled)

```
1. Check if aso-controller-settings secret exists:
   └─ kubectl get secret aso-controller-settings -n capz-system
      ├─ Success → Continue
      └─ Failure → FAIL: "Secret not found"

2. Check required credential fields:
   └─ For each field in [AZURE_TENANT_ID, AZURE_SUBSCRIPTION_ID]:
      └─ kubectl get secret ... -o jsonpath={.data.<field>}
         ├─ Non-empty → Log "configured"
         └─ Empty     → Add to missing list

3. Check authentication method:
   └─ For each field in [AZURE_CLIENT_ID, AZURE_CLIENT_SECRET]:
      └─ kubectl get secret ... -o jsonpath={.data.<field>}
         ├─ Non-empty → Auth configured = true
         └─ Empty     → Continue

4. Report results:
   └─ Any missing?
      ├─ Yes → FAIL with remediation steps
      └─ No  → PASS
```

---

## Why Is This Test Skipped?

ASO in Kind clusters requires service principal credentials (`AZURE_CLIENT_ID` and `AZURE_CLIENT_SECRET`), which are:

1. Not always available in development environments
2. Require Azure AD application registration
3. May have security implications for local testing

**TODO:** Re-enable when service principal setup is documented or made optional.

---

## Remediation (If Test Were Enabled and Failing)

```bash
# Ensure Azure CLI is logged in
az login

# Set environment variables
export AZURE_TENANT_ID=$(az account show --query tenantId -o tsv)
export AZURE_SUBSCRIPTION_ID=$(az account show --query id -o tsv)

# For service principal auth (required for ASO):
export AZURE_CLIENT_ID=<your-service-principal-client-id>
export AZURE_CLIENT_SECRET=<your-service-principal-client-secret>

# Re-run the deployment script
```

---

## Example Output

### Current (Skipped)
```
=== RUN   TestKindCluster_ASOCredentialsConfigured
    03_cluster_test.go:292: Skipped: ASO credentials check requires service principal (AZURE_CLIENT_ID/SECRET) - not yet configured
--- SKIP: TestKindCluster_ASOCredentialsConfigured (0.00s)
```

---

## Related Tests

This test runs **BEFORE** `TestKindCluster_ASOControllerReady` to provide fast failure and clear error messages if credentials are missing, rather than waiting 10 minutes for ASO to timeout.
