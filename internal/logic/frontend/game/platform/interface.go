package platform

import (
	"context"
	v1 "jh_game_service/api/frontend/game/v1"
	"jh_game_service/internal/model/entity"
)

// GetGameConfigFunc 获取游戏配置的函数类型（用于依赖注入，避免循环导入）
var GetGameConfigFunc func(ctx context.Context, siteId int, gameId int) (*entity.SiteGame, error)

// GamePlatform 游戏平台接口
// 所有游戏平台都需要实现这个接口
type GamePlatform interface {

	//
	// GetGameUrl
	//  @Description: 获取游戏链接（登录/进入游戏）
	//  @param ctx
	//  @param params
	//  @return map[string]interface{}
	//  @return error
	//
	GetGameUrl(ctx context.Context, params *GameParams) (map[string]interface{}, error)

	//
	// GetBalance
	//  @Description: 查询游戏余额
	//  @param ctx
	//  @param params
	//  @return float64
	//  @return error
	//
	GetBalance(ctx context.Context, params *GameParams) (float64, error)

	//
	// TransferIn
	//  @Description: 转入游戏（从中心钱包转到游戏钱包）
	//  @param ctx
	//  @param params
	//  @return error
	//
	TransferIn(ctx context.Context, params *TransferParams) error

	//
	// TransferOut
	//  @Description: 转出游戏（从游戏钱包转到中心钱包）
	//  @param ctx
	//  @param params
	//  @return error
	//
	TransferOut(ctx context.Context, params *TransferParams) error

	//
	// PullBetRecords
	//  @Description: 拉取注单 ./jh_game_service pull ONE
	// 	              注单偿 ./jh_game_service pull ONE "2026-01-23 19:00:00" "2026-01-24 01:00:00"
	//  @param ctx
	//  @param params
	//  @return []BetRecord
	//  @return error
	//
	PullBetRecords(ctx context.Context, params *PullRecordsParams) ([]BetRecord, error)

	//
	// GetGameSource
	//  @Description: 获取游戏图片资源同步数据库
	//  @param ctx
	//  @param params
	//  @return error
	//
	GetGameSource(ctx context.Context, params *GetGameSourceParams) error
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
	SiteCode  string          // 站点代码
	StartTime string          // 开始时间
	EndTime   string          // 结束时间
	Page      int             // 页码
	PageSize  int             // 每页数量
	SiteGame  entity.SiteGame // 游戏配置
}

// GetGameSourceParams 游戏资源参数
type GetGameSourceParams struct {
	SiteId   int             // 站点ID
	SiteCode string          // 站点代码
	GameId   int             // 游戏ID
	SiteGame entity.SiteGame // 游戏配置
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
