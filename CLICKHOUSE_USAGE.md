# ClickHouse 注单统计使用指南

## 📋 概述

ClickHouse 用于存储和统计**游戏注单数据**，提供高性能的实时分析能力。

**重要说明**：

- ✅ ClickHouse：存储注单数据，统计投注相关指标
- ✅ MySQL：存储用户数据、财务数据，统计用户和充值提现指标

## 🗂️ 数据职责划分

### ClickHouse 负责统计（注单相关）

- 有效投注笔数（注单数）
- 有效投注金额 (`valid_amount`)
- 净输赢 (`net_amount` = `win_amount` - `bet_amount`)
- 派彩金额 (`win_amount`)
- 投注金额 (`bet_amount`)
- 返水率 = (净输赢 / 有效投注金额) × 100

### MySQL 负责统计（用户和财务相关）

- 新注册会员数
- 活跃会员数
- 首存用户数
- 充值总额
- 提现总额

## 📊 核心字段说明

### 金额字段（与 MySQL 完全一致）

| 字段名         | 类型          | 说明                | 用途                                        |
| -------------- | ------------- | ------------------- | ------------------------------------------- |
| `bet_amount`   | Decimal(18,6) | 投注金额            | 用户下注的总金额                            |
| `win_amount`   | Decimal(18,6) | 赢取金额            | 用户赢得的金额                              |
| `valid_amount` | Decimal(18,6) | **有效投注金额** ⭐ | 用于统计和返水计算                          |
| `net_amount`   | Decimal(18,6) | **净输赢** ⭐       | `win_amount - bet_amount`，正数=赢，负数=输 |

### 状态字段

| 值  | 说明   | 是否统计    |
| --- | ------ | ----------- |
| 0   | 未结算 | ❌ 不统计   |
| 1   | 已结算 | ✅ **统计** |
| 2   | 取消   | ❌ 不统计   |
| 3   | 无效   | ❌ 不统计   |

**统计规则**：只统计 `status = 1` 的记录

## 🏗️ 表结构

### 主表：bet_records

```sql
CREATE TABLE game_analytics.bet_records (
    -- 站点信息
    site_id     UInt32,
    site_code   String,

    -- 用户信息
    user_id     UInt64,
    username    String,

    -- 游戏信息
    platform    LowCardinality(String),
    game_id     UInt32,
    game_name   String,
    game_type   LowCardinality(String),

    -- 订单信息
    order_no    String,
    round_id    String,

    -- 金额字段（核心）
    bet_amount    Decimal(18,6),
    win_amount    Decimal(18,6),
    valid_amount  Decimal(18,6),  -- ⭐ 有效投注
    net_amount    Decimal(18,6),  -- ⭐ 净输赢

    -- 时间字段
    bet_time    DateTime,
    settle_time Nullable(DateTime),

    -- 状态
    status      UInt8,

    -- 自动生成字段（MATERIALIZED）
    bet_date Date MATERIALIZED toDate(bet_time),
    bet_hour UInt8 MATERIALIZED toHour(bet_time)
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(bet_time)
ORDER BY (site_id, bet_time, user_id)
TTL bet_time + INTERVAL 2 YEAR;
```

**特点**：

- 按月分区 (`toYYYYMM`)
- 自动生成 `bet_date` 和 `bet_hour`（无需手动插入）
- 数据保留 2 年（TTL）

## 💻 Go 代码使用

### 1. 插入单条记录

```go
import (
    "jh_game_service/internal/model/entity"
    "jh_game_service/internal/util"
)

record := &entity.ClickHouseBetRecord{
    SiteID:      1,
    SiteCode:    "jh",
    UserID:      12345,
    Username:    "test_user",
    Platform:    "ONE",
    GameID:      100,
    GameName:    "Fortune Tiger",
    GameType:    "slot",
    OrderNo:     "ORDER123456",
    RoundID:     "ROUND789",
    BetAmount:   100.00,
    WinAmount:   150.00,
    ValidAmount: 100.00,  // 有效投注
    NetAmount:   50.00,   // 净输赢 = 150 - 100
    BetTime:     time.Now(),
    Status:      1,       // 已结算
}

err := util.InsertBetRecordToCH(ctx, record)
```

### 2. 批量插入（推荐）

```go
records := []*entity.ClickHouseBetRecord{
    // ... 多条记录
}

err := util.BatchInsertBetRecordsToCH(ctx, records)
```

### 3. 查询统计数据

```go
// 查询指定日期范围的注单统计（仅注单部分）
stats, err := util.QueryBetStatsByDateRange(ctx, siteID, "2026-01-01", "2026-01-31")

// stats 包含：
// - BetCount: 注单数
// - ValidAmount: 有效投注金额
// - NetAmount: 净输赢
// - WinAmount: 派彩金额
// - BetAmount: 投注金额
// - ReturnRate: 返水率

// 注意：NewUsers, ActiveUsers, FirstDepositUsers, DepositAmount, WithdrawAmount
// 需要从 MySQL 单独查询后填充
```

### 4. 按平台统计

```go
results, err := util.QueryBetStatsByPlatform(ctx, siteID, "2026-01-01", "2026-01-31")

// 返回每个平台的：
// - platform: 平台名称
// - bet_count: 注单数
// - valid_amount: 有效投注
// - net_amount: 净输赢
// - unique_users: 参与用户数
```

### 5. 查询投注排行榜

```go
topUsers, err := util.QueryTopUsers(ctx, siteID, "2026-01-26", 10)

// 返回前10名用户的：
// - user_id: 用户ID
// - username: 用户名
// - bet_count: 注单数
// - valid_amount: 有效投注
// - net_amount: 净输赢
```

## 📈 统计查询示例

### 查询每日统计（只统计已结算）

```sql
SELECT
    site_code,
    count() AS bet_count,
    sum(valid_amount) AS valid_amount,
    sum(net_amount) AS net_amount,
    sum(win_amount) AS win_amount,
    CASE
        WHEN sum(valid_amount) > 0
        THEN (sum(net_amount) / sum(valid_amount)) * 100
        ELSE 0
    END AS return_rate
FROM game_analytics.bet_records
WHERE toDate(bet_time) BETWEEN '2026-01-01' AND '2026-01-31'
  AND site_id = 1
  AND status = 1  -- ⭐ 只统计已结算
GROUP BY site_code;
```

### 按平台统计

```sql
SELECT
    platform,
    count() AS bet_count,
    sum(valid_amount) AS valid_amount,
    sum(net_amount) AS net_amount,
    uniqExact(user_id) AS unique_users
FROM game_analytics.bet_records
WHERE toDate(bet_time) = '2026-01-26'
  AND site_id = 1
  AND status = 1
GROUP BY platform
ORDER BY valid_amount DESC;
```

### 按小时统计（实时监控）

```sql
SELECT
    toHour(bet_time) AS hour,
    count() AS bet_count,
    sum(valid_amount) AS valid_amount,
    sum(net_amount) AS net_amount
FROM game_analytics.bet_records
WHERE toDate(bet_time) = today()
  AND site_id = 1
  AND status = 1
GROUP BY hour
ORDER BY hour;
```

## 🔄 数据同步流程

### 从 MySQL 同步到 ClickHouse

```
MySQL bet_records (主表)
    ↓
定时任务/实时同步
    ↓
ClickHouse bet_records (分析表)
```

**同步字段映射**（字段名完全一致）：

| MySQL          | ClickHouse     | 说明                    |
| -------------- | -------------- | ----------------------- |
| `bet_amount`   | `bet_amount`   | 投注金额                |
| `win_amount`   | `win_amount`   | 赢取金额                |
| `valid_amount` | `valid_amount` | 有效投注 ⭐             |
| `net_amount`   | `net_amount`   | 净输赢 ⭐               |
| `status`       | `status`       | 状态（只同步 status=1） |

## ⚠️ 注意事项

1. **字段名一致性**：ClickHouse 和 MySQL 字段名完全一致，便于数据同步
2. **只统计已结算**：查询时必须加 `status = 1` 条件
3. **净输赢计算**：`net_amount = win_amount - bet_amount`（正数=赢，负数=输）
4. **有效投注**：`valid_amount` 是统计的核心字段，用于返水计算
5. **自动字段**：`bet_date` 和 `bet_hour` 由 ClickHouse 自动生成，插入时不需要设置
6. **数据分离**：用户统计和财务统计从 MySQL 查询，不要混在 ClickHouse 中

## 📁 相关文件

- SQL 建表脚本：`jh_game_service/manifest/sql/clickhouse_bet_records.sql`
- Go 实体定义：`jh_game_service/internal/model/entity/clickhouse_bet_record.go`
- Go 工具函数：`jh_game_service/internal/util/clickhouse_bet_helper.go`
- MySQL 表结构：`jh_game_service/manifest/sql/mysql_bet_records.sql`

## 🎯 前端展示数据来源

根据前端统计需求，数据来源如下：

| 统计项           | 数据来源   | 查询表                       |
| ---------------- | ---------- | ---------------------------- |
| 新注册会员       | MySQL      | `user` 表                    |
| 活跃会员         | MySQL      | `user` 表 + 登录日志         |
| 首存用户数       | MySQL      | `balance_log` 表             |
| 充值总额         | MySQL      | `balance_log` 表 (type=充值) |
| 提现总额         | MySQL      | `balance_log` 表 (type=提现) |
| 有效投注（笔数） | ClickHouse | `bet_records` 表             |
| 有效投注金额     | ClickHouse | `bet_records.valid_amount`   |
| 净输赢           | ClickHouse | `bet_records.net_amount`     |
| 派彩金额         | ClickHouse | `bet_records.win_amount`     |
| 返水率           | 计算       | (净输赢 / 有效投注) × 100    |

**完整统计流程**：

1. 从 MySQL 查询用户和财务数据
2. 从 ClickHouse 查询注单数据
3. 在应用层合并两部分数据
4. 返回给前端展示
