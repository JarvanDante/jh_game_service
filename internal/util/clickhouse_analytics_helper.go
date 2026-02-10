package util

import (
	"context"
	"fmt"
	"jh_game_service/internal/model/entity"
	"jh_game_service/internal/registry"

	"github.com/shopspring/decimal"
)

// BatchInsertUsersToCH 批量插入用户到 ClickHouse
func BatchInsertUsersToCH(ctx context.Context, records []*entity.ClickHouseUser) error {
	if len(records) == 0 {
		return nil
	}

	conn := registry.GetClickHouseManager().GetConn()
	if conn == nil {
		return fmt.Errorf("ClickHouse连接未初始化")
	}

	batch, err := conn.PrepareBatch(ctx, `
		INSERT INTO game_analytics.user (
			event_id, site_id, user_id, inviter_id, username,
			status, register_ip, register_time, created_at, updated_at,
			pay_times
		)
	`)
	if err != nil {
		return fmt.Errorf("准备批量插入失败: %v", err)
	}

	for _, record := range records {
		if err := batch.Append(
			record.EventID,
			record.SiteID,
			record.UserID,
			record.InviterID,
			record.Username,
			record.Status,
			record.RegisterIP,
			record.RegisterTime,
			record.CreatedAt,
			record.UpdatedAt,
			record.PayTimes,
		); err != nil {
			return fmt.Errorf("添加记录到批次失败: %v", err)
		}
	}

	return batch.Send()
}

// BatchInsertBalanceLogsToCH 批量插入 balance_log 到 ClickHouse
func BatchInsertBalanceLogsToCH(ctx context.Context, records []*entity.ClickHouseBalanceLog) error {
	if len(records) == 0 {
		return nil
	}

	conn := registry.GetClickHouseManager().GetConn()
	if conn == nil {
		return fmt.Errorf("ClickHouse连接未初始化")
	}

	batch, err := conn.PrepareBatch(ctx, `
		INSERT INTO game_analytics.balance_log (
			event_id, site_id, id, user_id, trade_type,
			gateway, bank_type, game_id, trade_no,
			money, actual_money, fee, status, admin_id, remark,
			created_at, updated_at
		)
	`)
	if err != nil {
		return fmt.Errorf("准备批量插入失败: %v", err)
	}

	for _, record := range records {
		money := decimal.NewFromFloat(record.Money)
		actualMoney := decimal.NewFromFloat(record.ActualMoney)
		fee := decimal.NewFromFloat(record.Fee)
		if err := batch.Append(
			record.EventID,
			record.SiteID,
			record.ID,
			record.UserID,
			record.TradeType,
			record.Gateway,
			record.BankType,
			record.GameID,
			record.TradeNo,
			money,
			actualMoney,
			fee,
			record.Status,
			record.AdminID,
			record.Remark,
			record.CreatedAt,
			record.UpdatedAt,
		); err != nil {
			return fmt.Errorf("添加记录到批次失败: %v", err)
		}
	}

	return batch.Send()
}

// BatchInsertBalanceLogPaymentsToCH 批量插入 balance_log_payment 到 ClickHouse
func BatchInsertBalanceLogPaymentsToCH(ctx context.Context, records []*entity.ClickHouseBalanceLogPayment) error {
	if len(records) == 0 {
		return nil
	}

	conn := registry.GetClickHouseManager().GetConn()
	if conn == nil {
		return fmt.Errorf("ClickHouse连接未初始化")
	}

	batch, err := conn.PrepareBatch(ctx, `
		INSERT INTO game_analytics.balance_log_payment (
			event_id, site_id, id, user_id, username, activity_recharge_id,
			gateway, payment_id, payment_account_id, bank_value, trade_no,
			money, actual_money, fee, status, admin_id, remark,
			created_at, updated_at
		)
	`)
	if err != nil {
		return fmt.Errorf("准备批量插入失败: %v", err)
	}

	for _, record := range records {
		money := decimal.NewFromFloat(record.Money)
		actualMoney := decimal.NewFromFloat(record.ActualMoney)
		fee := decimal.NewFromFloat(record.Fee)
		if err := batch.Append(
			record.EventID,
			record.SiteID,
			record.ID,
			record.UserID,
			record.Username,
			record.ActivityRechargeID,
			record.Gateway,
			record.PaymentID,
			record.PaymentAccountID,
			record.BankValue,
			record.TradeNo,
			money,
			actualMoney,
			fee,
			record.Status,
			record.AdminID,
			record.Remark,
			record.CreatedAt,
			record.UpdatedAt,
		); err != nil {
			return fmt.Errorf("添加记录到批次失败: %v", err)
		}
	}

	return batch.Send()
}
