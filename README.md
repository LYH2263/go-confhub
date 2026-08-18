# go-confhub

Go 实现的多租户配置中心库与配套服务：命名空间 → 配置键 → 不可变版本。支持灰度（百分比 / 白名单 / 标签）、客户端 Watch、HMAC/Ed25519 内容签名、审计、配额与快照。

## 运行

```bash
go test ./... -count=1
go run ./cmd/confhubd -addr :8080 -web web
```

浏览器打开 `http://localhost:8080/` 使用管理页（命名空间列表、发布、灰度百分比、最近审计）。

## 库面

`CreateNS` / `DeleteNS`、`Put`、`Get`（按灰度选版本）、`Rollback`、`Watch`、`ExportSnapshot` / `ImportSnapshot`。
