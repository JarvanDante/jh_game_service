// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// GfgGame is the golang structure of table gfg_game for DAO operations like Where/Data.
type GfgGame struct {
	g.Meta                `orm:"table:gfg_game, do:true"`
	Id                    any         // 主键ID
	GameCategoryCode      any         //
	Username              any         //
	CurrencyCode          any         // 币种
	GameCode              any         // 游戏代码
	VendorCode            any         // 提供商code
	AccountId             any         // 用户ID
	Platform              any         // 平台ID
	InningNo              any         //
	MainRoundId           any         // 主局号
	GameNo                any         //
	ExternalTransactionId any         // 由游戏供应商提供的外部交易ID
	GameResult            any         // 游戏结果
	FieldId               any         // 场ID
	FieldName             any         // 场名称
	TableId               any         // 桌ID
	Chair                 any         // 椅子ID
	BetAmount             any         //
	WinAmount             any         // 赢取的金额
	ValidAmount           any         //
	JackpotAmount         any         // 奖池金额
	BetStatus             any         // 投注状态
	VendorSettleTime      any         // 供应商结算时间
	IsFreeSpin            any         // 是否免费旋转
	VendorBetId           any         // 供应商投注ID
	Win                   any         // 赢
	Fee                   any         // 手续费
	ReservedAmount        any         // 预留金额
	Finished              any         // 是否完成
	EnterMoney            any         // 进入金额
	RoundBeginTime        *gtime.Time // 局开始时间
	BetTime               *gtime.Time //
	RoundEndTime          *gtime.Time // 局结束时间
	JackpotAmount         any         // 奖池金额
	Ip                    any         // IP地址
	Currency              any         // 币种
	CreatedAt             *gtime.Time // 创建时间
	UpdatedAt             *gtime.Time // 更新时间
	AgentId               any         // 代理ID，厅给站开的代理
	SiteCode              any         //
	ChannelCode           any         //
}
