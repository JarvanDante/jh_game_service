package game

import (
	"context"
	v1 "jh_game_service/api/frontend/game/v1"
	"jh_game_service/internal/service/frontend"

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
