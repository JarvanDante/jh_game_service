package util

import (
	"context"
	"fmt"
	"time"

	"jh_game_service/internal/registry"
)

// ClickHouseExample ClickHouse使用示例

// ExampleCreateTable 创建表示例
func ExampleCreateTable(ctx context.Context) error {
	chManager := registry.GetClickHouseManager()

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

	return chManager.Exec(ctx, query)
}

// ExampleInsertData 插入数据示例
func ExampleInsertData(ctx context.Context) error {
	chManager := registry.GetClickHouseManager()

	query := `
		INSERT INTO game_logs (id, user_id, game_id, action, amount, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	return chManager.Exec(ctx, query, 1, 100, 26, "bet", 10.50, time.Now())
}

// ExampleBatchInsert 批量插入示例
func ExampleBatchInsert(ctx context.Context) error {
	chManager := registry.GetClickHouseManager()
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

	return batch.Send()
}

// ExampleQueryData 查询数据示例
func ExampleQueryData(ctx context.Context) error {
	chManager := registry.GetClickHouseManager()

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

		fmt.Printf("ID: %d, UserID: %d, GameID: %d, Action: %s, Amount: %.2f, Time: %s\n",
			id, userId, gameId, action, amount, createdAt.Format("2006-01-02 15:04:05"))
	}

	return rows.Err()
}

// ExampleQueryRow 查询单行示例
func ExampleQueryRow(ctx context.Context) error {
	chManager := registry.GetClickHouseManager()

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

	fmt.Printf("Total records: %d\n", total)
	return nil
}

// ExampleAsyncInsert 异步插入示例
func ExampleAsyncInsert(ctx context.Context) error {
	chManager := registry.GetClickHouseManager()

	query := `
		INSERT INTO game_logs (id, user_id, game_id, action, amount, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	// wait=false 表示不等待插入完成，立即返回
	return chManager.AsyncInsert(ctx, query, false, 999, 100, 26, "bet", 99.99, time.Now())
}
