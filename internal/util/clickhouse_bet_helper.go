package util

import (
	"context"
	"fmt"
	"jh_game_service/internal/model/entity"
	"jh_game_service/internal/registry"

	"github.com/gogf/gf/v2/frame/g"
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

	query := `
		INSERT INTO game_analytics.bet_records (
			site_id, site_code,
			user_id, username,
			platform, game_id, game_name, game_type,
			order_no, round_id,
			bet_amount, win_amount, valid_amount, net_amount,
			bet_time, settle_time,
			status
		) VALUES (
			?, ?,
			?, ?,
			?, ?, ?, ?,
			?, ?,
			?, ?, ?, ?,
			?, ?,
			?
		)
	`

	err := conn.Exec(ctx, query,
		record.SiteID, record.SiteCode,
		record.UserID, record.Username,
		record.Platform, record.GameID, record.GameName, record.GameType,
		record.OrderNo, record.RoundID,
		record.BetAmount, record.WinAmount, record.ValidAmount, record.NetAmount,
		record.BetTime, record.SettleTime,
		record.Status,
	)

	if err != nil {
		g.Log().Errorf(ctx, "插入ClickHouse失败: %v", err)
		return err
	}

	return nil
}

// BatchInsertBetRecordsToCH 批量插入注单到 ClickHouse（推荐使用）
func BatchInsertBetRecordsToCH(ctx context.Context, records []*entity.ClickHouseBetRecord) error {
	if len(records) == 0 {
		return nil
	}

	chManager := registry.GetClickHouseManager()
	conn := chManager.GetConn()
	if conn == nil {
		return fmt.Errorf("ClickHouse连接未初始化")
	}

	batch, err := conn.PrepareBatch(ctx, `
		INSERT INTO game_analytics.bet_records (
			site_id, site_code,
			user_id, username,
			platform, game_id, game_name, game_type,
			order_no, round_id,
			bet_amount, win_amount, valid_amount, net_amount,
			bet_time, settle_time,
			status
		)
	`)
	if err != nil {
		return fmt.Errorf("准备批量插入失败: %v", err)
	}

	for _, record := range records {
		// 验证数据
		if err := record.Validate(); err != nil {
			g.Log().Warningf(ctx, "跳过无效记录: %v", err)
			continue
		}

		// 自动计算净输赢
		record.CalculateNetAmount()

		err = batch.Append(
			record.SiteID, record.SiteCode,
			record.UserID, record.Username,
			record.Platform, record.GameID, record.GameName, record.GameType,
			record.OrderNo, record.RoundID,
			record.BetAmount, record.WinAmount, record.ValidAmount, record.NetAmount,
			record.BetTime, record.SettleTime,
			record.Status,
		)
		if err != nil {
			return fmt.Errorf("添加记录到批次失败: %v", err)
		}
	}

	if err := batch.Send(); err != nil {
		g.Log().Errorf(ctx, "批量插入ClickHouse失败: %v", err)
		return err
	}

	g.Log().Infof(ctx, "批量插入ClickHouse成功: %d条记录", len(records))
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
