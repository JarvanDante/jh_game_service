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
		ToggleFavoriteGame(ctx context.Context, req *v1.ToggleFavoriteGameReq) (*v1.ToggleFavoriteGameRes, error)
		GetFavoriteGameList(ctx context.Context, req *v1.GetFavoriteGameListReq) (*v1.GetFavoriteGameListRes, error)
		GetGameBalance(ctx context.Context, req *v1.GetGameBalanceReq) (*v1.GetGameBalanceRes, error)
		RecallGameBalance(ctx context.Context, req *v1.RecallGameBalanceReq) (*v1.RecallGameBalanceRes, error)
		GetUserBetHistory(ctx context.Context, req *v1.GetUserBetHistoryReq) (*v1.GetUserBetHistoryRes, error)
		GetWashCodeQuota(ctx context.Context, req *v1.GetWashCodeQuotaReq) (*v1.GetWashCodeQuotaRes, error)
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
