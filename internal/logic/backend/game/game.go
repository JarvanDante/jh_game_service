package game

import (
	"context"
	"fmt"
	v1 "jh_game_service/api/backend/game/v1"
	"jh_game_service/internal/dao"
	"jh_game_service/internal/model/do"
	"jh_game_service/internal/model/entity"
	"jh_game_service/internal/service/backend"
	"jh_game_service/internal/tracing"

	"github.com/gogf/gf/v2/database/gdb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type (
	sGame struct{}
)

func init() {
	backend.RegisterGame(&sGame{})
}

// GetGames 获取游戏列表
func (s *sGame) GetGames(ctx context.Context, req *v1.GetGamesReq) (*v1.GetGamesRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "game.GetGames", trace.WithAttributes(
		attribute.String("method", "GetGames"),
		attribute.Int("site_id", int(req.SiteId)),
		attribute.Bool("search_balance", req.SearchBalance),
	))
	defer span.End()

	siteId := req.SiteId
	if siteId <= 0 {
		siteId = 1 // 默认站点ID
	}

	// 获取所有启用的游戏
	var games []entity.Game
	err := dao.Game.Ctx(ctx).Where(do.Game{Status: 1}).Scan(&games)
	if err != nil {
		return nil, fmt.Errorf("获取游戏列表失败: %v", err)
	}

	// 获取站点游戏配置
	var siteGames []entity.SiteGame
	err = dao.SiteGame.Ctx(ctx).
		Where(do.SiteGame{
			SiteId:      int(siteId),
			IsAvailable: 1,
		}).
		Order("type ASC, sort ASC").
		Scan(&siteGames)
	if err != nil {
		return nil, fmt.Errorf("获取站点游戏配置失败: %v", err)
	}

	// 构建游戏ID到游戏信息的映射
	gameMap := make(map[int]*entity.Game)
	for i := range games {
		gameMap[int(games[i].Id)] = &games[i]
	}

	// 按类型分组
	gamesByType := make(map[int32]*v1.GameTypeList)
	totalCount := int32(0)

	for _, siteGame := range siteGames {
		game, exists := gameMap[siteGame.GameId]
		if !exists {
			continue
		}

		gameInfo := &v1.GameInfo{
			Id:        int32(game.Id),
			Name:      game.Name,
			Status:    int32(siteGame.Status),
			Sort:      int32(siteGame.Sort),
			Type:      int32(game.Type),
			Platform:  game.Platform,
			Code:      game.Code,
			ImageCode: game.ImageCode,
			DbName:    game.DbName,
		}

		gameType := int32(game.Type)
		if gamesByType[gameType] == nil {
			gamesByType[gameType] = &v1.GameTypeList{
				Games: make([]*v1.GameInfo, 0),
			}
		}
		gamesByType[gameType].Games = append(gamesByType[gameType].Games, gameInfo)
		totalCount++
	}

	return &v1.GetGamesRes{
		Data:  gamesByType,
		Count: totalCount,
	}, nil
}

// UpdateGames 修改站点游戏开关
func (s *sGame) UpdateGames(ctx context.Context, req *v1.UpdateGamesReq) (*v1.UpdateGamesRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "game.UpdateGames", trace.WithAttributes(
		attribute.String("method", "UpdateGames"),
		attribute.Int("site_id", int(req.SiteId)),
		attribute.Int("data_count", len(req.Data)),
	))
	defer span.End()

	if len(req.Data) == 0 {
		return &v1.UpdateGamesRes{
			Success: false,
			Message: "游戏数据为空",
		}, nil
	}

	siteId := req.SiteId
	if siteId <= 0 {
		siteId = 1 // 默认站点ID
	}

	// 使用事务处理
	err := dao.SiteGame.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, gameUpdate := range req.Data {
			if gameUpdate == nil {
				continue
			}

			// 更新站点游戏配置
			_, err := dao.SiteGame.Ctx(ctx).TX(tx).
				Where(do.SiteGame{
					SiteId:      int(siteId),
					GameId:      int(gameUpdate.Id),
					IsAvailable: 1,
				}).
				Data(do.SiteGame{
					Status: int(gameUpdate.Status),
					Sort:   int(gameUpdate.Sort),
				}).
				Update()

			if err != nil {
				return fmt.Errorf("更新游戏配置失败, gameId: %d, error: %v", gameUpdate.Id, err)
			}
		}
		return nil
	})

	if err != nil {
		return &v1.UpdateGamesRes{
			Success: false,
			Message: fmt.Sprintf("设置游戏开关失败: %v", err),
		}, nil
	}

	return &v1.UpdateGamesRes{
		Success: true,
		Message: "设置游戏开关成功",
	}, nil
}
