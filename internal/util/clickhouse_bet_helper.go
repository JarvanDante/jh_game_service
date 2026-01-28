package util

import (
	"context"
	"fmt"
	"jh_game_service/internal/model/entity"
	"jh_game_service/internal/registry"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"
)

// InsertBetRecordToCH 插入注单到 ClickHouse
func InsertBetRecordToCH(ctx context.Context, record *entity.ClickHouseBetRecord) error {
	chManager := registry.GetClickHouseManager()
	conn := chManager.GetConn()
	if conn == nil {
		return fmt.Errorf("ClickHouse连接未初始化")
	}

	// 验证数据
	if err := record.Validate(); err != nil {
		return fmt.Errorf("数据验证失败: %w", err)
	}

	// 自动计算净输赢
	record.CalculateNetAmount()

	// 设置更新时间（用于 ReplacingMergeTree 去重）
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = time.Now()
	}

	// 将 float64 转换为 Decimal 类型
	betAmount := decimal.NewFromFloat(record.BetAmount)
	winAmount := decimal.NewFromFloat(record.WinAmount)
	validAmount := decimal.NewFromFloat(record.ValidAmount)
	netAmount := decimal.NewFromFloat(record.NetAmount)

	query := `
		INSERT INTO game_analytics.bet_records (
			site_id, site_code,
			user_id, username,
			platform, game_id, game_name, game_type,
			order_no, round_id,
			bet_amount, win_amount, valid_amount, net_amount,
			bet_time, settle_time,
			status, updated_at
		) VALUES (
			?, ?,
			?, ?,
			?, ?, ?, ?,
			?, ?,
			?, ?, ?, ?,
			?, ?,
			?, ?
		)
	`

	err := conn.Exec(ctx, query,
		record.SiteID, record.SiteCode,
		record.UserID, record.Username,
		record.Platform, record.GameID, record.GameName, record.GameType,
		record.OrderNo, record.RoundID,
		betAmount, winAmount, validAmount, netAmount,
		record.BetTime, record.SettleTime,
		record.Status, record.UpdatedAt,
	)

	if err != nil {
		g.Log().Errorf(ctx, "插入ClickHouse失败: %v", err)
		return err
	}

	return nil
}

// BatchInsertBetRecordsToCH 批量插入注单到 ClickHouse（推荐使用）
// 实现幂等性：插入前检查 order_no 是否已存在
func BatchInsertBetRecordsToCH(ctx context.Context, records []*entity.ClickHouseBetRecord) error {
	if len(records) == 0 {
		return nil
	}

	chManager := registry.GetClickHouseManager()
	conn := chManager.GetConn()
	if conn == nil {
		return fmt.Errorf("ClickHouse连接未初始化")
	}

	// 1. 收集所有 order_no
	orderNos := make([]string, 0, len(records))
	orderNoMap := make(map[string]*entity.ClickHouseBetRecord)
	for _, record := range records {
		if err := record.Validate(); err != nil {
			g.Log().Warningf(ctx, "跳过无效记录: %v", err)
			continue
		}
		orderNos = append(orderNos, record.OrderNo)
		orderNoMap[record.OrderNo] = record
	}

	if len(orderNos) == 0 {
		return nil
	}

	// 2. 查询已存在的 order_no（使用 IN 查询）
	existingOrderNos := make(map[string]bool)
	if len(orderNos) > 0 {
		// 构建 IN 查询
		query := "SELECT DISTINCT order_no FROM game_analytics.bet_records WHERE order_no IN ("
		params := make([]interface{}, len(orderNos))
		for i, orderNo := range orderNos {
			if i > 0 {
				query += ","
			}
			query += "?"
			params[i] = orderNo
		}
		query += ")"

		rows, err := conn.Query(ctx, query, params...)
		if err != nil {
			g.Log().Warningf(ctx, "查询已存在订单号失败: %v，继续插入", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var orderNo string
				if err := rows.Scan(&orderNo); err == nil {
					existingOrderNos[orderNo] = true
				}
			}
		}
	}

	// 3. 过滤掉已存在的记录
	newRecords := make([]*entity.ClickHouseBetRecord, 0, len(records))
	skippedCount := 0
	for _, record := range records {
		if existingOrderNos[record.OrderNo] {
			skippedCount++
			continue
		}
		newRecords = append(newRecords, record)
	}

	if skippedCount > 0 {
		g.Log().Infof(ctx, "跳过已存在的订单: %d条", skippedCount)
	}

	if len(newRecords) == 0 {
		g.Log().Infof(ctx, "所有记录都已存在，跳过插入")
		return nil
	}

	// 4. 批量插入新记录
	batch, err := conn.PrepareBatch(ctx, `
		INSERT INTO game_analytics.bet_records (
			site_id, site_code,
			user_id, username,
			platform, game_id, game_name, game_type,
			order_no, round_id,
			bet_amount, win_amount, valid_amount, net_amount,
			bet_time, settle_time,
			status, updated_at
		)
	`)
	if err != nil {
		return fmt.Errorf("准备批量插入失败: %v", err)
	}

	for _, record := range newRecords {
		// 自动计算净输赢
		record.CalculateNetAmount()

		// 设置更新时间（用于 ReplacingMergeTree 去重）
		if record.UpdatedAt.IsZero() {
			record.UpdatedAt = time.Now()
		}

		// 将 float64 转换为 Decimal 类型
		betAmount := decimal.NewFromFloat(record.BetAmount)
		winAmount := decimal.NewFromFloat(record.WinAmount)
		validAmount := decimal.NewFromFloat(record.ValidAmount)
		netAmount := decimal.NewFromFloat(record.NetAmount)

		err = batch.Append(
			record.SiteID, record.SiteCode,
			record.UserID, record.Username,
			record.Platform, record.GameID, record.GameName, record.GameType,
			record.OrderNo, record.RoundID,
			betAmount, winAmount, validAmount, netAmount,
			record.BetTime, record.SettleTime,
			record.Status, record.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("添加记录到批次失败: %v", err)
		}
	}

	if err := batch.Send(); err != nil {
		g.Log().Errorf(ctx, "批量插入ClickHouse失败: %v", err)
		return err
	}

	g.Log().Infof(ctx, "批量插入ClickHouse成功: %d条新记录（跳过%d条已存在）", len(newRecords), skippedCount)
	return nil
}

// QueryBetStatsByDateRange 查询指定日期范围的注单统计数据（仅注单部分）
// 注意：用户统计和财务统计需要从 MySQL 单独查询
func QueryBetStatsByDateRange(ctx context.Context, siteID uint32, startDate, endDate string) (*entity.BetStatsSummary, error) {
	chManager := registry.GetClickHouseManager()
	conn := chManager.GetConn()
	if conn == nil {
		return nil, fmt.Errorf("ClickHouse连接未初始化")
	}

	query := `
		SELECT 
			site_code,
			count() AS bet_count,
			sum(bet_amount) AS bet_amount,
			sum(win_amount) AS win_amount,
			sum(valid_amount) AS valid_amount,
			sum(net_amount) AS net_amount
		FROM game_analytics.bet_records
		WHERE site_id = ?
		  AND toDate(bet_time) BETWEEN ? AND ?
		  AND status = 1  -- 只统计已结算
		GROUP BY site_code
	`

	rows, err := conn.Query(ctx, query, siteID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("查询统计数据失败: %v", err)
	}
	defer rows.Close()

	var stats entity.BetStatsSummary
	if rows.Next() {
		err = rows.Scan(
			&stats.SiteCode,
			&stats.BetCount,
			&stats.BetAmount,
			&stats.WinAmount,
			&stats.ValidAmount,
			&stats.NetAmount,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描统计数据失败: %v", err)
		}
	}

	// 计算返水率
	if stats.ValidAmount > 0 {
		stats.ReturnRate = (stats.NetAmount / stats.ValidAmount) * 100
	}

	// 注意：NewUsers, ActiveUsers, FirstDepositUsers, DepositAmount, WithdrawAmount
	// 需要从 MySQL 查询后填充

	return &stats, nil
}

// QueryBetStatsByPlatform 按平台统计注单数据
func QueryBetStatsByPlatform(ctx context.Context, siteID uint32, startDate, endDate string) ([]map[string]interface{}, error) {
	chManager := registry.GetClickHouseManager()
	conn := chManager.GetConn()
	if conn == nil {
		return nil, fmt.Errorf("ClickHouse连接未初始化")
	}

	query := `
		SELECT 
			platform,
			count() AS bet_count,
			sum(valid_amount) AS valid_amount,
			sum(net_amount) AS net_amount,
			uniqExact(user_id) AS unique_users
		FROM game_analytics.bet_records
		WHERE site_id = ?
		  AND toDate(bet_time) BETWEEN ? AND ?
		  AND status = 1
		GROUP BY platform
		ORDER BY valid_amount DESC
	`

	rows, err := conn.Query(ctx, query, siteID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("查询平台统计失败: %v", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var platform string
		var betCount, uniqueUsers uint64
		var validAmount, netAmount float64

		err = rows.Scan(&platform, &betCount, &validAmount, &netAmount, &uniqueUsers)
		if err != nil {
			return nil, fmt.Errorf("扫描数据失败: %v", err)
		}

		results = append(results, map[string]interface{}{
			"platform":     platform,
			"bet_count":    betCount,
			"valid_amount": validAmount,
			"net_amount":   netAmount,
			"unique_users": uniqueUsers,
		})
	}

	return results, nil
}

// QueryTopUsers 查询投注排行榜
func QueryTopUsers(ctx context.Context, siteID uint32, date string, limit int) ([]map[string]interface{}, error) {
	chManager := registry.GetClickHouseManager()
	conn := chManager.GetConn()
	if conn == nil {
		return nil, fmt.Errorf("ClickHouse连接未初始化")
	}

	query := `
		SELECT 
			user_id,
			username,
			count() AS bet_count,
			sum(valid_amount) AS valid_amount,
			sum(net_amount) AS net_amount
		FROM game_analytics.bet_records
		WHERE site_id = ?
		  AND toDate(bet_time) = ?
		  AND status = 1
		GROUP BY user_id, username
		ORDER BY valid_amount DESC
		LIMIT ?
	`

	rows, err := conn.Query(ctx, query, siteID, date, limit)
	if err != nil {
		return nil, fmt.Errorf("查询排行榜失败: %v", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var userID uint64
		var username string
		var betCount uint64
		var validAmount, netAmount float64

		err = rows.Scan(&userID, &username, &betCount, &validAmount, &netAmount)
		if err != nil {
			return nil, fmt.Errorf("扫描数据失败: %v", err)
		}

		results = append(results, map[string]interface{}{
			"user_id":      userID,
			"username":     username,
			"bet_count":    betCount,
			"valid_amount": validAmount,
			"net_amount":   netAmount,
		})
	}

	return results, nil
}
