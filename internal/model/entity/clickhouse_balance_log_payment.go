package entity

import "time"

// ClickHouseBalanceLogPayment ClickHouse balance_log_payment 记录
type ClickHouseBalanceLogPayment struct {
	EventID            uint64    `ch:"event_id" json:"event_id"`
	SiteID             uint32    `ch:"site_id" json:"site_id"`
	ID                 uint64    `ch:"id" json:"id"`
	UserID             uint64    `ch:"user_id" json:"user_id"`
	Username           string    `ch:"username" json:"username"`
	ActivityRechargeID uint32    `ch:"activity_recharge_id" json:"activity_recharge_id"`
	Gateway            uint8     `ch:"gateway" json:"gateway"`
	PaymentID          uint32    `ch:"payment_id" json:"payment_id"`
	PaymentAccountID   uint32    `ch:"payment_account_id" json:"payment_account_id"`
	BankValue          string    `ch:"bank_value" json:"bank_value"`
	TradeNo            string    `ch:"trade_no" json:"trade_no"`
	Money              float64   `ch:"money" json:"money"`
	ActualMoney        float64   `ch:"actual_money" json:"actual_money"`
	Fee                float64   `ch:"fee" json:"fee"`
	Status             uint8     `ch:"status" json:"status"`
	AdminID            uint32    `ch:"admin_id" json:"admin_id"`
	Remark             string    `ch:"remark" json:"remark"`
	CreatedAt          time.Time `ch:"created_at" json:"created_at"`
	UpdatedAt          time.Time `ch:"updated_at" json:"updated_at"`
}
