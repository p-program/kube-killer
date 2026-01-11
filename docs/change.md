# change logs

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

