# Kube-Killer Scan 命令文档

## 概述

`scan` 命令用于扫描 Kubernetes 集群中的反模式（anti-patterns）和潜在问题。该命令基于 [Cloud Native Development Best Practices](../docs/Cloud-Native-Development-Best-Practices.zh.md) 文档中提到的反面教材，帮助识别和修复常见的 Kubernetes 资源配置问题。

## 使用方法

```bash
# 扫描所有命名空间
kube-killer scan --all-namespaces

# 扫描特定命名空间
kube-killer scan --namespace default

# 输出为 JSON 格式
kube-killer scan --output json

# 输出为 YAML 格式
kube-killer scan --output yaml

# 默认输出为表格格式
kube-killer scan
```

### 命令行参数

- `--namespace, -n`: 指定要扫描的命名空间（默认：空，扫描所有命名空间）
- `--all-namespaces, -A`: 扫描所有命名空间（排除系统命名空间）
- `--output, -o`: 输出格式，可选值：`table`（默认）、`json`、`yaml`

## 已实现的功能

### 1. CRD 扫描器 (CRD Scanner)

检测 CustomResourceDefinition 中的以下问题：

#### ❌ CRD 没有 Schema（Kubernetes 1.17-）
- **问题描述**: CRD 版本中缺少 OpenAPI schema 或 schema 为空
- **严重程度**: Error
- **影响**: 允许无效数据，绕过验证
- **建议**: 为 CRD 版本添加完整的 OpenAPI schema，并设置 `preserveUnknownFields: false`

#### ⚠️ CRD 没有 Conversion Webhook
- **问题描述**: CRD 有多个版本但未配置 conversion webhook
- **严重程度**: Warning
- **影响**: 版本迁移需要手动更新 YAML
- **建议**: 考虑添加 conversion webhook 进行版本迁移，或使用单版本策略

#### ⚠️ Status 字段可能在 Spec 中
- **问题描述**: CRD 的 spec 中可能包含状态相关字段（如 ready、phase、state）
- **严重程度**: Warning
- **影响**: 违反了 Kubernetes 资源设计原则
- **建议**: 将状态字段移至 status 子资源。Spec 应只包含期望状态

#### ❌ preserveUnknownFields 启用
- **问题描述**: CRD 启用了 preserveUnknownFields（已废弃）
- **严重程度**: Error
- **影响**: 允许未知字段，绕过验证
- **建议**: 设置 `preserveUnknownFields: false` 并定义正确的 schema

### 2. Webhook 扫描器 (Webhook Scanner)

检测 ValidatingWebhookConfiguration 和 MutatingWebhookConfiguration 中的以下问题：

#### ⚠️ Webhook Timeout 过短
- **问题描述**: Webhook 的 `timeoutSeconds` 设置为 1 秒或更少
- **严重程度**: Warning
- **影响**: 在高负载下可能导致超时
- **建议**: 将 `timeoutSeconds` 增加到至少 10-30 秒，或对非关键验证设置 `failurePolicy: Ignore`

#### ℹ️ 未使用 cert-manager
- **问题描述**: Webhook 配置中未检测到 cert-manager 注解
- **严重程度**: Info
- **影响**: 证书管理需要手动操作
- **建议**: 考虑使用 cert-manager 自动管理 webhook 证书：
  ```bash
  kubectl annotate validatingwebhookconfiguration <name> cert-manager.io/inject-ca-from=<namespace>/<certificate>
  ```

#### ⚠️ 短 Timeout 配合 Fail 策略
- **问题描述**: Webhook 有短超时时间但 `failurePolicy` 设置为 `Fail`
- **严重程度**: Warning
- **影响**: 可能阻塞 API 操作
- **建议**: 对于非关键验证，考虑设置 `failurePolicy: Ignore`，或增加 `timeoutSeconds`

### 3. Controller 扫描器 (Controller Scanner)

检测 Deployment、Job 等控制器资源中的以下问题：

#### ⚠️ 潜在的 Reconcile 循环
- **问题描述**: Deployment 可能包含在 Reconcile 中更新自身的模式（通过标签/注解启发式检测）
- **严重程度**: Warning
- **影响**: 可能导致无限循环，消耗资源
- **建议**: 审查控制器代码：避免在 Reconcile 中对触发 Reconcile 的同一资源调用 `Update()`。使用 `Patch()` 并进行适当的比较

#### ⚠️ 可能的事件滥用
- **问题描述**: Deployment 可能在每次 reconcile 时都生成事件，而不检查实际变化
- **严重程度**: Warning
- **影响**: 产生大量事件，可能影响 etcd 性能
- **建议**: 仅在状态实际发生变化时发出事件。在发出事件之前使用 `reflect.DeepEqual()` 比较新旧状态

#### ❌ 极长的 RequeueAfter 时间
- **问题描述**: Job 的 `RequeueAfter` 可能设置为极长的持续时间（例如 1000000000 年）
- **严重程度**: Error
- **影响**: 资源可能永远不会被重新处理
- **建议**: 审查控制器代码并设置合理的 `RequeueAfter` 值（秒、分钟或小时，而不是年）

### 4. Owner Reference 扫描器 (Owner Reference Scanner)

检测 Pod、ConfigMap 等资源中的以下问题：

#### ❌ 有问题的 Owner Reference 配置
- **问题描述**: Pod 的 owner reference 指向不应该直接拥有 Pod 的资源类型（如 ConfigMap、Secret）
- **严重程度**: Error
- **影响**: 可能导致级联删除问题（例如，子资源删除时父资源也被删除）
- **建议**: 审查 owner reference 设置。确保父资源不依赖于子资源的存在。正确使用 `controllerutil.SetControllerReference()`

#### ⚠️ ConfigMap 的 Owner Reference 指向子资源
- **问题描述**: ConfigMap 的 owner reference 指向依赖它的资源（如 Pod）
- **严重程度**: Warning
- **影响**: 反向依赖关系可能导致意外的资源删除
- **建议**: 审查 owner reference 层次结构。父资源应该拥有子资源，而不是相反

### 5. Kubernetes 版本适配

- **自动版本检测**: 扫描器会自动检测 Kubernetes 集群版本
- **版本特定逻辑**: 根据检测到的版本调整检测逻辑（例如，Kubernetes 1.17- 的 schema 检查）

## 输出格式

### 表格格式（默认）

```
====================================================================================================
KUBE-KILLER SCAN RESULTS
====================================================================================================
Total Issues Found: 5

📁 Category: CRD (2 issues)
----------------------------------------------------------------------------------------------------

[1] ❌ CRD without schema (Kubernetes 1.17-)
   Resource: CustomResourceDefinition/example.com
   Description: CRD example.com has empty or missing schema. This is unsafe and allows invalid data.
   💡 Recommendation: Add proper OpenAPI schema to CRD versions with preserveUnknownFields: false

[2] ⚠️ CRD without conversion webhook
   Resource: CustomResourceDefinition/example.com
   Description: CRD example.com has multiple versions but no conversion webhook configured.
   💡 Recommendation: Consider adding a conversion webhook for version migrations
```

### JSON 格式

```json
[
  {
    "category": "CRD",
    "severity": "error",
    "resource": "CustomResourceDefinition",
    "namespace": "",
    "name": "example.com",
    "issue": "CRD without schema (Kubernetes 1.17-)",
    "description": "CRD example.com has empty or missing schema.",
    "recommendation": "Add proper OpenAPI schema to CRD versions"
  }
]
```

### YAML 格式

```yaml
results:
- category: CRD
  severity: error
  resource: CustomResourceDefinition
  namespace: ""
  name: example.com
  issue: CRD without schema (Kubernetes 1.17-)
  description: CRD example.com has empty or missing schema.
  recommendation: Add proper OpenAPI schema to CRD versions
```

## 严重程度说明

- **Error (❌)**: 严重问题，可能导致安全漏洞、数据丢失或系统不稳定
- **Warning (⚠️)**: 潜在问题，可能导致性能问题或不符合最佳实践
- **Info (ℹ️)**: 信息性提示，建议改进但不影响功能

## 架构设计

扫描器采用模块化设计，每个扫描器负责特定类型的资源：

```
cmd/scanner/
├── scanner.go          # 主扫描器，协调所有子扫描器
├── crd_scanner.go      # CRD 扫描器
├── webhook_scanner.go  # Webhook 扫描器
├── controller_scanner.go # Controller 扫描器
└── ownerref_scanner.go # Owner Reference 扫描器
```

## 扩展性

要添加新的扫描器：

1. 在 `cmd/scanner/` 目录下创建新的扫描器文件（如 `new_scanner.go`）
2. 实现扫描器结构体和 `Scan()` 方法
3. 在 `scanner.go` 的 `NewClusterScanner()` 中初始化新扫描器
4. 在 `ClusterScanner.Scan()` 中调用新扫描器的 `Scan()` 方法

## 注意事项

1. **启发式检测**: 某些检测（如 Reconcile 循环）使用启发式方法，可能产生误报
2. **代码分析限制**: 无法直接分析控制器代码，只能通过资源配置推断
3. **权限要求**: 需要足够的 Kubernetes 权限来列出和读取各种资源
4. **性能考虑**: 扫描大量资源可能需要一些时间

## 相关文档

- [Cloud Native Development Best Practices](../docs/Cloud-Native-Development-Best-Practices.zh.md)
- [Kubernetes CRD Best Practices](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
- [Webhook Configuration](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)

