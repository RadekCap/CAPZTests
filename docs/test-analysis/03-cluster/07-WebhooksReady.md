# Test 7: TestKindCluster_WebhooksReady

**Location:** `test/03_cluster_test.go:439-535`

**Purpose:** Wait for all admission webhooks (CAPI, CAPZ, ASO, MCE) to be responsive and accepting connections.

---

## Configuration

| Parameter | Value |
|-----------|-------|
| Timeout per webhook | 5 minutes |
| Poll interval | 5 seconds |
| Test method | HTTPS connectivity from ephemeral pod |

---

## Webhooks Checked

| Webhook | Namespace | Service | Port |
|---------|-----------|---------|------|
| CAPI | `capi-system` | `capi-webhook-service` | 443 |
| CAPZ | `capz-system` | `capz-webhook-service` | 443 |
| ASO | `capz-system` | `azureserviceoperator-webhook-service` | 443 |
| MCE | `capi-system` | `mce-capi-webhook-config-service` | 9443 |

---

## Detailed Flow

```
For each webhook:
│
├─► Configuration:
│   └── Timeout: 5 minutes, Poll interval: 5 seconds
│
├─► Loop until timeout:
│   │
│   ├─► Check if endpoint has addresses:
│   │   └─ kubectl get endpoints <service> -n <namespace>
│   │      -o jsonpath={.subsets[0].addresses[0].ip}
│   │      ├─ Empty → Wait and retry
│   │      └─ Has IP → Continue
│   │
│   ├─► Test HTTPS connectivity:
│   │   └─ kubectl run webhook-test-<timestamp> --rm -i
│   │      --restart=Never --image=curlimages/curl:latest
│   │      -- curl -k -s -o /dev/null -w '%{http_code}'
│   │         --connect-timeout 3 --max-time 5
│   │         https://<service>.<namespace>.svc:<port>/
│   │      │
│   │      └─► HTTP code returned?
│   │          ├─ Non-000 → Webhook responsive, PASS
│   │          └─ 000     → Connection failed, retry
│   │
│   └─► Report progress, sleep, repeat
│
└─► Timeout → FAIL
```

---

## Why Check HTTP Codes?

Any HTTP response (even 400, 404, 405) indicates the webhook server is listening:

| HTTP Code | Meaning |
|-----------|---------|
| 000 | Connection failed (server not ready) |
| 400/404/405 | Server is up, just rejecting our test request |
| 200 | Server responding normally |

The test considers any non-000 response as success because we're only testing connectivity, not actual webhook functionality.

---

## Example Output

```
=== Checking webhook readiness ===
Webhooks to verify: 4
Timeout per webhook: 5m0s | Poll interval: 5s

--- Checking CAPI webhook ---
Service: capi-webhook-service.capi-system.svc:443
[1] Waiting for CAPI endpoint to have addresses...
[2] Waiting for CAPI endpoint to have addresses...
[3] 📊 CAPI endpoint IP: 10.244.0.15
[3] ✅ CAPI webhook is responsive (HTTP 404) - took 15s

--- Checking CAPZ webhook ---
Service: capz-webhook-service.capz-system.svc:443
[1] 📊 CAPZ endpoint IP: 10.244.0.18
[1] ✅ CAPZ webhook is responsive (HTTP 404) - took 2s

--- Checking ASO webhook ---
Service: azureserviceoperator-webhook-service.capz-system.svc:443
[1] 📊 ASO endpoint IP: 10.244.0.21
[1] ✅ ASO webhook is responsive (HTTP 400) - took 3s

--- Checking MCE webhook ---
Service: mce-capi-webhook-config-service.capi-system.svc:9443
[1] 📊 MCE endpoint IP: 10.244.0.16
[1] ✅ MCE webhook is responsive (HTTP 404) - took 2s

=== Webhook readiness check complete ===
```

---

## Why This Test Matters

Webhooks are critical for Kubernetes admission control. If webhooks are not ready:

1. Resource creation will fail with timeout errors
2. CR deployment (Phase 5) will fail with cryptic webhook errors
3. Debugging becomes difficult without knowing webhook state

This test ensures all webhooks are responsive before proceeding to CR deployment.

---

## Technical Details

### Ephemeral Pod Approach

The test uses an ephemeral `curlimages/curl` pod to test HTTPS connectivity because:
1. Can't reach cluster services from outside the cluster
2. Self-signed certificates require `-k` flag (insecure)
3. Pod gets automatic cleanup with `--rm` flag
4. Unique pod name (`webhook-test-<timestamp>`) avoids conflicts

### Service DNS

Kubernetes service DNS format:
```
<service>.<namespace>.svc:<port>
```

Examples:
- `capi-webhook-service.capi-system.svc:443`
- `mce-capi-webhook-config-service.capi-system.svc:9443`
