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
