package game

import (
	"context"
	v1 "jh_game_service/api/backend/game/v1"
	"jh_game_service/internal/middleware"
	"jh_game_service/internal/service/backend"
	"strings"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

// Controller 游戏控制器
type Controller struct {
	v1.UnimplementedGameServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterGameServer(s.Server, &Controller{})
}

// GetGames 获取游戏列表
func (*Controller) GetGames(ctx context.Context, req *v1.GetGamesReq) (res *v1.GetGamesRes, err error) {
	return backend.Game().GetGames(ctx, req)
}

// UpdateGames 修改站点游戏开关
func (*Controller) UpdateGames(ctx context.Context, req *v1.UpdateGamesReq) (res *v1.UpdateGamesRes, err error) {
	return backend.Game().UpdateGames(ctx, req)
}

// GetGameRecords 获取游戏记录
func (*Controller) GetGameRecords(ctx context.Context, req *v1.GetGameRecordsReq) (res *v1.GetGameRecordsRes, err error) {
	return backend.Game().GetGameRecords(ctx, req)
}

// SupplementBetRecords 补单
func (*Controller) SupplementBetRecords(ctx context.Context, req *v1.SupplementBetRecordsReq) (res *v1.SupplementBetRecordsRes, err error) {
	return backend.Game().SupplementBetRecords(ctx, req)
}

// GetGameList 获取游戏列表
func (*Controller) GetGameList(ctx context.Context, req *v1.GetGameListReq) (res *v1.GetGameListRes, err error) {
	return backend.Game().GetGameList(ctx, req)
}

// UpdateGameListItem 更新游戏列表项
func (*Controller) UpdateGameListItem(ctx context.Context, req *v1.UpdateGameListItemReq) (res *v1.UpdateGameListItemRes, err error) {
	return backend.Game().UpdateGameListItem(ctx, req)
}

// GetGamePlatforms 获取游戏平台列表
func (*Controller) GetGamePlatforms(ctx context.Context, req *v1.GetGamePlatformsReq) (res *v1.GetGamePlatformsRes, err error) {
	return backend.Game().GetGamePlatforms(ctx, req)
}

// GetUserBetStats 获取用户注单统计
func (*Controller) GetUserBetStats(ctx context.Context, req *v1.GetUserBetStatsReq) (res *v1.GetUserBetStatsRes, err error) {
	return backend.Game().GetUserBetStats(ctx, req)
}

// GetUserBetSummary 获取用户时间范围注单汇总
func (*Controller) GetUserBetSummary(ctx context.Context, req *v1.GetUserBetSummaryReq) (res *v1.GetUserBetSummaryRes, err error) {
	if req == nil {
		return &v1.GetUserBetSummaryRes{
			Success: false,
			Message: "请求参数不能为空",
		}, nil
	}

	if req.GetStartTime() == "" || req.GetEndTime() == "" {
		return &v1.GetUserBetSummaryRes{
			Success: false,
			Message: "开始时间和结束时间不能为空",
		}, nil
	}

	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return &v1.GetUserBetSummaryRes{
			Success: false,
			Message: "site_id 缺失或无效",
		}, nil
	}
	siteCode, ok := middleware.GetSiteCodeFromContext(ctx)
	if !ok || siteCode == "" {
		return &v1.GetUserBetSummaryRes{
			Success: false,
			Message: "site_code 缺失或无效",
		}, nil
	}
	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok || userId == 0 {
		return &v1.GetUserBetSummaryRes{
			Success: false,
			Message: "user_id 缺失或无效",
		}, nil
	}

	return backend.Game().GetUserBetSummary(ctx, req)
}

// GetUserBetIncrementBatch 批量获取用户注单增量（MySQL，基于用户游标）
func (*Controller) GetUserBetIncrementBatch(ctx context.Context, req *v1.GetUserBetIncrementBatchReq) (res *v1.GetUserBetIncrementBatchRes, err error) {
	if req == nil {
		return &v1.GetUserBetIncrementBatchRes{
			Success: false,
			Message: "请求参数不能为空",
		}, nil
	}

	if req.GetSiteId() <= 0 {
		return &v1.GetUserBetIncrementBatchRes{
			Success: false,
			Message: "site_id 缺失或无效",
		}, nil
	}
	if strings.TrimSpace(req.GetSiteCode()) == "" {
		return &v1.GetUserBetIncrementBatchRes{
			Success: false,
			Message: "site_code 缺失或无效",
		}, nil
	}
	if len(req.GetCursors()) == 0 {
		return &v1.GetUserBetIncrementBatchRes{
			Success: true,
			Message: "查询成功",
		}, nil
	}

	// 兜底设置动态数据库路由，避免调用方未传metadata导致路由失败。
	ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+strings.TrimSpace(req.GetSiteCode()))

	return backend.Game().GetUserBetIncrementBatch(ctx, req)
}

// GetSyncTasks 获取同步任务列表
func (*Controller) GetSyncTasks(ctx context.Context, req *v1.GetSyncTasksReq) (res *v1.GetSyncTasksRes, err error) {
	return backend.Game().GetSyncTasks(ctx, req)
}

// GetAgentDailyBetStats 获取代理日统计（供结算）
func (*Controller) GetAgentDailyBetStats(ctx context.Context, req *v1.GetAgentDailyBetStatsReq) (res *v1.GetAgentDailyBetStatsRes, err error) {
	return backend.Game().GetAgentDailyBetStats(ctx, req)
}
