package frontend

import (
	"context"
	v1 "jh_game_service/api/frontend/game/v1"
)

type (
	IGame interface {
		RedirectToGame(ctx context.Context, req *v1.RedirectToGameReq) (*v1.RedirectToGameRes, error)
		GetGameCategories(ctx context.Context, req *v1.GetGameCategoriesReq) (*v1.GetGameCategoriesRes, error)
		GetGameList(ctx context.Context, req *v1.GetGameListReq) (*v1.GetGameListRes, error)
	}
)

var (
	localGame IGame
)

func Game() IGame {
	if localGame == nil {
		panic("implement not found for interface IGame, forgot register?")
	}
	return localGame
}

func RegisterGame(i IGame) {
	localGame = i
}
