package game

import (
	"context"
	"fmt"
	"jh_game_service/internal/dao"
	"jh_game_service/internal/model/do"
	"jh_game_service/internal/model/entity"
)

// GetGameConfigById 根据gameId查询游戏配置（供平台内部使用）
func GetGameConfigById(ctx context.Context, siteId int, gameId int) (*entity.SiteGame, error) {
	var siteGame entity.SiteGame
	err := dao.SiteGame.Ctx(ctx).
		Where(do.SiteGame{
			SiteId: siteId,
			GameId: gameId,
		}).
		Scan(&siteGame)
	if err != nil {
		return nil, fmt.Errorf("查询游戏配置失败: %v", err)
	}
	if siteGame.Id == 0 {
		return nil, fmt.Errorf("游戏不存在: gameId=%d", gameId)
	}
	return &siteGame, nil
}
