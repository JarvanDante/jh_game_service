// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BetRecords is the golang structure of table bet_records for DAO operations like Where/Data.
type BetRecords struct {
	g.Meta       `orm:"table:bet_records, do:true"`
	Id           any         // 主键ID
	SiteId       any         // 站点ID
	SiteCode     any         // 站点代码
	UserId       any         // 用户ID
	Username     any         // 用户名
	Platform     any         // 游戏平台
	Currency     any         // 币种
	GameId       any         // 游戏ID
	GameName     any         // 游戏名称
	GameType     any         // 游戏类型
	OrderNo      any         // 订单号
	RoundId      any         // 局号/回合号
	BetAmount    any         // 投注金额
	WinAmount    any         // 赢取金额
	ValidAmount  any         // 有效投注金额
	NetAmount    any         // 净输赢
	BetTime      *gtime.Time // 投注时间
	SettleTime   *gtime.Time // 结算时间
	Status       any         // 状态:0未结算,1已结算,2取消,3无效
	RawData      any         // 原始数据JSON
	ChSyncStatus any         // 0未同步,1成功,2失败
	ChSyncAt     *gtime.Time // 同步时间
	ChSyncError  any         // 同步错误
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
}
