# ARO-HCP Resource Creation Analysis

Analysis of resources created during ARO-HCP deployment on Azure via CAPZ/ASO.
Based on live cluster data from namespace `capz-test-20260208-184931`.

---

## 1. Logical Dependency View (what waits for what)

Resources are organized by dependency level. Each level requires the resources above it to exist first. Arrows show `ownerReference` relationships.

```
LEVEL 0 - YAML Inputs (no dependencies, applied by kubectl)
├── Secret/aso-credential                    (credentials.yaml)
├── Secret/cluster-identity-secret           (credentials.yaml)
├── AzureClusterIdentity/cluster-identity    (is.yaml)
├── ResourceGroup/rcape-stage-resgroup       (is.yaml)
└── Cluster/rcape-stage                      (aro.yaml - top-level CAPI resource)

LEVEL 1 - Direct children of Cluster or ResourceGroup
│
├─── owned by Cluster/rcape-stage:
│    ├── AROCluster/rcape-stage                    (infrastructure ref)
│    ├── AROControlPlane/rcape-stage-control-plane  (control plane ref)
│    └── MachinePool/rcape-stage-mp-0              (worker pool)
│
└─── owned by ResourceGroup/rcape-stage-resgroup:
     ├── VirtualNetwork/rcape-stage-vnet
     ├── NetworkSecurityGroup/rcape-stage-nsg
     ├── Vault/rcape-stage-kv
     └── UserAssignedIdentity (x13):
          ├── cp-control-plane
          ├── cp-cluster-api-azure
          ├── cp-cloud-controller-manager
          ├── cp-cloud-network-config
          ├── cp-disk-csi-driver
          ├── cp-file-csi-driver
          ├── cp-image-registry
          ├── cp-ingress
          ├── cp-kms
          ├── dp-disk-csi-driver
          ├── dp-file-csi-driver
          ├── dp-image-registry
          └── service-managed-identity

LEVEL 2 - Children of Level 1 resources
│
├─── owned by VirtualNetwork:
│    └── VirtualNetworksSubnet/rcape-stage-vnet-rcape-stage-subnet
│
├─── owned by MachinePool:
│    └── AROMachinePool/rcape-stage-mp-0
│
├─── RoleAssignments on NSG (owned by NetworkSecurityGroup, x5):
│    ├── cloudcontrollermanagerroleid-nsg
│    ├── hcpcontrolplaneoperatorroleid-nsg
│    ├── filestorageoperatorroleid-nsg (cp)
│    ├── filestorageoperatorroleid-nsg (dp)
│    └── hcpservicemanagedidentityroleid-nsg
│
├─── RoleAssignments on VNet (owned by VirtualNetwork, x3):
│    ├── networkoperatorroleid-vnet
│    ├── hcpcontrolplaneoperatorroleid-vnet
│    └── hcpservicemanagedidentityroleid-vnet
│
├─── RoleAssignments on Vault (owned by Vault, x1):
│    └── keyvaultcryptouserroleid-keyvault
│
└─── RoleAssignments on UserAssignedIdentities ("reader" grants, x12):
     ├── readerroleid-controlplanemi
     ├── readerroleid-clusterapiazuremi
     ├── readerroleid-cloudcontrollermanagermi
     ├── readerroleid-cloudnetworkconfigmi
     ├── readerroleid-diskcsidrivermi
     ├── readerroleid-filecsidrivermi
     ├── readerroleid-imageregistrymi
     ├── readerroleid-ingressmi
     ├── readerroleid-kmsmi
     ├── federatedcredentialsroleid-dpdiskcsidrivermi
     ├── federatedcredentialsroleid-dpfilecsidrivermi
     └── federatedcredentialsroleid-dpimageregistrymi

LEVEL 3 - Children of Level 2 resources
│
├─── RoleAssignments on Subnet (owned by VirtualNetworksSubnet, x7):
│    ├── cloudcontrollermanagerroleid-subnet
│    ├── networkoperatorroleid-subnet
│    ├── hcpclusterapiproviderroleid-subnet
│    ├── filestorageoperatorroleid-subnet (cp)
│    ├── filestorageoperatorroleid-subnet (dp)
│    ├── ingressoperatorroleid-subnet
│    └── hcpservicemanagedidentityroleid-subnet
│
└─── owned by AROControlPlane + ResourceGroup:
     └── HcpOpenShiftCluster/rcape-stage    (the actual Azure HCP cluster)

LEVEL 4 - Final resources (depend on HCP cluster)
│
└─── owned by AROMachinePool + HcpOpenShiftCluster:
     └── HcpOpenShiftClustersNodePool/w-uksouth-mp-0

LEVEL 5 - Controller-generated (identity mappings, kubeconfig)
│
├── ConfigMap/identity-map-* (x13)   - one per managed identity
└── Secret/rcape-stage-kubeconfig    - workload cluster access
```

### Summary of dependency chain:

```
credentials.yaml → is.yaml → aro.yaml
                                │
                    Cluster ────┤
                                ├── AROCluster
                                ├── AROControlPlane ──→ HcpOpenShiftCluster
                                └── MachinePool ──→ AROMachinePool ──→ HcpOpenShiftClustersNodePool

ResourceGroup ──┬── VNet ──→ Subnet ──→ RoleAssignments (x7)
                ├── NSG ──→ RoleAssignments (x5)
                ├── Vault ──→ RoleAssignment (x1)
                └── UserAssignedIdentities (x13) ──→ RoleAssignments (x12)
```

---

## 2. Chronological View (sorted by creation time)

### T+0s (17:52:45) - Initial resource creation from YAML apply

| Resource | Name |
|----------|------|
| Secret | aso-credential |
| Secret | cluster-identity-secret |
| AzureClusterIdentity | cluster-identity |
| ResourceGroup | rcape-stage-resgroup |
| NetworkSecurityGroup | rcape-stage-nsg |
| VirtualNetwork | rcape-stage-vnet |
| VirtualNetworksSubnet | rcape-stage-vnet-rcape-stage-subnet |
| Vault | rcape-stage-kv |
| UserAssignedIdentity | cp-cluster-api-azure |
| UserAssignedIdentity | cp-control-plane |

### T+1s (17:52:46) - CAPI reconciliation creates child resources

| Resource | Name |
|----------|------|
| Cluster | rcape-stage |
| AROCluster | rcape-stage |
| AROControlPlane | rcape-stage-control-plane |
| MachinePool | rcape-stage-mp-0 |
| AROMachinePool | rcape-stage-mp-0 |
| UserAssignedIdentity (x11) | remaining managed identities |
| RoleAssignment (x28) | all role assignments |

### T+2s to T+22s (17:52:47 - 17:53:07) - Azure resources provisioning

| Time | Event |
|------|-------|
| T+2s | ResourceGroup provisioning starts |
| T+6s | ResourceGroup ready |
| T+9s | NSG created in Azure |
| T+9s | VNet created in Azure |
| T+21s | ResourceGroupReady condition = True |
| T+21s | NetworkSecurityGroupsReady = True |
| T+22s | SubnetsReady = True |
| T+22s | VNetReady = True |
| T+22s | VaultReady = True |
| T+23s | UserIdentitiesReady = True |

### T+23s to T+68s (17:53:08 - 17:53:53) - Role assignments reconciled in Azure

| Time | Event |
|------|-------|
| T+23-68s | All 28 RoleAssignments reconciled against Azure |
| T+47s | Identity-map ConfigMaps start appearing |
| T+58s | All identity-map ConfigMaps created (x13) |

### T+112s (17:54:37) - RoleAssignmentReady, HCP cluster creation begins

| Time | Event |
|------|-------|
| T+112s | RoleAssignmentReady = True |
| T+112s | HcpOpenShiftCluster/rcape-stage created |
| T+113s | Secret/rcape-stage-kubeconfig generated |

### T+10m+ (18:03:28) - HCP cluster ready

| Time | Event |
|------|-------|
| ~T+10m43s | HcpClusterReady = True (Succeeded) |
| ~T+10m53s | HcpOpenShiftClustersNodePool/w-uksouth-mp-0 created |

---

## Resource Count Summary

| Category | Count |
|----------|-------|
| CAPI resources (Cluster, AROCluster, AROControlPlane, MachinePool, AROMachinePool) | 5 |
| Azure infra (ResourceGroup, VNet, Subnet, NSG, Vault) | 5 |
| Managed Identities (UserAssignedIdentity) | 13 |
| Role Assignments | 28 |
| Azure HCP resources (HcpOpenShiftCluster, NodePool) | 2 |
| Secrets | 3 |
| ConfigMaps (identity maps + kube-root-ca) | 14 |
| AzureClusterIdentity | 1 |
| **Total** | **71** |

---

## Timeline Summary

```
0s        Apply credentials.yaml + is.yaml + aro.yaml
          ├── Secrets, AzureClusterIdentity, ResourceGroup, VNet, NSG, Vault, Subnet created
1s        CAPI reconciliation: Cluster → AROCluster, AROControlPlane, MachinePool
          └── All 13 UserAssignedIdentities + 28 RoleAssignments created
~22s      Azure infra ready: ResourceGroup, VNet, Subnet, NSG, Vault, Identities
~68s      All RoleAssignments reconciled in Azure
~112s     All conditions met → HcpOpenShiftCluster created (Azure API call)
~10m43s   HcpClusterReady = True
~10m53s   NodePool created → cluster fully operational
```
