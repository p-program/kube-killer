# Kube-Killer Operator

This directory contains the Kubernetes Operator implementation for kube-killer.

## Quick Start

1. **Install the CRD**:
   ```bash
   kubectl apply -f crd.yaml
   ```

2. **Deploy RBAC and Operator**:
   ```bash
   kubectl apply -f rbac.yaml
   kubectl apply -f deployment.yaml
   ```

3. **Verify the operator is running**:
   ```bash
   kubectl get pods -n kube-system | grep kube-killer
   kubectl logs -n kube-system -l app=kube-killer-operator
   ```

4. **Create a KubeKiller resource**:
   ```bash
   kubectl apply -f example.yaml
   ```

5. **Check the status**:
   ```bash
   kubectl get kubekiller
   kubectl describe kubekiller kube-killer-illidan
   ```

## Files

- `crd.yaml`: Custom Resource Definition for KubeKiller
- `rbac.yaml`: ServiceAccount, Roles, and RoleBindings for the operator
- `deployment.yaml`: Deployment manifest for the operator
- `example.yaml`: Example KubeKiller resources (demon and illidan modes)

## Modes

### Demon Mode
Kills ALL pods in all namespaces (except kube-system) at the specified interval.

### Illidan Mode
Hunts and kills unhealthy resources:
- Completed/Failed pods
- Completed jobs
- Unused PVCs and PVs
- Services without pods
- Unused ConfigMaps and Secrets

## Building

To build the operator:

```bash
make build
```

The operator will be started with:

```bash
./kube-killer server run
```

