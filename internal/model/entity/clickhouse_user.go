package entity

import "time"

// ClickHouseUser ClickHouse 用户表记录
type ClickHouseUser struct {
	EventID      uint64    `ch:"event_id" json:"event_id"`
	SiteID       uint32    `ch:"site_id" json:"site_id"`
	UserID       uint64    `ch:"user_id" json:"user_id"`
	InviterID    uint64    `ch:"inviter_id" json:"inviter_id"`
	Username     string    `ch:"username" json:"username"`
	Status       uint8     `ch:"status" json:"status"`
	RegisterIP   string    `ch:"register_ip" json:"register_ip"`
	RegisterTime time.Time `ch:"register_time" json:"register_time"`
	CreatedAt    time.Time `ch:"created_at" json:"created_at"`
	UpdatedAt    time.Time `ch:"updated_at" json:"updated_at"`
	PayTimes     uint32    `ch:"pay_times" json:"pay_times"`
}
