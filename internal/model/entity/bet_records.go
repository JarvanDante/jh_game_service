// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BetRecords is the golang structure for table bet_records.
type BetRecords struct {
	Id           uint64      `json:"id"           orm:"id"             description:"主键ID"`
	SiteId       int         `json:"siteId"       orm:"site_id"        description:"站点ID"`
	SiteCode     string      `json:"siteCode"     orm:"site_code"      description:"站点代码"`
	AgentId      int         `json:"agentId"      orm:"agent_id"       description:"代理id"`
	UserId       uint64      `json:"userId"       orm:"user_id"        description:"用户ID"`
	Username     string      `json:"username"     orm:"username"       description:"用户名"`
	Platform     string      `json:"platform"     orm:"platform"       description:"游戏平台"`
	Currency     string      `json:"currency"     orm:"currency"       description:"币种"`
	GameId       int         `json:"gameId"       orm:"game_id"        description:"游戏ID"`
	GameName     string      `json:"gameName"     orm:"game_name"      description:""`
	GameType     string      `json:"gameType"     orm:"game_type"      description:"游戏类型"`
	OrderNo      string      `json:"orderNo"      orm:"order_no"       description:"订单号"`
	RoundId      string      `json:"roundId"      orm:"round_id"       description:"局号/回合号"`
	BetAmount    float64     `json:"betAmount"    orm:"bet_amount"     description:"投注金额"`
	WinAmount    float64     `json:"winAmount"    orm:"win_amount"     description:"赢取金额"`
	ValidAmount  float64     `json:"validAmount"  orm:"valid_amount"   description:"有效投注金额"`
	NetAmount    float64     `json:"netAmount"    orm:"net_amount"     description:"净输赢"`
	BetTime      *gtime.Time `json:"betTime"      orm:"bet_time"       description:"投注时间"`
	SettleTime   *gtime.Time `json:"settleTime"   orm:"settle_time"    description:"结算时间"`
	Status       int         `json:"status"       orm:"status"         description:"状态:0未结算,1已结算,2取消,3无效"`
	RawData      string      `json:"rawData"      orm:"raw_data"       description:"原始数据JSON"`
	ChSyncStatus int         `json:"chSyncStatus" orm:"ch_sync_status" description:"0未同步,1成功,2失败"`
	TaskId       int64       `json:"taskId"       orm:"task_id"        description:"同步任务ID"`
	ChSyncAt     *gtime.Time `json:"chSyncAt"     orm:"ch_sync_at"     description:"同步时间"`
	ChSyncError  string      `json:"chSyncError"  orm:"ch_sync_error"  description:"同步错误"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"     description:"创建时间"`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"     description:"更新时间"`
}
