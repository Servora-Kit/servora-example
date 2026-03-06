<!-- Parent: ../AGENTS.md -->
# Redis 客户端封装 (pkg/redis)

**最后更新时间**: 2026-03-06

## 模块目的

封装 `github.com/redis/go-redis/v9`，统一配置转换、连通性探测、日志与基础操作。

## 当前实现事实

- 默认超时：`Dial=5s`、`Read=3s`、`Write=3s`
- `NewConfigFromProto` 从 `conf.Data_Redis` 构造本地配置
- `NewClient` 会先 `Ping` 校验连接，并返回 `cleanup func()`
- 初始化日志统一带 `module=redis/pkg`

## 暴露能力

- `Ping`
- `Set` / `Get` / `Del` / `Has` / `Keys`
- `SAdd` / `SMembers`
- `Expire`

## 使用示例

```go
cfg := &redis.Config{Addr: "localhost:6379", DB: 0}
client, cleanup, err := redis.NewClient(cfg, l)
defer cleanup()

_ = client.Set(context.Background(), "key", "value", time.Hour)
```

## 测试

```bash
go test ./pkg/redis/...
```

需要本地 Redis；不可用时应在测试里 `t.Skipf(...)`。
