`To be or not to be,that's your question`

# kube-killer

[中文文档](README_ZH.md)

## Inspiration burst

During the Cold War between one of my female friends / (ex-)girlfriends ( She said that we were done ) in these days , there was a crazy idea flashed through my mind:

**How about deleting ALL resources in production Kubernetes cluster environment without reasons?**

So, I create such a **super Kubernetes virus**, I would like to call it [kube-killer](https://github.com/p-program/kube-killer).

## What is it?

🤣 `kube-killer` is a tool helping you kill (unused) kubernetes‘s resource.

You can delete kubernetes‘s resource （deploy，pod，statefulset and so on） based on time schedule ⏰,
custom metrics or custom condition.

It is a humane killer, he could also freeze the deploy without killing it (scale to 0）.

You could run as web server, binary and CLI mode.

## Architecture

A long run web server using endless loop.

First of all,please make sure that:

1. You are the master of the MYSQL . `root` is the best! And you should make sure that the MYSQL database is reachable for the remote Kubernetes cluster.
1. You are the administrator of the Kubernetes cluster. Admin of the “kube-system” will be the best！

## Positive Usage

You can create a scalable test environment by deleting those unused Kubernetes resources.

`kube-killer` is another implementation of `serverless` - automatically cleaning up unused resources to keep your cluster lean and efficient.

## Malicious usage

## Malicious Usage

You can DELETE KEY RESOURCES SNEAKILY if your boss has no plan to raise your salary.

![image](/docs/img/rm.gif)

Please do not use it for bad purposes. (🤣~~I bet you will~~)

Just remember:
`Easy to hurt, hard to forgive, just make FUN.`

**⚠️ WARNING:** Always use `--dry-run` mode first to preview what will be deleted. Use `--interactive` mode for additional safety. The `--mafia` flag will delete ALL resources regardless of their state - use with extreme caution!

## Server mode

kube-killer can run as a Kubernetes Operator that watches `KubeKiller` Custom Resources and automatically cleans up Kubernetes resources based on the configured mode and schedule.

### Installation

1. Install the CRD:

```bash
kubectl apply -f deploy/operator/crd.yaml
```

2. Deploy the Operator:

```bash
kubectl apply -f deploy/operator/rbac.yaml
kubectl apply -f deploy/operator/deployment.yaml
```

3. Verify the operator is running:

```bash
kubectl get pods -n kube-system | grep kube-killer
```

### Usage

After deploying the operator, you can create `KubeKiller` resources to control the cleanup behavior.

#### Demon Mode

Demon mode will **KILL ALL PODS** at every interval. It is unstoppable and will kill any pod that gets created.

```yaml
apiVersion: kubekiller.p-program.github.io/v1alpha1
kind: KubeKiller
metadata:
  name: kube-killer-demon
  namespace: default
spec:
  mode: demon
  interval: "5m"
  dryRun: false
```

Apply it:

```bash
kubectl apply -f - <<EOF
apiVersion: kubekiller.p-program.github.io/v1alpha1
kind: KubeKiller
metadata:
  name: kube-killer-demon
  namespace: default
spec:
  mode: demon
  interval: "5m"
  dryRun: false
EOF
```

**⚠️ WARNING**: Demon mode will kill ALL pods in all namespaces (except kube-system) at the specified interval. Use with extreme caution!

#### Illidan Mode

Illidan mode hunts unhealthy Kubernetes resources at every period. It will clean up:
- Completed/Failed pods
- Completed jobs
- Unused PVCs and PVs
- Services without pods
- Unused ConfigMaps and Secrets

```yaml
apiVersion: kubekiller.p-program.github.io/v1alpha1
kind: KubeKiller
metadata:
  name: kube-killer-illidan
  namespace: default
spec:
  mode: illidan
  interval: "10m"
  dryRun: false
  resources:
    - pod
    - job
    - pvc
    - pv
    - service
    - configmap
    - secret
  excludeNamespaces:
    - kube-system
    - kube-public
    - kube-node-lease
```

Apply it:

```bash
kubectl apply -f deploy/operator/example.yaml
```

### Configuration Options

The `KubeKiller` CRD supports the following configuration:

- **mode**: Operation mode - `demon` (kills all pods) or `illidan` (hunts unhealthy resources)
- **interval**: How often the killer should run (e.g., "5m", "1h", "30s")
- **namespaces**: Specific namespaces to operate on (empty means all except kube-system)
- **excludeNamespaces**: Namespaces to exclude from operations
- **dryRun**: If true, only log what would be deleted without actually deleting
- **resources**: Resource types to kill (only used in illidan mode): pod, job, pvc, pv, service, configmap, secret

### Monitoring

Check the status of your KubeKiller resources:

```bash
kubectl get kubekiller
kubectl describe kubekiller kube-killer-illidan
```

The status will show:
- `lastRunTime`: When the killer last ran
- `lastRunResult`: Result of the last run
- `resourcesKilled`: Number of resources killed
- `phase`: Current phase (Ready, Running, Error)

### CLI usage

Once the [kube-killer server](#Web-server-mode) is ready，you can use the CLI mode .

#### Supported Resource Types

kube-killer supports killing the following Kubernetes resources:

- **Pod** (`pod`, `po`, `p`) - Deletes completed/failed pods or all pods (in mafia mode)
- **ConfigMap** (`configmap`, `cm`) - Deletes unused ConfigMaps
- **Secret** (`secret`, `secrets`) - Deletes unused Secrets
- **Service** (`service`, `svc`, `s`) - Deletes services without pods
- **PVC** (`pvc`) - Deletes unbound or unused PersistentVolumeClaims
- **PV** (`pv`) - Deletes unused PersistentVolumes (cluster-scoped)
- **Job** (`job`, `jobs`) - Deletes completed/failed jobs
- **Node** (`node`, `n`, `no`) - Cordon and drain a node (requires node name)
- **Namespace** (`namespace`, `ns`) - Deletes a namespace and all its resources
- **StatefulSet** (`statefulset`, `sts`) - Deletes StatefulSets
- **Deployment** (`deployment`, `deploy`, `d`) - Deletes deployments
- **CustomResource** (`cr`, `customresource`) - Deletes Custom Resources by group pattern (e.g., `*.example.com`)
- **CustomResourceDefinition** (`crd`, `customresourcedefinition`) - Deletes CRDs by group pattern

#### Kill Resources

```bash
# Delete unused pods in default namespace
kube-killer kill pod

# Delete unused pods in all namespaces (except kube-system)
kube-killer kill pod -A

# Delete unused pods with dry-run mode (preview only)
kube-killer kill pod -d

# Delete unused pods with interactive confirmation
kube-killer kill pod -i

# Delete all pods (mafia mode)
kube-killer kill pod --mafia

# Delete half of the pods randomly (mafia + half mode)
kube-killer kill pod --mafia --half

# Delete unused ConfigMaps
kube-killer kill configmap -n default

# Delete unused Secrets
kube-killer kill secret -A

# Delete unused Services
kube-killer kill service -n default

# Delete unused PVCs
kube-killer kill pvc -A

# Delete unused PVs (cluster-scoped, no namespace needed)
kube-killer kill pv

# Delete completed/failed Jobs
kube-killer kill job -n default

# Delete a StatefulSet
kube-killer kill statefulset my-sts -n default

# Delete a Deployment
kube-killer kill deployment my-app -n default

# Cordon and drain a node
kube-killer kill node my-node-name

# Delete a namespace (and all its resources)
kube-killer kill namespace my-namespace

# Delete a namespace forcefully (removes finalizers if stuck)
kube-killer kill namespace my-namespace --mafia

# Delete Custom Resources by group pattern
kube-killer kill cr "*.example.com" -n default

# Delete Custom Resources in all namespaces
kube-killer kill cr "*.example.com" -A

# Delete CustomResourceDefinitions by group pattern
kube-killer kill crd "*.example.com"
```

#### Command Flags

- `-n, --namespace`: Target namespace (default: "default")
- `-A, --all-namespaces`: Operate on all namespaces (except kube-system)
- `-d, --dryrun`: Dry-run mode - preview what would be deleted without actually deleting
- `-i, --interactive`: Interactive mode - prompt for confirmation before deleting each resource
- `--mafia`: Mafia mode - kill all resources regardless of their state
- `--half`: Half mode - when used with `--mafia`, randomly delete half of the resources

#### Freeze Deployments

```bash
# scale “my-girlfriends” deployment’s spec.replicas to 0 now
kube-killer freeze deployment my-girlfriends -a
# scale “my-girlfriends” deployment’s spec.replicas to 0 after 1 hour
kube-killer freeze deployment my-girlfriends -a 1h

```

You can find more examples in the [test cases](https://github.com/p-program/kube-killer/tree/main/cmd/killer)

### CURL Usage

You can expose the [kube-killer server](#Web-server-mode) by using a NodePort service.

Then the [kube-killer server](#Web-server-mode) would become some kind of backdoor.

Finally, you are free to destroy the whole production Kubernetes cluster remotely (Hhhhhhhhhhhhhhhhhhhhh).

## Serverless Mode

The following features are implemented:

1. [x] Kill completed/failed pods automatically
1. [x] Kill unused PVs
1. [x] Kill unused PVCs
1. [x] Kill services without pods
1. [x] Kill unused ConfigMaps
1. [x] Kill unused Secrets
1. [x] Kill completed/failed jobs
1. [x] Kill StatefulSets
1. [x] Kill Deployments
1. [x] Kill nodes gracefully (cordon and drain)
1. [x] Kill namespaces (including stuck Terminating namespaces)
1. [x] Kill Custom Resources by group pattern
1. [x] Kill CustomResourceDefinitions by group pattern
1. [x] Support `--all-namespaces` flag
1. [x] Support `--interactive` mode
1. [x] Support `--dry-run` mode
1. [x] Support `--mafia` mode (kill all resources)
1. [x] Support `--half` mode (randomly kill half of resources)
1. [ ] Kill satan (⚠️ Do not use)

### kubectl Plugin

kube-killer can be used as a kubectl plugin! Install it and use `kubectl kill` to delete unused Kubernetes resources.

**Installation:**

```bash
# Build and install the plugin
make build-kubectl-plugin
make install-kubectl-plugin

# Or install to a custom location
make install-kubectl-plugin PREFIX=/usr/local/bin
```

**Usage:**

```bash
# Delete unused pods
kubectl kill pod

# Delete unused pods in all namespaces
kubectl kill pod -A

# Dry run to see what would be deleted
kubectl kill pod -d

# Delete unused services
kubectl kill service -n default
```

For more details, see the [kubectl plugin documentation](docs/KUBECTL_PLUGIN.md).

### Binary CLI Usage

```bash
# Delete unused pods
kube-killer kill pod

# Delete unused pods in all namespaces
kube-killer kill pod -A

# Delete with dry-run mode
kube-killer kill pod -d

# Delete with interactive mode
kube-killer kill pod -i

# Delete all pods (mafia mode)
kube-killer kill pod --mafia

# Delete half of pods randomly
kube-killer kill pod --mafia --half
```

For more examples, see the [CLI usage](#cli-usage) section above.

## Bazinga Punk

[!["Bazinga Punk!" - Sheldon Cooper - The Big Bang Theory](http://img.youtube.com/vi/HS7YZhsjRAo/0.jpg)](http://www.youtube.com/watch?v=HS7YZhsjRAo)

```bash
kube-killer kill me
```

**!!!WARNING!!!**:PLEASE DO NOT USE.

It‘s an unpredictable command🤣.

## One more thing

Coding is easy, but it is really hard to figure out why she is so angry.

Should I tell her to DRINK MORE HOT WATER ?

![image](/docs/img/hot-water.png)

What exactly does she want? Please tell me if you know the answer. Thank you very much！
