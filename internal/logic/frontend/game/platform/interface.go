package platform

import (
	"context"
	v1 "jh_game_service/api/frontend/game/v1"
	"jh_game_service/internal/model/entity"
)

// GamePlatform 游戏平台接口
// 所有游戏平台都需要实现这个接口
type GamePlatform interface {
	// GetGameUrl 获取游戏链接（登录/进入游戏）
	// 返回: map["type"]="url/html", map["url"]="游戏链接"
	GetGameUrl(ctx context.Context, params *GameParams) (map[string]interface{}, error)

	// GetBalance 查询游戏余额
	GetBalance(ctx context.Context, params *GameParams) (float64, error)

	// TransferIn 转入游戏（从中心钱包转到游戏钱包）
	TransferIn(ctx context.Context, params *TransferParams) error

	// TransferOut 转出游戏（从游戏钱包转到中心钱包）
	TransferOut(ctx context.Context, params *TransferParams) error

	// PullBetRecords 拉取注单
	// startTime和endTime格式: "2006-01-02 15:04:05"
	PullBetRecords(ctx context.Context, params *PullRecordsParams) ([]BetRecord, error)

	// SupplementBetRecords 补单（定时任务调用）
	// 用于补充拉单失败或遗漏的注单
	SupplementBetRecords(ctx context.Context, params *SupplementParams) error
}

// GameParams 游戏参数
type GameParams struct {
	SiteId   int                   // 站点ID
	UserId   uint                  // 用户ID
	Username string                // 用户名
	SiteCode string                // 站点代码
	Currency string                // 货币
	Domain   string                // 域名
	ClientIP string                // 客户端IP
	SiteGame entity.SiteGame       // 游戏配置
	Request  *v1.RedirectToGameReq // 原始请求
}

// TransferParams 转账参数
type TransferParams struct {
	SiteId   int             // 站点ID
	UserId   uint            // 用户ID
	Username string          // 用户名
	SiteCode string          // 站点代码
	Amount   float64         // 金额
	Currency string          // 货币
	OrderNo  string          // 订单号
	SiteGame entity.SiteGame // 游戏配置
}

// PullRecordsParams 拉单参数
type PullRecordsParams struct {
	SiteId    int             // 站点ID
	StartTime string          // 开始时间
	EndTime   string          // 结束时间
	Page      int             // 页码
	PageSize  int             // 每页数量
	SiteGame  entity.SiteGame // 游戏配置
}

// SupplementParams 补单参数
type SupplementParams struct {
	SiteId    int             // 站点ID
	StartTime string          // 开始时间
	EndTime   string          // 结束时间
	SiteGame  entity.SiteGame // 游戏配置
}

// BetRecord 注单记录
type BetRecord struct {
	OrderNo     string  // 订单号
	Username    string  // 用户名
	GameName    string  // 游戏名称
	BetAmount   float64 // 投注金额
	WinAmount   float64 // 赢得金额
	ValidAmount float64 // 有效投注
	BetTime     string  // 投注时间
	SettleTime  string  // 结算时间
	Status      int     // 状态: 0=未结算 1=已结算 2=已取消
	RawData     string  // 原始数据(JSON)
}
