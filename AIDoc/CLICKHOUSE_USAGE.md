# ClickHouse 使用指南

## 配置说明

在 `manifest/config/config.yaml` 中配置 ClickHouse 连接：

```yaml
clickhouse:
  host: "172.19.0.35" # ClickHouse服务地址
  port: 9000 # ClickHouse原生协议端口（9000）
  database: "default" # 数据库名称
  username: "default" # 用户名
  password: "123456" # 密码
```

## 使用方法

### 1. 获取 ClickHouse 管理器

```go
import "jh_game_service/internal/registry"

chManager := registry.GetClickHouseManager()
```

### 2. 创建表

```go
query := `
    CREATE TABLE IF NOT EXISTS game_logs (
        id UInt64,
        user_id UInt32,
        game_id UInt32,
        action String,
        amount Decimal(18, 2),
        created_at DateTime
    ) ENGINE = MergeTree()
    ORDER BY (created_at, user_id)
    PARTITION BY toYYYYMM(created_at)
`

err := chManager.Exec(ctx, query)
```

### 3. 插入单条数据

```go
query := `
    INSERT INTO game_logs (id, user_id, game_id, action, amount, created_at)
    VALUES (?, ?, ?, ?, ?, ?)
`

err := chManager.Exec(ctx, query, 1, 100, 26, "bet", 10.50, time.Now())
```

### 4. 批量插入

```go
conn := chManager.GetConn()
batch, err := conn.PrepareBatch(ctx, "INSERT INTO game_logs")
if err != nil {
    return err
}

for i := 0; i < 100; i++ {
    err := batch.Append(
        uint64(i),
        uint32(100+i),
        uint32(26),
        "bet",
        float64(i)*10.5,
        time.Now(),
    )
    if err != nil {
        return err
    }
}

err = batch.Send()
```

### 5. 查询数据

```go
query := `
    SELECT id, user_id, game_id, action, amount, created_at
    FROM game_logs
    WHERE user_id = ?
    ORDER BY created_at DESC
    LIMIT 10
`

rows, err := chManager.Query(ctx, query, 100)
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    var (
        id        uint64
        userId    uint32
        gameId    uint32
        action    string
        amount    float64
        createdAt time.Time
    )

    if err := rows.Scan(&id, &userId, &gameId, &action, &amount, &createdAt); err != nil {
        return err
    }

    // 处理数据
    fmt.Printf("ID: %d, UserID: %d, Amount: %.2f\n", id, userId, amount)
}
```

### 6. 查询单行

```go
query := `
    SELECT COUNT(*) as total
    FROM game_logs
    WHERE user_id = ?
`

var total uint64
row := chManager.QueryRow(ctx, query, 100)
if err := row.Scan(&total); err != nil {
    return err
}

fmt.Printf("Total: %d\n", total)
```

### 7. 异步插入

```go
query := `
    INSERT INTO game_logs (id, user_id, game_id, action, amount, created_at)
    VALUES (?, ?, ?, ?, ?, ?)
`

// wait=false 表示不等待插入完成，立即返回
err := chManager.AsyncInsert(ctx, query, false, 999, 100, 26, "bet", 99.99, time.Now())
```

## 常用场景

### 游戏日志记录

```go
func LogGameAction(ctx context.Context, userId, gameId uint32, action string, amount float64) error {
    chManager := registry.GetClickHouseManager()

    query := `
        INSERT INTO game_logs (id, user_id, game_id, action, amount, created_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `

    id := generateID() // 生成唯一ID
    return chManager.Exec(ctx, query, id, userId, gameId, action, amount, time.Now())
}
```

### 统计查询

```go
func GetUserGameStats(ctx context.Context, userId uint32, startDate, endDate time.Time) (map[string]interface{}, error) {
    chManager := registry.GetClickHouseManager()

    query := `
        SELECT
            COUNT(*) as total_bets,
            SUM(amount) as total_amount,
            AVG(amount) as avg_amount
        FROM game_logs
        WHERE user_id = ?
          AND created_at >= ?
          AND created_at <= ?
          AND action = 'bet'
    `

    var totalBets uint64
    var totalAmount, avgAmount float64

    row := chManager.QueryRow(ctx, query, userId, startDate, endDate)
    if err := row.Scan(&totalBets, &totalAmount, &avgAmount); err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "total_bets":   totalBets,
        "total_amount": totalAmount,
        "avg_amount":   avgAmount,
    }, nil
}
```

## 注意事项

1. **端口说明**
   - 9000: ClickHouse 原生协议端口（Go 驱动使用）
   - 8123: HTTP 接口端口

2. **数据类型映射**
   - Go `uint64` → ClickHouse `UInt64`
   - Go `uint32` → ClickHouse `UInt32`
   - Go `string` → ClickHouse `String`
   - Go `float64` → ClickHouse `Decimal(18, 2)` 或 `Float64`
   - Go `time.Time` → ClickHouse `DateTime`

3. **性能优化**
   - 使用批量插入提高性能
   - 合理设置分区键（PARTITION BY）
   - 合理设置排序键（ORDER BY）
   - 使用异步插入处理高并发写入

4. **错误处理**
   - 连接失败不会中断服务启动
   - 使用前检查连接是否可用
   - 记录错误日志便于排查

## 参考资料

- [ClickHouse 官方文档](https://clickhouse.com/docs)
- [clickhouse-go 驱动文档](https://github.com/ClickHouse/clickhouse-go)
