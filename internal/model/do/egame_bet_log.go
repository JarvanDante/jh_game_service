// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// EgameBetLog is the golang structure of table egame_bet_log for DAO operations like Where/Data.
type EgameBetLog struct {
	g.Meta       `orm:"table:egame_bet_log, do:true"`
	Id           any         // 自增主键
	SiteId       any         // 站点ID
	UserId       any         // 用户ID
	Username     any         // 用户名快照
	PlatformCode any         // 游戏平台代码 egame/jili/pg
	GameCode     any         // 游戏代码
	GameName     any         // 游戏名称
	GameType     any         // 游戏类型
	OrderId      any         // 厂商注单号
	RoundId      any         // 局号
	BetAmount    any         // 投注金额
	WinAmount    any         // 中奖金额
	Currency     any         // 币种
	BetTime      *gtime.Time // 投注时间
	SettleTime   *gtime.Time // 结算时间
	Status       any         // 状态 0=未结算 1=已结算 2=取消 3=回滚
	RawData      any         // 厂商原始注单数据
	Remark       any         // 备注
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
}
