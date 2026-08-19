# ConfHub 配置中心

多租户配置中心库与配套服务：命名空间、配置键、不可变版本、灰度、Watch、签名与审计。

- 管理页：`go run ./cmd/confhubd -addr :8080 -web web`，浏览器打开对应地址。

## 环境

- 镜像：`benzhi.Dockerfile` 基于 `golang:1.22`（官方多架构）
- `go.mod` 语言版本：go 1.22
- 容器内使用镜像自带工具链即可

## 标准命令

```bash
go build ./...
go test ./... -count=1
go vet ./...
```

## 构建评测镜像（须双架构）

验证请用 `bash -c`（勿用 `bash -lc`）。

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh go-confhub linux/amd64
docker run --platform linux/amd64 --rm go-confhub:latest bash -c 'go build ./...'

./build_benzhi_docker.sh go-confhub linux/arm64
docker run --platform linux/arm64 --rm go-confhub:latest bash -c 'go build ./...'
```

构建阶段已 `go mod download`；容器内编译不应再出现 `downloading ...`。
