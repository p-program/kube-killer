# kube-killer 使用文档

`kube-killer` 是一个用于删除未使用的 Kubernetes 资源的工具，参考了 [kubectl-reap](https://github.com/micnncim/kubectl-reap) 的实现。

## 目录

- [安装](#安装)
- [基本用法](#基本用法)
- [支持的资源类型](#支持的资源类型)
- [命令详解](#命令详解)
- [标志选项](#标志选项)
- [使用示例](#使用示例)
- [注意事项](#注意事项)

## 安装

### 从源码编译

```bash
git clone https://github.com/p-program/kube-killer.git
cd kube-killer
make build
```

### 使用二进制文件

编译完成后，可执行文件位于项目根目录的 `kube-killer`。

## 基本用法

```bash
kube-killer [command] [resource-type] [flags]
```

### 可用命令

- `kill` - 删除 Kubernetes 资源
- `freeze` - 将资源缩放到 0（冻结）
- `version` - 显示版本信息
- `scan` - 扫描 Kubernetes 集群中的反模式和问题
- `server` - 作为 Kubernetes Operator 运行

## 支持的资源类型

### kill 命令支持

| 资源类型 | 别名 | 说明 |
|---------|------|------|
| Pod | `p`, `po`, `pod` | 删除非 Running 状态的 Pod |
| ConfigMap | `cm`, `configmap` | 删除未被 Pod 使用的 ConfigMap |
| Secret | `secret`, `secrets` | 删除未被 Pod 或 ServiceAccount 使用的 Secret |
| Service | `s`, `svc`, `service` | 删除没有匹配 Pod 的 Service |
| PersistentVolume | `pv` | 删除未绑定或未使用的 PV |
| PersistentVolumeClaim | `pvc` | 删除未绑定或未被 Pod 使用的 PVC |
| Job | `job`, `jobs` | 删除已完成或失败的 Job |
| Deployment | `d`, `deploy`, `deployment` | 删除 Deployment |
| Node | `n`, `no`, `node` | 优雅地删除 Node（需要指定节点名） |
| Namespace | `ns`, `namespace` | 删除 Namespace |
| CustomResource | `cr`, `customresource` | 根据 group 模式删除 Custom Resource（需要指定 group 模式） |
| CustomResourceDefinition | `crd`, `customresourcedefinition` | 根据 group 模式删除 CustomResourceDefinition（需要指定 group 模式） |

### freeze 命令支持

- `deployment` / `deploy` / `d` - 将 Deployment 的副本数缩放到 0
- `statefulset` / `ss` - 将 StatefulSet 的副本数缩放到 0

## 命令详解

### kill 命令

删除未使用的 Kubernetes 资源。

**语法：**
```bash
kube-killer kill <resource-type> [flags]
```

**资源删除规则：**

1. **Pod** (`pod`, `po`, `p`)
   - 删除所有非 Running 状态的 Pod（包括 Completed、Failed、Evicted、Unknown、Pending 等）
   - 使用 `status.phase!=Running` 字段选择器

2. **ConfigMap** (`configmap`, `cm`)
   - 删除未被任何 Pod 使用的 ConfigMap
   - 检查 Pod 的 volumes、env、envFrom 等配置

3. **Secret** (`secret`, `secrets`)
   - 删除未被 Pod 或 ServiceAccount 使用的 Secret
   - 检查 Pod 的 volumes、env、envFrom、projected volumes
   - 检查 ServiceAccount 的 secrets 和 imagePullSecrets

4. **Service** (`service`, `svc`, `s`)
   - 删除没有匹配 Pod 的 Service（通过 selector 匹配）
   - 保留 Headless Service（ClusterIP 为 None）

5. **PersistentVolume** (`pv`)
   - 删除 Available、Released 或 Failed 状态的 PV
   - 检查是否被任何 PVC 绑定

6. **PersistentVolumeClaim** (`pvc`)
   - 删除未绑定或未被 Pod 使用的 PVC
   - 检查 Pod 的 volumes 配置

7. **Job** (`job`, `jobs`)
   - 删除已完成（Completed）或失败（Failed）的 Job
   - 通过 Job 状态和条件判断

8. **Deployment** (`deployment`, `deploy`, `d`)
   - 删除 Deployment 及其关联资源

9. **Node** (`node`, `no`, `n`)
   - 需要指定节点名称作为第二个参数
   - 先执行 `kubectl cordon`，然后可以执行 `kubectl drain`

10. **Namespace** (`namespace`, `ns`)
    - 删除命名空间及其所有资源
    - 支持强制删除（使用 `--mafia` 标志）

11. **CustomResource** (`cr`, `customresource`)
    - 根据 group 模式删除 Custom Resource
    - 需要指定 group 模式作为第二个参数（例如：`example.com` 或 `*.example.com`）
    - 支持通配符匹配：
      - `*.example.com` - 匹配所有以 `.example.com` 结尾的 group
      - `example.com` - 精确匹配
      - `example.*` - 匹配所有以 `example.` 开头的 group
    - 自动识别 cluster-scoped 和 namespace-scoped 的 CR
    - 自动发现匹配的 CRD 并删除对应的所有 CR

12. **CustomResourceDefinition** (`crd`, `customresourcedefinition`)
    - 根据 group 模式删除 CustomResourceDefinition
    - 需要指定 group 模式作为第二个参数（例如：`example.com` 或 `*.example.com`）
    - 支持通配符匹配：
      - `*.example.com` - 匹配所有以 `.example.com` 结尾的 group
      - `example.com` - 精确匹配
      - `example.*` - 匹配所有以 `example.` 开头的 group
    - CRD 是集群级别的资源，不需要指定命名空间
    - ⚠️ **警告：删除 CRD 会同时删除该 CRD 定义的所有 CR 实例**

### freeze 命令

将资源缩放到 0，相当于 `kubectl scale --replicas=0`。

**语法：**
```bash
kube-killer freeze <resource-type> <resource-name> [flags]
```

**示例：**
```bash
# 冻结 Deployment
kube-killer freeze deployment my-app -n default

# 冻结 StatefulSet
kube-killer freeze statefulset my-statefulset -n default
```

### scan 命令

扫描 Kubernetes 集群中的反模式和问题，基于云原生开发最佳实践。

**语法：**
```bash
kube-killer scan [flags]
```

**扫描内容：**
- CRD schema 问题
- Conversion webhook 问题
- Controller 协调循环问题
- Webhook 配置问题
- Owner reference 问题

### server 命令

作为 Kubernetes Operator 运行，监听 KubeKiller CRD 并管理资源清理。

**语法：**
```bash
kube-killer server run
```

## 标志选项

### kill 命令标志

| 标志 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| `--namespace` | `-n` | 指定命名空间 | `default` |
| `--all-namespaces` | `-A` | 在所有命名空间执行（排除 kube-system） | `false` |
| `--dryrun` | `-d` | 仅显示将要删除的资源，不实际删除 | `false` |
| `--interactive` | `-i` | 交互式模式，删除前询问确认 | `false` |
| `--mafia` | - | 黑手党模式：删除所有资源（忽略使用情况检查） | `false` |
| `--half` | - | 与 `--mafia` 一起使用，随机删除一半的资源 | `false` |

### freeze 命令标志

| 标志 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| `--namespace` | `-n` | 指定命名空间 | `default` |

**注意：** freeze 命令目前不支持 `--dryrun` 标志，但会在实际执行前使用 Kubernetes 的 DryRun 模式进行验证。

### scan 命令标志

| 标志 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| `--namespace` | `-n` | 扫描指定命名空间（默认：所有命名空间） | `""` |
| `--all-namespaces` | `-A` | 扫描所有命名空间 | `false` |
| `--output` | `-o` | 输出格式：`table`、`json`、`yaml` | `table` |

### 全局标志

| 标志 | 说明 | 默认值 |
|------|------|--------|
| `--config` | 配置文件路径 | `""` |

## 使用示例

### 基本用法

```bash
# 删除当前命名空间中非 Running 的 Pod
kube-killer kill pod

# 删除指定命名空间中未使用的 ConfigMap
kube-killer kill configmap -n production

# 删除所有命名空间中未使用的 Secret（排除 kube-system）
kube-killer kill secret --all-namespaces

# 删除已完成的 Job
kube-killer kill job -n default
```

### 使用 Dry-run 模式

```bash
# 查看将要删除的 Pod（不实际删除）
kube-killer kill pod --dryrun

# 查看将要删除的 ConfigMap
kube-killer kill cm -n production --dryrun
```

### 使用交互式模式

```bash
# 交互式删除 ConfigMap，每个资源删除前都会询问
kube-killer kill configmap --interactive

# 结合 dry-run 和 interactive
kube-killer kill secret -n default --dryrun --interactive
```

### 使用黑手党模式（Mafia Mode）

```bash
# ⚠️ 危险：删除所有 Pod（忽略使用情况检查）
kube-killer kill pod --mafia

# ⚠️ 危险：随机删除一半的 ConfigMap
kube-killer kill configmap --mafia --half

# 结合 dry-run 查看将要删除的资源
kube-killer kill pod --mafia --dryrun
```

### 跨命名空间操作

```bash
# 删除所有命名空间（除 kube-system）中未使用的 Service
kube-killer kill service --all-namespaces

# 删除所有命名空间中非 Running 的 Pod
kube-killer kill pod -A
```

### 删除特定资源

```bash
# 删除未使用的 PV（PV 是集群级别的，不需要指定命名空间）
kube-killer kill pv

# 删除未使用的 PVC
kube-killer kill pvc -n default

# 优雅删除 Node（需要指定节点名）
kube-killer kill node worker-node-1

# 删除 Namespace（会删除命名空间及其所有资源）
kube-killer kill namespace my-namespace

# 强制删除 Namespace（使用 mafia 模式）
kube-killer kill namespace my-namespace --mafia
```

### 删除 Custom Resource

```bash
# 删除指定 group 下的所有 CR（精确匹配）
kube-killer kill cr example.com

# 删除所有匹配通配符 group 的 CR（支持通配符）
kube-killer kill cr *.example.com

# 在指定命名空间中删除 CR
kube-killer kill cr example.com -n production

# 在所有命名空间中删除匹配的 CR（排除 kube-system）
kube-killer kill cr *.example.com --all-namespaces

# 使用 dry-run 模式查看将要删除的 CR
kube-killer kill cr example.com --dryrun

# 删除所有以 example. 开头的 group 下的 CR
kube-killer kill cr example.*
```

### 删除 CustomResourceDefinition

```bash
# 删除指定 group 下的所有 CRD（精确匹配）
kube-killer kill crd example.com

# 删除所有匹配通配符 group 的 CRD（支持通配符）
kube-killer kill crd *.example.com

# 使用 dry-run 模式查看将要删除的 CRD
kube-killer kill crd example.com --dryrun

# 删除所有以 example. 开头的 group 下的 CRD
kube-killer kill crd example.*

# CRD 是集群级别的，不需要指定命名空间
kube-killer kill crd *.example.com
```

### Freeze 命令示例

```bash
# 冻结 Deployment（将副本数设为 0）
kube-killer freeze deployment my-app -n production

# 冻结 StatefulSet
kube-killer freeze statefulset my-db -n production

# 使用别名
kube-killer freeze d my-app -n production
kube-killer freeze ss my-db -n production
```

### Scan 命令示例

```bash
# 扫描所有命名空间
kube-killer scan

# 扫描指定命名空间
kube-killer scan -n production

# 扫描所有命名空间并输出 JSON 格式
kube-killer scan --all-namespaces -o json

# 输出 YAML 格式
kube-killer scan -o yaml
```

### 组合使用

```bash
# 删除所有命名空间中未使用的 ConfigMap 和 Secret
kube-killer kill configmap --all-namespaces
kube-killer kill secret --all-namespaces

# 交互式删除所有命名空间中的 completed Job
kube-killer kill job -A -i

# 先扫描问题，再清理资源
kube-killer scan
kube-killer kill pod --dryrun
```

## 注意事项

### ⚠️ 警告

1. **生产环境使用前请务必使用 `--dryrun` 模式测试**
   ```bash
   kube-killer kill pod --dryrun
   ```

2. **`--all-namespaces` 标志会自动排除 `kube-system` 命名空间**
   - 这是为了防止误删系统关键资源
   - 即使手动指定 `-n kube-system`，也不会删除该命名空间的资源

3. **交互式模式 (`-i`)**
   - 启用后，每个资源删除前都会提示确认
   - 输入 `y` 或 `yes` 确认删除，其他输入则跳过

4. **黑手党模式 (`--mafia`)**
   - ⚠️ **极度危险**：会删除所有资源，忽略使用情况检查
   - 建议始终结合 `--dryrun` 使用
   - 与 `--half` 一起使用时，会随机删除一半的资源

5. **资源删除规则**
   - Pod: 只删除非 Running 状态的 Pod
   - ConfigMap/Secret: 只删除未被使用的资源（除非使用 `--mafia`）
   - Service: 只删除没有匹配 Pod 的 Service
   - PV/PVC: 只删除未绑定或未使用的资源
   - Job: 只删除已完成或失败的 Job
   - CustomResource: 根据 group 模式匹配并删除所有对应的 CR（包括 cluster-scoped 和 namespace-scoped）
   - CustomResourceDefinition: 根据 group 模式匹配并删除所有对应的 CRD（⚠️ 删除 CRD 会同时删除该 CRD 定义的所有 CR 实例）

6. **特殊命令**
   - `kube-killer kill me` - ⚠️ **危险命令，请勿使用**
   - `kube-killer kill satan` - ⚠️ **危险命令，请勿使用**

### 最佳实践

1. **首次使用建议：**
   ```bash
   # 1. 先查看将要删除的资源
   kube-killer kill <resource-type> --dryrun
   
   # 2. 使用交互式模式逐个确认
   kube-killer kill <resource-type> --interactive
   
   # 3. 确认无误后再执行实际删除
   kube-killer kill <resource-type>
   ```

2. **定期清理：**
   ```bash
   # 清理所有命名空间中的未使用资源
   kube-killer kill configmap --all-namespaces --dryrun
   kube-killer kill secret --all-namespaces --dryrun
   kube-killer kill pvc --all-namespaces --dryrun
   ```

3. **清理完成的 Job：**
   ```bash
   # 定期清理已完成的 Job
   kube-killer kill job -n default
   ```

4. **清理 Custom Resource：**
   ```bash
   # 清理特定 group 下的所有 CR（建议先使用 dry-run）
   kube-killer kill cr example.com --dryrun
   kube-killer kill cr example.com
   
   # 清理所有匹配通配符的 CR
   kube-killer kill cr *.example.com --all-namespaces --dryrun
   ```

5. **清理 CustomResourceDefinition：**
   ```bash
   # ⚠️ 警告：删除 CRD 会同时删除该 CRD 定义的所有 CR 实例
   # 建议先使用 dry-run 查看将要删除的 CRD
   kube-killer kill crd example.com --dryrun
   
   # 确认无误后再执行实际删除
   kube-killer kill crd example.com
   
   # 清理所有匹配通配符的 CRD
   kube-killer kill crd *.example.com --dryrun
   ```

6. **使用 Scan 命令进行健康检查：**
   ```bash
   # 定期扫描集群中的反模式和问题
   kube-killer scan --all-namespaces
   
   # 将结果导出为 JSON 进行分析
   kube-killer scan -o json > scan-results.json
   ```

### 限制和已知问题

1. **Freeze 命令**
   - 目前不支持 `--dryrun` 标志，但会在实际执行前使用 Kubernetes 的 DryRun 模式进行验证
   - 仅支持 Deployment 和 StatefulSet

2. **Node Killer**
   - Drain 功能可能需要额外配置

3. **Server 命令**
   - 需要部署相应的 CRD 和 RBAC 配置

### 故障排查

1. **权限问题**
   - 确保有足够的 Kubernetes 集群权限
   - 建议使用集群管理员权限

2. **kubeconfig 配置**
   - 确保 `~/.kube/config` 文件存在且配置正确
   - 或通过环境变量 `KUBECONFIG` 指定配置文件

3. **网络问题**
   - 确保能够访问 Kubernetes API Server
   - 检查防火墙和网络策略

4. **资源删除失败**
   - 检查资源是否有 finalizers
   - 检查是否有其他资源依赖
   - 使用 `--interactive` 模式查看详细错误信息

## 参考

- 项目地址: https://github.com/p-program/kube-killer
- 参考项目: https://github.com/micnncim/kubectl-reap
- Kubernetes 官方文档: https://kubernetes.io/docs/

## 许可证

Mulan PSL v2
