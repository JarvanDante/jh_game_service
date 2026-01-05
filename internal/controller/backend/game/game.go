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
