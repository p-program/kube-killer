# change logs

## 优化 NodeKiller 实现：使用 Evict API 并完善节点删除流程

2026-01-12

参考 `kubectl drain` 和 `kubectl delete node` 的标准流程，对 `cmd/killer/node_killer.go` 进行了全面优化，使用 Evict API 替代 Delete API，并实现了完整的节点删除流程。

### 主要改进

1. **使用 Evict API 替代 Delete API**
   - 使用 `PolicyV1().Evictions().Evict()` 替代 `Pods().Delete()`
   - 符合 `kubectl drain` 的标准行为
   - 自动尊重 PodDisruptionBudgets（PDB）
   - 支持 pod 的优雅终止（graceful termination）

2. **实现完整的三步节点删除流程**
   - **Step 1: `cordonNode()`** - 标记节点为不可调度（`kubectl cordon`）
     - 设置 `node.Spec.Unschedulable = true`
     - 防止新的 pod 调度到该节点
     - 检查节点是否已处于 cordon 状态，避免重复操作
   - **Step 2: `drainNodePods()`** - 驱逐除 DaemonSet 外的所有 pod（`kubectl drain --ignore-daemonsets`）
     - 使用 Evict API 驱逐所有非 DaemonSet pod
     - 自动跳过 DaemonSet pod
     - 跳过已处于终止状态的 pod
     - 等待所有 pod 优雅终止（最多 5 分钟）
   - **Step 3: `deleteNode()`** - 删除节点（`kubectl delete node`）
     - 删除节点资源
     - 支持 dry-run 模式

3. **添加等待和超时机制**
   - `waitForPodsToTerminate()` 方法等待所有 pod 终止
   - 超时时间：5 分钟
   - 检查间隔：5 秒
   - 智能检测 pod 状态（已删除、已终止、运行中等）

4. **代码结构优化**
   - 将功能拆分为独立的方法：`cordonNode()`、`drainNodePods()`、`evictPod()`、`waitForPodsToTerminate()`、`deleteNode()`
   - 改进错误处理和日志记录
   - 使用 `fmt.Errorf` 包装错误，提供更详细的上下文信息
   - 单个 pod 驱逐失败不会中断整个流程

5. **增强的错误处理**
   - 每个步骤都有独立的错误处理
   - 超时后给出警告而非直接失败
   - 完善的日志输出，便于调试和追踪

### 技术实现

- 使用 `k8s.io/api/policy/v1` 包的 `Eviction` API
- 使用 `clientset.PolicyV1().Evictions(namespace).Evict()` 方法
- 通过 `fieldSelector` 查询节点上的所有 pod
- 检查 pod 的 `OwnerReferences` 判断是否为 DaemonSet pod
- 使用 `DeletionTimestamp` 和 `Status.Phase` 判断 pod 状态

### 使用示例

```bash
# 正常删除节点（cordon -> drain -> delete）
kube-killer kill node my-node-name

# 强制删除节点（跳过 cordon 和 drain，直接 delete）
kube-killer kill node my-node-name --mafia

# 预览模式
kube-killer kill node my-node-name --dryrun
```

### 流程对比

**改进前：**
- 使用 `Delete()` API 直接删除 pod（不够优雅）
- 缺少等待 pod 终止的逻辑
- 缺少删除节点的实现（只有注释）

**改进后：**
- 使用 `Evict()` API 优雅驱逐 pod
- 完整的等待和超时机制
- 实现完整的三步流程：cordon → drain → delete
- 与 `kubectl drain` 行为完全一致

### 参考

- [kubectl drain](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#drain)
- [kubectl cordon](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#cordon)
- [Kubernetes API Eviction](https://kubernetes.io/docs/concepts/scheduling-eviction/api-eviction/)
- [PodDisruptionBudgets](https://kubernetes.io/docs/tasks/run-application/configure-pdb/)

## Operator 模式增强：支持特定命名空间删除和特定时间点执行

2026-01-12

为 operator 模式的 kube-killer 添加了特定命名空间删除和特定时间点执行删除任务的功能，提供了更灵活的调度和执行控制。

### 主要改进

1. **特定命名空间删除功能**
   - 通过 `namespaces` 字段支持指定要操作的命名空间列表
   - 如果设置了 `namespaces`，只会在指定的命名空间中执行删除操作
   - 如果未设置，则默认操作除 `kube-system` 外的所有命名空间（原有行为）
   - 可与 `excludeNamespaces` 配合使用，实现更精确的命名空间控制

2. **特定时间点执行删除任务功能**
   - 新增 `scheduleAt` 字段，支持在指定时间点执行一次性删除任务
   - 使用 RFC3339 格式指定时间（如 `"2026-01-15T10:30:00Z"`）
   - 如果设置了 `scheduleAt`，`interval` 字段将被忽略
   - 任务将在指定时间执行一次后不再重复执行
   - 自动检测任务是否已执行，避免重复执行

3. **智能调度逻辑**
   - 如果当前时间早于 `scheduleAt`，自动计算等待时间并 requeue
   - 如果任务已执行（`LastRunTime >= scheduleTime`），自动跳过
   - 执行完成后不再 requeue，实现一次性任务
   - 完善的日志记录，便于追踪任务执行状态

### 技术实现

- **CRD 扩展**：在 `deploy/operator/crd.yaml` 中添加 `scheduleAt` 字段定义
- **类型定义**：在 `cmd/server/api/v1alpha1/kubekiller_types.go` 中添加 `ScheduleAt *metav1.Time` 字段
- **控制器逻辑**：在 `cmd/server/controllers/kubekiller_controller.go` 中实现时间点检查和调度逻辑
- **示例配置**：在 `deploy/operator/example.yaml` 中添加了三个使用示例

### 使用示例

#### 特定命名空间删除

```yaml
apiVersion: kubekiller.p-program.github.io/v1alpha1
kind: KubeKiller
metadata:
  name: kube-killer-specific-namespace
  namespace: default
spec:
  mode: illidan
  interval: "15m"
  dryRun: false
  # 只操作指定的命名空间
  namespaces:
    - production
    - staging
    - development
  resources:
    - pod
    - job
    - configmap
    - secret
```

#### 特定时间点执行

```yaml
apiVersion: kubekiller.p-program.github.io/v1alpha1
kind: KubeKiller
metadata:
  name: kube-killer-scheduled
  namespace: default
spec:
  mode: illidan
  # scheduleAt 优先于 interval，任务将在指定时间执行一次
  scheduleAt: "2026-01-15T10:30:00Z"
  dryRun: false
  namespaces:
    - production
  resources:
    - pod
    - job
```

#### 组合使用：特定时间点 + 特定命名空间

```yaml
apiVersion: kubekiller.p-program.github.io/v1alpha1
kind: KubeKiller
metadata:
  name: kube-killer-scheduled-namespace
  namespace: default
spec:
  mode: demon
  # 在指定时间点执行，只操作特定命名空间
  scheduleAt: "2026-01-20T02:00:00Z"
  dryRun: false
  namespaces:
    - test
    - dev
```

### 功能特性

1. **向后兼容**：未设置 `scheduleAt` 时，保持原有的 `interval` 定期执行行为
2. **精确控制**：通过 `namespaces` 字段精确控制操作范围
3. **一次性任务**：`scheduleAt` 任务执行后自动停止，不会重复执行
4. **智能调度**：自动计算等待时间，在正确的时间点执行任务
5. **状态追踪**：通过 `LastRunTime` 状态字段追踪任务执行情况

### 参考

- [RFC3339 时间格式](https://tools.ietf.org/html/rfc3339)
- Kubernetes CustomResourceDefinition API
- Controller Runtime Reconcile Loop

## 完成 DeploymentKiller 实现并添加 Half 模式支持

2026-01-12

参考 `kubectl delete deploy` 的实现，完成了 `cmd/killer/deployment_killer.go` 的完整实现，并添加了 half 模式支持。

### 主要改进

1. **完整的 DeploymentKiller 实现**
   - 实现了之前缺失的 `Kill()` 方法
   - 使用 `apps/v1` API（与 kubectl delete deploy 一致）
   - 支持删除命名空间中的所有 deployments
   - 自动级联删除相关的 ReplicaSet 和 Pod（通过 owner reference）

2. **Half 模式支持**
   - 新增 `half` 字段和 `SetHalf()` 方法
   - 实现 `KillHalfDeployments()` 方法
   - 随机打乱 deployments 列表后删除一半（向下取整）
   - 如果只有一个 deployment，至少删除一个

3. **Mafia 模式支持**
   - 支持通过 `--mafia` 标志删除所有 deployments
   - 结合 `--half` 标志可随机删除一半 deployments
   - 更新了 `cmd/killer/kill.go` 中的 deployment 处理逻辑

4. **错误处理和日志**
   - 完善的错误处理和日志记录
   - 使用 `retry.RetryOnConflict` 处理删除冲突
   - 支持 dry-run 模式预览操作
   - 清晰的日志输出，便于调试和追踪

### 技术实现

- 使用 `kubernetes.Clientset.AppsV1().Deployments()` API 操作 deployments
- 使用 `retry.RetryOnConflict` 处理删除时的资源冲突
- 使用 `math/rand` 实现随机打乱列表
- 通过 `metav1.DeleteOptions` 配置删除选项（gracePeriodSeconds、dryRun 等）
- 代码风格与其他 killer（如 `JobKiller`、`ConfigmapKiller`）保持一致

### 使用示例

```bash
# 删除命名空间中的所有 deployments（默认行为）
kube-killer kill deploy --namespace my-namespace

# 删除所有 deployments（mafia 模式）
kube-killer kill deploy --namespace my-namespace --mafia

# 随机删除一半 deployments（mafia + half 模式）
kube-killer kill deploy --namespace my-namespace --mafia --half

# 预览模式
kube-killer kill deploy --namespace my-namespace --dryrun

# 跨所有 namespace 删除
kube-killer kill deploy --all-namespaces --mafia
```

### 参考

- [kubectl delete deployment](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#delete)
- Kubernetes apps/v1 Deployment API

## 新增测试用例

2026-01-12

## 优化 NamespaceKiller 实现

2026-01-12

参考 [knsk](https://github.com/thyarles/knsk) 项目的实现，对 `cmd/killer/namespace_killer.go` 进行了全面优化和改进。

### 主要改进

1. **完整的 Kill() 方法实现**
   - 实现了之前缺失的 `Kill()` 方法
   - 自动检测 namespace 是否处于 Terminating 状态
   - 支持正常删除和强制删除两种模式

2. **处理卡住的 Terminating Namespace**
   - 自动发现并删除 namespace 中的所有资源
   - 使用 dynamic client 动态发现所有 API 资源类型
   - 智能跳过系统资源（events、bindings、endpoints 等）
   - 支持通过 `--mafia` 标志强制移除 finalizers

3. **资源清理机制**
   - 通过 Kubernetes Discovery API 自动发现所有命名空间资源
   - 使用 dynamic client 统一删除各种资源类型
   - 跳过不应删除的系统资源（如 events、serviceaccounts 等）
   - 优雅处理资源访问失败的情况

4. **Force 模式支持**
   - 新增 `Force()` 方法，支持强制删除模式
   - 在 force 模式下自动移除 namespace 的 finalizers
   - 通过 `--mafia` 标志启用 force 模式

5. **错误处理和日志**
   - 完善的错误处理和日志记录
   - 清晰的日志输出，便于调试和追踪
   - 支持 dry-run 模式预览操作

### 技术实现

- 使用 `discoveryClient` 自动发现集群中的所有 API 资源
- 使用 `dynamic.Interface` 统一处理各种资源类型的删除
- 通过 `Namespaces().Finalize()` API 移除 finalizers
- 实现了等待和重试机制，确保删除操作的可靠性

### 使用示例

```bash
# 正常删除 namespace
kube-killer kill namespace my-namespace

# 强制删除卡住的 namespace（移除 finalizers）
kube-killer kill namespace my-namespace --mafia

# 预览模式
kube-killer kill namespace my-namespace --dryrun

# 跨所有 namespace 删除
kube-killer kill namespace --all-namespaces
```

### 参考

- [knsk - Kubernetes namespace killer](https://github.com/thyarles/knsk)
- [Kubernetes Issue #60807](https://github.com/kubernetes/kubernetes/issues/60807)

