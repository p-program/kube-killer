# kubectl-kill Plugin

`kubectl-kill` 是一个 kubectl 插件，提供了 kube-killer 的 Serverless 功能，可以直接通过 `kubectl kill` 命令来删除未使用的 Kubernetes 资源。

## 安装

### 方法 1: 使用 Makefile（推荐）

```bash
# 构建插件
make build-kubectl-plugin

# 安装到 ~/bin（默认）
make install-kubectl-plugin

# 或安装到 /usr/local/bin
make install-kubectl-plugin PREFIX=/usr/local/bin
```

### 方法 2: 手动安装

```bash
# 构建插件
go build -o kubectl-kill ./cmd/kubectl-kill

# 复制到 PATH 中的目录
cp kubectl-kill ~/bin/
chmod +x ~/bin/kubectl-kill

# 确保 ~/bin 在 PATH 中
export PATH=$PATH:$HOME/bin
```

### 方法 3: 使用 krew（如果可用）

```bash
# 如果使用 krew 管理 kubectl 插件
kubectl krew install kill
```

## 验证安装

安装完成后，运行以下命令验证插件是否正常工作：

```bash
kubectl kill --help
```

如果看到帮助信息，说明插件安装成功。

## 使用方法

### 基本用法

```bash
# 删除未使用的 Pod
kubectl kill pod

# 删除未使用的 Pod（指定命名空间）
kubectl kill pod -n default

# 删除所有命名空间（除 kube-system）中的未使用 Pod
kubectl kill pod -A

# 干运行模式（只显示将要删除的资源，不实际删除）
kubectl kill pod -d

# 交互模式（删除前确认）
kubectl kill pod -i
```

### 支持的资源类型

- `pod`, `po`, `p` - 删除 Completed/Failed 状态的 Pod
- `deployment`, `deploy`, `d` - 删除 Deployment
- `service`, `svc`, `s` - 删除没有 Pod 的 Service
- `pvc` - 删除未使用的 PVC
- `pv` - 删除未使用的 PV
- `job`, `jobs` - 删除 Completed/Failed 的 Job
- `configmap`, `cm` - 删除未使用的 ConfigMap
- `secret`, `secrets` - 删除未使用的 Secret
- `node`, `no`, `n` - 删除 Node（需要指定节点名称）

### 常用命令示例

```bash
# 删除所有命名空间中未使用的 Pod
kubectl kill pod -A

# 删除 default 命名空间中未使用的 Service
kubectl kill service -n default

# 删除所有未使用的 PVC（所有命名空间）
kubectl kill pvc -A

# 删除未使用的 PV（集群级别，不需要命名空间）
kubectl kill pv

# 删除未使用的 Secret（交互模式）
kubectl kill secret -A -i

# 干运行：查看将要删除的 ConfigMap
kubectl kill configmap -n default -d
```

## 选项说明

- `-n, --namespace string`: 指定工作命名空间（默认: "default"）
- `-A, --all-namespaces`: 如果为 true，删除所有命名空间（除 kube-system）中的目标资源
- `-d, --dryrun`: 干运行模式，只显示将要删除的资源，不实际删除
- `-i, --interactive`: 交互模式，删除前会提示确认

## 注意事项

1. **权限要求**: 需要 Kubernetes 集群的管理员权限
2. **谨慎使用**: 删除操作不可逆，建议先使用 `-d` 参数进行干运行
3. **命名空间保护**: `kube-system` 命名空间默认被排除，不会被删除
4. **资源类型**: 不同资源类型的删除逻辑不同：
   - Pod: 删除 Completed/Failed 状态的 Pod
   - Service: 删除没有关联 Pod 的 Service
   - PVC/PV: 删除未使用的存储资源
   - Job: 删除 Completed/Failed 的 Job

## 故障排除

### 插件未找到

如果运行 `kubectl kill` 时提示 "kubectl: 'kill' is not a kubectl command"，请检查：

1. 插件文件是否在 PATH 中：
   ```bash
   which kubectl-kill
   ```

2. 插件文件是否有执行权限：
   ```bash
   ls -l $(which kubectl-kill)
   ```

3. 确保插件文件名正确：必须是 `kubectl-kill`（不是 `kubectl-kill.exe` 或其他）

### 权限错误

如果遇到权限错误，请确保：

1. kubeconfig 配置正确
2. 当前用户有足够的 Kubernetes 权限
3. 可以运行 `kubectl get pods` 等基本命令

## 与 kube-killer CLI 的区别

- `kube-killer kill` - 原始 CLI 命令
- `kubectl kill` - kubectl 插件形式，功能相同但更符合 kubectl 使用习惯

两者功能相同，只是调用方式不同。

