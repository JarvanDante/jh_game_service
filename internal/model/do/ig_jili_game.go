// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// IgJiliGame is the golang structure of table ig_jili_game for DAO operations like Where/Data.
type IgJiliGame struct {
	g.Meta           `orm:"table:ig_jili_game, do:true"`
	Id               any         // 主键ID
	SiteCode         any         // 站点编码
	ChannelCode      any         // 渠道编码
	Username         any         // 用户名
	PlatformUsername any         // 平台用户名
	BetStatus        any         // 投注状态：0=未结算，1=已结算
	GameCategoryCode any         // 游戏类别代码
	VendorCode       any         // 供应商代码
	GameNo           any         // 游戏编号/注单号
	InningNo         any         // 局号/回合号
	BetAmount        any         // 投注金额
	ValidAmount      any         // 有效投注额
	Win              any         // 输赢金额（正数=赢，负数=输）
	BetTime          *gtime.Time // 投注时间
	CreatedAt        *gtime.Time // 创建时间
	UpdatedAt        *gtime.Time // 更新时间
	WinAmount        any         // 派彩金额
	GameCode         any         // 厅内游戏ID
	VendorSettleTime any         // 供应商结算时间（毫秒时间戳）
	Status           any         // 注单状态： 1: 赢 2: 输 3: 平局
	Currency         any         // 币种
	AgentId          any         // 代理appId
	Platform         any         // IG_JILI厅的ID
	CurrencyCode     any         // 币种
}
