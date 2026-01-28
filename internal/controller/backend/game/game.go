package game

import (
	"context"
	v1 "jh_game_service/api/backend/game/v1"
	"jh_game_service/internal/service/backend"

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

// GetSyncTasks 获取同步任务列表
func (*Controller) GetSyncTasks(ctx context.Context, req *v1.GetSyncTasksReq) (res *v1.GetSyncTasksRes, err error) {
	return backend.Game().GetSyncTasks(ctx, req)
}
