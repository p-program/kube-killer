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
| Deployment | `d`, `deploy`, `deployment` | 删除 Deployment（功能待完善） |
| Node | `n`, `no`, `node` | 优雅地删除 Node（需要指定节点名） |
| Namespace | `ns`, `namespace` | 删除 Namespace（功能待完善） |

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

8. **Node** (`node`, `no`, `n`)
   - 需要指定节点名称作为第二个参数
   - 先执行 `kubectl cordon`，然后可以执行 `kubectl drain`（功能待完善）

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

## 标志选项

### kill 命令标志

| 标志 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| `--namespace` | `-n` | 指定命名空间 | `default` |
| `--all-namespaces` | `-A` | 在所有命名空间执行（排除 kube-system） | `false` |
| `--dryrun` | `-d` | 仅显示将要删除的资源，不实际删除 | `false` |
| `--interactive` | `-i` | 交互式模式，删除前询问确认 | `false` |

### freeze 命令标志

| 标志 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| `--namespace` | `-n` | 指定命名空间 | `default` |
| `--dryrun` | `-d` | 仅显示将要执行的操作，不实际执行 | `false` |

### 全局标志

| 标志 | 说明 | 默认值 |
|------|------|--------|
| `--config` | 配置文件路径 | `$HOME/.config.yaml` |
| `--viper` | 使用 Viper 进行配置 | `true` |

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
```

### Freeze 命令示例

```bash
# 冻结 Deployment（将副本数设为 0）
kube-killer freeze deployment my-app -n production

# 冻结 StatefulSet
kube-killer freeze statefulset my-db -n production

# Dry-run 模式查看将要执行的操作
kube-killer freeze deployment my-app -n production --dryrun
```

### 组合使用

```bash
# 删除所有命名空间中未使用的 ConfigMap 和 Secret
kube-killer kill configmap --all-namespaces
kube-killer kill secret --all-namespaces

# 交互式删除所有命名空间中的 completed Job
kube-killer kill job -A -i
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

4. **资源删除规则**
   - Pod: 只删除非 Running 状态的 Pod
   - ConfigMap/Secret: 只删除未被使用的资源
   - Service: 只删除没有匹配 Pod 的 Service
   - PV/PVC: 只删除未绑定或未使用的资源
   - Job: 只删除已完成或失败的 Job

5. **特殊命令**
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

### 限制和已知问题

1. **Deployment Killer** - 功能尚未完全实现
2. **Namespace Killer** - Kill() 方法尚未实现
3. **Node Killer** - Drain 功能尚未完全实现
4. **Custom Metrics** - 自定义指标条件支持尚未实现
5. **Event Output** - 事件输出功能尚未实现

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

## 参考

- 项目地址: https://github.com/p-program/kube-killer
- 参考项目: https://github.com/micnncim/kubectl-reap
- Kubernetes 官方文档: https://kubernetes.io/docs/

## 许可证

Mulan PSL v2

