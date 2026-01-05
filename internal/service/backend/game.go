package backend

import (
	"context"
	v1 "jh_game_service/api/backend/game/v1"
)

type (
	IGame interface {
		GetGames(ctx context.Context, req *v1.GetGamesReq) (*v1.GetGamesRes, error)
		UpdateGames(ctx context.Context, req *v1.UpdateGamesReq) (*v1.UpdateGamesRes, error)
		GetGameRecords(ctx context.Context, req *v1.GetGameRecordsReq) (*v1.GetGameRecordsRes, error)
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
