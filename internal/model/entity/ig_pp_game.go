// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// IgPpGame is the golang structure for table ig_pp_game.
type IgPpGame struct {
	Id               uint        `json:"id"               orm:"id"                description:"主键ID"`
	SiteCode         string      `json:"siteCode"         orm:"site_code"         description:"站点编码"`
	ChannelCode      string      `json:"channelCode"      orm:"channel_code"      description:"渠道编码"`
	Username         string      `json:"username"         orm:"username"          description:"用户名"`
	PlatformUsername string      `json:"platformUsername" orm:"platform_username" description:"平台用户名"`
	BetStatus        int         `json:"betStatus"        orm:"bet_status"        description:"投注状态：0=未结算，1=已结算"`
	GameCategoryCode string      `json:"gameCategoryCode" orm:"gameCategoryCode"  description:"游戏类别代码"`
	VendorCode       string      `json:"vendorCode"       orm:"vendorCode"        description:"供应商代码"`
	GameNo           string      `json:"gameNo"           orm:"game_no"           description:"游戏编号/注单号"`
	InningNo         string      `json:"inningNo"         orm:"inning_no"         description:"局号/回合号"`
	BetAmount        float64     `json:"betAmount"        orm:"bet_amount"        description:"投注金额"`
	ValidAmount      float64     `json:"validAmount"      orm:"valid_amount"      description:"有效投注额"`
	Win              float64     `json:"win"              orm:"win"               description:"输赢金额（正数=赢，负数=输）"`
	BetTime          *gtime.Time `json:"betTime"          orm:"bet_time"          description:"投注时间"`
	WinAmount        float64     `json:"winAmount"        orm:"winAmount"         description:"派彩金额"`
	GameCode         string      `json:"gameCode"         orm:"gameCode"          description:"厅内游戏ID"`
	VendorSettleTime int64       `json:"vendorSettleTime" orm:"vendorSettleTime"  description:"供应商结算时间（毫秒时间戳）"`
	Status           uint        `json:"status"           orm:"status"            description:"注单状态： 1: 赢 2: 输 3: 平局"`
	Currency         string      `json:"currency"         orm:"currency"          description:"币种"`
	AgentId          string      `json:"agentId"          orm:"agent_id"          description:"代理appId"`
	Platform         uint        `json:"platform"         orm:"platform"          description:"IG_PP厅的ID"`
	CurrencyCode     string      `json:"currencyCode"     orm:"currencyCode"      description:"币种"`
	CreatedAt        *gtime.Time `json:"createdAt"        orm:"created_at"        description:"创建时间"`
	UpdatedAt        *gtime.Time `json:"updatedAt"        orm:"updated_at"        description:"更新时间"`
}
