package game

import (
	"context"
	v1 "jh_game_service/api/frontend/game/v1"
	"jh_game_service/internal/service/frontend"
	"time"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	v1.UnimplementedGameServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterGameServer(s.Server, &Controller{})
}

// RedirectToGame 进入游戏
func (*Controller) RedirectToGame(ctx context.Context, req *v1.RedirectToGameReq) (res *v1.RedirectToGameRes, err error) {
	return frontend.Game().RedirectToGame(ctx, req)
}

// GetGameCategories 获取游戏种类
func (*Controller) GetGameCategories(ctx context.Context, req *v1.GetGameCategoriesReq) (res *v1.GetGameCategoriesRes, err error) {
	return frontend.Game().GetGameCategories(ctx, req)
}

// GetGameList 获取游戏列表
func (*Controller) GetGameList(ctx context.Context, req *v1.GetGameListReq) (res *v1.GetGameListRes, err error) {
	return frontend.Game().GetGameList(ctx, req)
}

// GetGameBalance 获取游戏余额
func (*Controller) GetGameBalance(ctx context.Context, req *v1.GetGameBalanceReq) (res *v1.GetGameBalanceRes, err error) {
	return frontend.Game().GetGameBalance(ctx, req)
}

// RecallGameBalance 一键回收游戏余额
func (*Controller) RecallGameBalance(ctx context.Context, req *v1.RecallGameBalanceReq) (res *v1.RecallGameBalanceRes, err error) {
	return frontend.Game().RecallGameBalance(ctx, req)
}

// GetUserBetHistory 获取用户历史投注
func (*Controller) GetUserBetHistory(ctx context.Context, req *v1.GetUserBetHistoryReq) (res *v1.GetUserBetHistoryRes, err error) {
	if req == nil {
		req = &v1.GetUserBetHistoryReq{}
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 5
	}
	if req.Size > 100 {
		req.Size = 100
	}

	if req.StartTime == "" || req.EndTime == "" {
		now := time.Now()
		if req.StartTime == "" {
			req.StartTime = now.Format("2006-01-02") + " 00:00:00"
		}
		if req.EndTime == "" {
			req.EndTime = now.Format("2006-01-02") + " 23:59:59"
		}
	}

	return frontend.Game().GetUserBetHistory(ctx, req)
}

// GetWashCodeQuota 获取当前用户可洗码额度
func (*Controller) GetWashCodeQuota(ctx context.Context, req *v1.GetWashCodeQuotaReq) (res *v1.GetWashCodeQuotaRes, err error) {
	if req == nil {
		req = &v1.GetWashCodeQuotaReq{}
	}
	return frontend.Game().GetWashCodeQuota(ctx, req)
}
