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

// GetGameRecords 获取游戏记录
func (s *sGame) GetGameRecords(ctx context.Context, req *v1.GetGameRecordsReq) (*v1.GetGameRecordsRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "game.GetGameRecords", trace.WithAttributes(
		attribute.String("method", "GetGameRecords"),
		attribute.Int("site_id", int(req.SiteId)),
		attribute.String("platform", req.Platform),
		attribute.String("start_date", req.StartDate),
		attribute.String("end_date", req.EndDate),
		attribute.Int("page", int(req.Page)),
		attribute.Int("size", int(req.Size)),
	))
	defer span.End()

	// 参数验证
	if req.Platform == "" {
		return nil, fmt.Errorf("游戏平台不能为空")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	if req.Size > 100 {
		req.Size = 100 // 限制最大页面大小
	}

	// 设置默认时间范围（如果未提供）
	startDate := req.StartDate
	endDate := req.EndDate
	if startDate == "" {
		startDate = "2024-01-01 00:00:00"
	}
	if endDate == "" {
		endDate = "2026-12-31 23:59:59"
	}

	// 构建查询条件
	query := dao.EgameBetLog.Ctx(ctx).
		Where("platform_code = ?", req.Platform).
		Where("bet_time >= ?", startDate).
		Where("bet_time <= ?", endDate)

	// 添加用户名过滤
	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}

	// 添加记录类型过滤
	if req.Type > 0 {
		if req.Type == 1 {
			// 投注记录
			query = query.Where("bet_amount > 0")
		} else if req.Type == 2 {
			// 中奖记录
			query = query.Where("win_amount > 0")
		}
	}

	// 添加关键词搜索
	if req.Keywords != "" {
		query = query.Where("(order_id LIKE ? OR round_id LIKE ? OR game_name LIKE ?)",
			"%"+req.Keywords+"%", "%"+req.Keywords+"%", "%"+req.Keywords+"%")
	}

	// 获取总记录数
	total, err := query.Count()
	if err != nil {
		return nil, fmt.Errorf("获取记录总数失败: %v", err)
	}

	// 计算分页信息
	offset := (req.Page - 1) * req.Size
	lastPage := int32((total + int(req.Size) - 1) / int(req.Size))

	// 获取分页数据
	var gameRecords []entity.EgameBetLog
	err = query.
		Order("bet_time DESC").
		Limit(int(req.Size)).
		Offset(int(offset)).
		Scan(&gameRecords)
	if err != nil {
		return nil, fmt.Errorf("获取游戏记录失败: %v", err)
	}

	// 计算当前平台总金额
	totalAmountQuery := dao.EgameBetLog.Ctx(ctx).
		Where("platform_code = ?", req.Platform).
		Where("bet_time >= ?", startDate).
		Where("bet_time <= ?", endDate)

	if req.Username != "" {
		totalAmountQuery = totalAmountQuery.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Type > 0 {
		if req.Type == 1 {
			totalAmountQuery = totalAmountQuery.Where("bet_amount > 0")
		} else if req.Type == 2 {
			totalAmountQuery = totalAmountQuery.Where("win_amount > 0")
		}
	}
	if req.Keywords != "" {
		totalAmountQuery = totalAmountQuery.Where("(order_id LIKE ? OR round_id LIKE ? OR game_name LIKE ?)",
			"%"+req.Keywords+"%", "%"+req.Keywords+"%", "%"+req.Keywords+"%")
	}

	var totalAmount float64
	err = totalAmountQuery.Fields("COALESCE(SUM(bet_amount), 0)").Scan(&totalAmount)
	if err != nil {
		totalAmount = 0
	}

	// 计算所有游戏总金额（暂时写死，后续可以通过快照获取）
	allGameTotalAmount := totalAmount * 1.2 // 模拟所有游戏总金额

	// 转换为响应格式
	records := make([]*v1.GameRecord, 0, len(gameRecords))
	for _, record := range gameRecords {
		// 处理时间格式
		betTime := ""
		if record.BetTime != nil {
			betTime = record.BetTime.Format("2006-01-02 15:04:05")
		}

		settleTime := ""
		if record.SettleTime != nil {
			settleTime = record.SettleTime.Format("2006-01-02 15:04:05")
		}

		createdAt := ""
		if record.CreatedAt != nil {
			createdAt = record.CreatedAt.Format("2006-01-02 15:04:05")
		}

		updatedAt := ""
		if record.UpdatedAt != nil {
			updatedAt = record.UpdatedAt.Format("2006-01-02 15:04:05")
		}

		gameRecord := &v1.GameRecord{
			Id:           record.Id,
			SiteId:       int32(record.SiteId),
			UserId:       record.UserId,
			Username:     record.Username,
			PlatformCode: record.PlatformCode,
			GameCode:     record.GameCode,
			GameName:     record.GameName,
			GameType:     record.GameType,
			OrderId:      record.OrderId,
			RoundId:      record.RoundId,
			BetAmount:    fmt.Sprintf("%.6f", record.BetAmount),
			WinAmount:    fmt.Sprintf("%.6f", record.WinAmount),
			Currency:     record.Currency,
			BetTime:      betTime,
			SettleTime:   settleTime,
			Status:       int32(record.Status),
			RawData:      record.RawData,
			Remark:       record.Remark,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		}
		records = append(records, gameRecord)
	}

	return &v1.GetGameRecordsRes{
		Data:               records,
		Total:              int32(total),
		CurrentPage:        req.Page,
		PerPage:            req.Size,
		LastPage:           lastPage,
		TotalAmount:        fmt.Sprintf("%.2f", totalAmount),
		AllGameTotalAmount: fmt.Sprintf("%.2f", allGameTotalAmount),
	}, nil
}
