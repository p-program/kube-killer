# kube-killer 二进制发布说明

本文档记录一次本地交叉编译产物：平台包路径、校验和与构建元数据。制品默认生成在仓库根目录下的 `dist/`。

## 本次构建

| 项 | 值 |
| --- | --- |
| 版本（`git describe`） | `v1.0.1-16-g0f7a6b9` |
| Git commit | `0f7a6b9` |
| 构建时间（UTC） | `2026-04-17T04:07:19Z` |
| Go 工具链 | `go1.25.6 darwin/arm64` |
| 编译选项 | `CGO_ENABLED=0`，`go build -trimpath`，`-ldflags "-s -w -X github.com/p-program/kube-killer/cmd.VERSION=…"`（版本字符串与上表一致，不含前缀 `v`） |

## 制品与 SHA256

算法：**SHA-256**。下列文件均位于 `dist/`。

| 平台 | 文件名 | SHA256 |
| --- | --- | --- |
| Linux x86_64 | `kube-killer-v1.0.1-16-g0f7a6b9-linux-amd64.tar.gz` | `3bc22bdca960a7e40d2095cd54991ef609352af9a2e1fee02c3a6e383d236dfd` |
| Linux arm64 | `kube-killer-v1.0.1-16-g0f7a6b9-linux-arm64.tar.gz` | `baef894b5359c57ffbfe92d01654547b358971fa78ed4ea9276e907a9337ed21` |
| macOS x86_64 | `kube-killer-v1.0.1-16-g0f7a6b9-darwin-amd64.tar.gz` | `4b92e66a3b61a5b815d39e267dd39b18addefef0e5c943570cbc4a6e0db0a7eb` |
| macOS arm64 | `kube-killer-v1.0.1-16-g0f7a6b9-darwin-arm64.tar.gz` | `3bfeea427b8e998389a93041067a3f0270551edb4175a5144a1c9247d18807fe` |
| Windows x86_64 | `kube-killer-v1.0.1-16-g0f7a6b9-windows-amd64.zip` | `9b908c120f6a7f4c325557c6e48748d7107ab86c900cae89dd50f2f599276419` |

- **tar.gz**：内含可执行文件 `kube-killer`（类 Unix）。
- **zip**：内含 `kube-killer.exe`。

## 校验示例

在 `dist/` 中：

```bash
shasum -a 256 kube-killer-v1.0.1-16-g0f7a6b9-linux-amd64.tar.gz
```

预期输出首列为上表对应 SHA256，后接文件名。

一次性校验全部包（`dist/SHA256SUMS` 与本次构建一致时）：

```bash
cd dist && shasum -c SHA256SUMS
```

## 重新生成制品

在仓库根目录执行（需已安装与 `go.mod` 匹配的 Go）：

```bash
VERSION=$(git describe --tags --always --dirty)
VERSION_X=${VERSION#v}
DIST_ROOT=dist
rm -rf "$DIST_ROOT" && mkdir -p "$DIST_ROOT"
LDFLAGS="-s -w -X github.com/p-program/kube-killer/cmd.VERSION=${VERSION_X}"

build_one() {
  local goos=$1 goarch=$2 suffix=$3
  local name="kube-killer-${VERSION}-$suffix"
  local outdir="$DIST_ROOT/build-$suffix"
  mkdir -p "$outdir"
  if [ "$goos" = "windows" ]; then
    GOOS=$goos GOARCH=$goarch CGO_ENABLED=0 go build -trimpath -ldflags "$LDFLAGS" -o "$outdir/kube-killer.exe" .
    (cd "$outdir" && zip -q "../${name}.zip" kube-killer.exe)
  else
    GOOS=$goos GOARCH=$goarch CGO_ENABLED=0 go build -trimpath -ldflags "$LDFLAGS" -o "$outdir/kube-killer" .
    (cd "$outdir" && tar -czf "../${name}.tar.gz" kube-killer)
  fi
  rm -rf "$outdir"
}

build_one linux amd64 linux-amd64
build_one linux arm64 linux-arm64
build_one darwin amd64 darwin-amd64
build_one darwin arm64 darwin-arm64
build_one windows amd64 windows-amd64

(cd "$DIST_ROOT" && shasum -a 256 kube-killer-*.tar.gz kube-killer-*.zip 2>/dev/null | sort | tee SHA256SUMS)
```

重新构建后请同步更新本文件中的版本、时间与哈希列。
