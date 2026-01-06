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

## Positive usage

You can create a scalable test environment by deleting those unused Kubernetes resources.

`kube-killer` is another implement of `serverless`.

## Malicious usage

You can DELETE THE KEY RESOURCE SNEAKILY if your boss have no plan to raise your salary.

![image](/docs/img/rm.gif)

Please do not use it for bad . (🤣~~I bet you will~~)

Just remember:
`Easy to hurt, hard to forgive, just make FUN.`

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

#### kill resource

```bash
# delete "my-wife" deployment after 10 mins
kube-killer kill deploy my-wife -a 10m
kube-killer kill deployment my-wife -a 10m


# delete deployment by label
kube-killer kill deploy -l age=two-hundred
kube-killer kill deployment -l age=two-hundred

# delete deployment by namespace and labels
kube-killer kill deploy -l age=two-hundred -n default
kube-killer kill deployment -l age=two-hundred -n default

```

#### freeze deploy

```bash
# scale “my-girlfriends” deployment’s spec.replicas to 0 now
kube-killer freeze deployment my-girlfriends -a
# scale “my-girlfriends” deployment’s spec.replicas to 0 after 1 hour
kube-killer freeze deployment my-girlfriends -a 1h

```

You can find more examples in my [test cases]()

### CURL usage

You can expose the [kube-killer server](#Web-server-mode) by using nodePort service .

Then the [kube-killer server](#Web-server-mode) would become some kind of backdoor.

Finally，you are free to destroy the whole production Kubernetes cluster  remotely （Hhhhhhhhhhhhhhhhhhhhh).

## Serverless mode

1. [ ] kill node gracefully
1. [ ] kill satan
1. [x] kill completed/failed pod automatically
1. [x] kill unused PV
1. [x] kill unused PVC
1. [x] kill service without pod
1. [x] kill unused configmap
1. [x] kill unused secret
1. [x] kill completed jobs
1. [x] 支持 all-namespaces 标志
1. [x] 支持 interactive 模式
1. [x] 支持 dry-run 模式

### kubectl Plugin

kube-killer can be used as a kubectl plugin! Install it and use `kubectl kill` to delete unused Kubernetes resources.

**Installation:**

```bash
# Build and install the plugin
# 构建插件
make build-kubectl-plugin
# 安装插件
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

### Binary CLI usage

```bash
kube-killer kill po
kube-killer kill pod

```

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
