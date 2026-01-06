# GitHub Actions Workflows

本项目包含两个 GitHub Actions workflow：

## 1. CI Workflow (`.github/workflows/ci.yml`)

**触发条件：**
- 推送到 `main` 或 `master` 分支
- 创建 Pull Request 到 `main` 或 `master` 分支

**功能：**
- 运行所有测试
- 在多个平台上测试构建（Ubuntu, macOS, Windows）
- 生成代码覆盖率报告

## 2. Release Workflow (`.github/workflows/release.yml`)

**触发条件：**
- 推送以 `v` 开头的 tag（例如：`v1.0.0`）
- 手动触发（workflow_dispatch）

**功能：**
- 在多个平台和架构上构建：
  - Linux: amd64, arm64
  - macOS: amd64, arm64
  - Windows: amd64
- 构建两个二进制文件：
  - `kube-killer` - 主程序
  - `kubectl-kill` - kubectl 插件
- 自动创建 GitHub Release
- 上传所有平台的构建产物
- 生成 SHA256 校验和文件

## 使用方法

### 创建新版本发布

1. 更新版本号（如果需要）
2. 提交更改
3. 创建并推送 tag：
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
4. GitHub Actions 会自动：
   - 构建所有平台的二进制文件
   - 创建 GitHub Release
   - 上传所有构建产物

### 手动触发构建

1. 前往 GitHub 仓库的 Actions 页面
2. 选择 "Release" workflow
3. 点击 "Run workflow"
4. 选择分支并运行

## 构建产物

每个 Release 包含以下文件：
- `kube-killer-<version>-linux-amd64.tar.gz`
- `kube-killer-<version>-linux-arm64.tar.gz`
- `kube-killer-<version>-darwin-amd64.tar.gz`
- `kube-killer-<version>-darwin-arm64.tar.gz`
- `kube-killer-<version>-windows-amd64.zip`
- `checksums.txt` - 所有文件的 SHA256 校验和

每个压缩包包含：
- `kube-killer` (或 `kube-killer.exe` on Windows)
- `kubectl-kill` (或 `kubectl-kill.exe` on Windows)

