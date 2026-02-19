package game

import (
	"context"
	"fmt"
	v1 "jh_game_service/api/backend/game/v1"
	"jh_game_service/internal/dao"
	"jh_game_service/internal/domain"
	"jh_game_service/internal/logic/frontend/game/platform"
	"jh_game_service/internal/logic/sync_task"
	"jh_game_service/internal/middleware"
	"jh_game_service/internal/model/do"
	"jh_game_service/internal/model/entity"
	"jh_game_service/internal/registry"
	"jh_game_service/internal/service/backend"
	"jh_game_service/internal/tracing"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type (
	sGame struct{}
)

func init() {
	backend.RegisterGame(&sGame{})
}

// convertToFloat64 将interface{}转换为float64
func convertToFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}

	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case uint64:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	case column.Decimal:
		// ClickHouse Decimal类型，尝试转换
		if s := fmt.Sprintf("%v", val); s != "" {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				return f
			}
		}
	default:
		// 尝试通过fmt.Sprintf转换
		if s := fmt.Sprintf("%v", val); s != "" {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				return f
			}
		}
	}
	return 0
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

	// 直接从site_game表获取站点游戏配置
	var siteGames []entity.SiteGame
	err := dao.SiteGame.Ctx(ctx).
		Where(do.SiteGame{
			SiteId: int(siteId),
		}).
		Order("type ASC, sort ASC").
		Scan(&siteGames)
	if err != nil {
		return nil, fmt.Errorf("获取站点游戏配置失败: %v", err)
	}

	// 按类型分组
	gamesByType := make(map[int32]*v1.GameTypeList)
	totalCount := int32(0)

	for _, siteGame := range siteGames {
		gameInfo := &v1.GameInfo{
			Id:        int32(siteGame.GameId),
			Name:      siteGame.Name,
			Status:    int32(siteGame.Status),
			Sort:      int32(siteGame.Sort),
			Type:      int32(siteGame.Type),
			Platform:  siteGame.Platform,
			Code:      "",
			ImageCode: siteGame.ImageCode,
			DbName:    siteGame.DbName,
		}

		gameType := int32(siteGame.Type)
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

	// 使用事务处理 - 使用 Ctx(ctx) 保持动态数据库路由
	err := dao.SiteGame.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, gameUpdate := range req.Data {
			if gameUpdate == nil {
				continue
			}

			// 更新站点游戏配置
			_, err := dao.SiteGame.Ctx(ctx).TX(tx).
				Where(do.SiteGame{
					SiteId: int(siteId),
					GameId: int(gameUpdate.Id),
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
	query := dao.BetRecords.Ctx(ctx).
		Where("bet_time >= ?", startDate).
		Where("bet_time <= ?", endDate)

	// 如果指定了平台，添加平台过滤
	if req.Platform != "" {
		query = query.Where("platform = ?", req.Platform)
	}

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
		query = query.Where("(order_no LIKE ? OR round_id LIKE ? OR game_name LIKE ?)",
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
	var betRecords []entity.BetRecords
	err = query.
		Order("bet_time DESC").
		Limit(int(req.Size)).
		Offset(int(offset)).
		Scan(&betRecords)
	if err != nil {
		return nil, fmt.Errorf("获取游戏记录失败: %v", err)
	}

	// 转换为响应格式
	records := make([]*v1.GameRecord, 0, len(betRecords))
	for _, record := range betRecords {
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
			UserId:       int64(record.UserId),
			Username:     record.Username,
			PlatformCode: record.Platform,
			GameCode:     "",
			GameName:     record.GameName,
			GameType:     record.GameType,
			OrderId:      record.OrderNo,
			RoundId:      record.RoundId,
			BetAmount:    fmt.Sprintf("%.6f", record.BetAmount),
			WinAmount:    fmt.Sprintf("%.6f", record.WinAmount),
			Currency:     record.Currency,
			BetTime:      betTime,
			SettleTime:   settleTime,
			Status:       int32(record.Status),
			RawData:      record.RawData,
			Remark:       "",
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		}
		records = append(records, gameRecord)
	}

	return &v1.GetGameRecordsRes{
		Data:        records,
		Total:       int32(total),
		CurrentPage: req.Page,
		PerPage:     req.Size,
		LastPage:    lastPage,
	}, nil
}

// SupplementBetRecords 补单
func (s *sGame) SupplementBetRecords(ctx context.Context, req *v1.SupplementBetRecordsReq) (*v1.SupplementBetRecordsRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "game.SupplementBetRecords", trace.WithAttributes(
		attribute.String("method", "SupplementBetRecords"),
		attribute.String("platform", req.Platform),
		attribute.String("start_time", req.StartTime),
		attribute.String("end_time", req.EndTime),
	))
	defer span.End()

	// 参数验证
	if req.Platform == "" {
		return &v1.SupplementBetRecordsRes{
			Success: false,
			Message: "游戏平台不能为空",
		}, nil
	}
	if req.StartTime == "" || req.EndTime == "" {
		return &v1.SupplementBetRecordsRes{
			Success: false,
			Message: "开始时间和结束时间不能为空",
		}, nil
	}

	// 获取所有站点配置
	dbManager := registry.GetDBConfigManager()
	siteCodes := dbManager.GetAllSiteCodes()

	totalRecords := 0
	successSites := 0
	failedSites := 0

	// 遍历所有站点
	for _, siteCode := range siteCodes {
		// 获取站点配置
		siteConfig, err := dbManager.GetSiteConfig(siteCode)
		if err != nil {
			g.Log().Errorf(ctx, "获取站点配置失败: 站点=%s, 错误=%v", siteCode, err)
			failedSites++
			continue
		}

		siteId := siteConfig.Base.SiteID

		// 查找该站点是否有指定平台的游戏
		var targetGame *registry.GameConfig
		for i := range siteConfig.Game {
			if siteConfig.Game[i].Platform == req.Platform {
				targetGame = &siteConfig.Game[i]
				break
			}
		}

		if targetGame == nil {
			continue // 该站点没有此平台的游戏，跳过
		}

		g.Log().Infof(ctx, "开始补单: 站点=%s, 平台=%s, 游戏=%s", siteCode, req.Platform, targetGame.Name)

		// 创建同步任务
		taskId, err := sync_task.CreateTask(ctx, siteId, siteCode, req.Platform, 2, req.StartTime, req.EndTime)
		if err != nil {
			g.Log().Errorf(ctx, "创建同步任务失败: 站点=%s, 错误=%v", siteCode, err)
			taskId = 0 // 继续执行，不中断补单
		}

		// 创建游戏平台实例
		gamePlatform, err := platform.GetFactory().Create(req.Platform)
		if err != nil {
			g.Log().Errorf(ctx, "创建游戏平台实例失败: 站点=%s, 错误=%v", siteCode, err)
			if taskId > 0 {
				sync_task.CompleteTask(ctx, taskId, 3, fmt.Sprintf("创建游戏平台实例失败: %v", err))
			}
			failedSites++
			continue
		}

		// 构建拉单参数
		params := &platform.PullRecordsParams{
			SiteId:   siteId,
			SiteCode: siteCode,
			SiteGame: entity.SiteGame{
				GameId:   targetGame.GameId,
				Type:     targetGame.Type,
				Name:     targetGame.Name,
				Platform: targetGame.Platform,
				DbName:   targetGame.DbName,
				ApiId:    targetGame.ApiId,
				ApiKey:   targetGame.ApiKey,
				ApiUrl:   targetGame.ApiUrl,
			},
			StartTime: req.StartTime,
			EndTime:   req.EndTime,
			TaskId:    taskId, // 传递任务ID
		}

		// 调用拉单方法
		records, err := gamePlatform.PullBetRecords(ctx, params)
		if err != nil {
			g.Log().Errorf(ctx, "拉取注单失败: 站点=%s, 错误=%v", siteCode, err)
			if taskId > 0 {
				sync_task.CompleteTask(ctx, taskId, 3, fmt.Sprintf("拉取注单失败: %v", err))
			}
			failedSites++
			continue
		}

		recordCount := len(records)

		// 添加详细日志
		if recordCount == 0 {
			g.Log().Warningf(ctx, "补单返回0条记录: 站点=%s, 平台=%s, 时间=%s~%s",
				siteCode, req.Platform, req.StartTime, req.EndTime)
		} else {
			g.Log().Infof(ctx, "补单拉取成功: 站点=%s, 记录数=%d", siteCode, recordCount)
		}

		// 完成任务
		if taskId > 0 {
			sync_task.CompleteTask(ctx, taskId, 2, "")
		}

		totalRecords += recordCount
		successSites++

		g.Log().Infof(ctx, "补单完成: 站点=%s, 记录数=%d", siteCode, recordCount)
	}

	message := fmt.Sprintf("补单完成: 平台=%s, 成功站点=%d, 失败站点=%d, 总记录数=%d",
		req.Platform, successSites, failedSites, totalRecords)

	return &v1.SupplementBetRecordsRes{
		Success:      successSites > 0,
		Message:      message,
		TotalRecords: int32(totalRecords),
	}, nil
}

// GetGameList 获取游戏列表
func (s *sGame) GetGameList(ctx context.Context, req *v1.GetGameListReq) (*v1.GetGameListRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "game.GetGameList", trace.WithAttributes(
		attribute.String("method", "GetGameList"),
		attribute.Int("site_id", int(req.SiteId)),
		attribute.String("platform", req.Platform),
		attribute.Int("page", int(req.Page)),
		attribute.Int("size", int(req.Size)),
	))
	defer span.End()

	// 参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	if req.Size > 100 {
		req.Size = 100 // 限制最大页面大小
	}

	siteId := req.SiteId
	if siteId <= 0 {
		siteId = 1 // 默认站点ID
	}

	// 构建查询条件
	query := dao.GameList.Ctx(ctx).Where("site_id = ?", siteId)

	// 平台过滤
	if req.Platform != "" {
		query = query.Where("platform = ?", req.Platform)
	}

	// 运营商过滤
	if req.Vendor != "" {
		query = query.Where("vendor = ?", req.Vendor)
	}

	// 游戏名称模糊搜索
	if req.Name != "" {
		query = query.Where("(name LIKE ? OR alias_name LIKE ?)", "%"+req.Name+"%", "%"+req.Name+"%")
	}

	// 游戏代码过滤
	if req.Code != "" {
		query = query.Where("code = ?", req.Code)
	}

	// 游戏类型过滤
	if req.Style != "" {
		query = query.Where("style = ?", req.Style)
	}

	// 热门过滤
	if req.Hot != -1 {
		query = query.Where("hot = ?", req.Hot)
	}

	// 推荐过滤
	if req.Recommend != -1 {
		query = query.Where("recommend = ?", req.Recommend)
	}

	// 新游戏过滤
	if req.IsNew != -1 {
		query = query.Where("is_new = ?", req.IsNew)
	}

	// 状态过滤
	if req.Status != -1 {
		query = query.Where("status = ?", req.Status)
	}

	// 获取总记录数
	total, err := query.Count()
	if err != nil {
		return nil, fmt.Errorf("获取记录总数失败: %v", err)
	}

	// 计算分页信息
	offset := (req.Page - 1) * req.Size

	// 获取分页数据
	var gameList []entity.GameList
	err = query.
		Order("sort DESC").
		Limit(int(req.Size)).
		Offset(int(offset)).
		Scan(&gameList)
	if err != nil {
		return nil, fmt.Errorf("获取游戏列表失败: %v", err)
	}

	// 转换为响应格式
	list := make([]*v1.GameListItem, 0, len(gameList))
	for _, game := range gameList {
		// 处理时间格式
		createdAt := ""
		if game.CreatedAt != nil {
			createdAt = game.CreatedAt.Format("2006-01-02 15:04:05")
		}

		updatedAt := ""
		if game.UpdatedAt != nil {
			updatedAt = game.UpdatedAt.Format("2006-01-02 15:04:05")
		}

		item := &v1.GameListItem{
			Id:        int32(game.Id),
			SiteId:    int32(game.SiteId),
			Platform:  game.Platform,
			Code:      game.Code,
			Name:      game.Name,
			AliasName: game.AliasName,
			Style:     game.Style,
			Img:       game.Img,
			Img1:      game.Img1,
			Language:  game.Language,
			Device:    game.Device,
			Currency:  "", // 数据库中没有此字段
			Vendor:    game.Vendor,
			Hot:       int32(game.Hot),
			Recommend: int32(game.Recommend),
			IsNew:     int32(game.IsNew),
			Status:    int32(game.Status),
			Sort:      int32(game.Sort),
			LikeNum:   int32(game.LikeNum),
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		list = append(list, item)
	}

	return &v1.GetGameListRes{
		List:  list,
		Total: int32(total),
		Page:  req.Page,
		Size:  req.Size,
	}, nil
}

// UpdateGameListItem 更新游戏列表项
func (s *sGame) UpdateGameListItem(ctx context.Context, req *v1.UpdateGameListItemReq) (*v1.UpdateGameListItemRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "game.UpdateGameListItem", trace.WithAttributes(
		attribute.String("method", "UpdateGameListItem"),
		attribute.Int("site_id", int(req.SiteId)),
		attribute.Int("id", int(req.Id)),
	))
	defer span.End()

	// 参数验证
	if req.Id <= 0 {
		return &v1.UpdateGameListItemRes{
			Success: false,
			Message: "游戏ID不能为空",
		}, nil
	}

	siteId := req.SiteId
	if siteId <= 0 {
		siteId = 1 // 默认站点ID
	}

	// 检查游戏是否存在
	count, err := dao.GameList.Ctx(ctx).
		Where("id = ? AND site_id = ?", req.Id, siteId).
		Count()
	if err != nil {
		return &v1.UpdateGameListItemRes{
			Success: false,
			Message: fmt.Sprintf("查询游戏失败: %v", err),
		}, nil
	}
	if count == 0 {
		return &v1.UpdateGameListItemRes{
			Success: false,
			Message: "游戏不存在",
		}, nil
	}

	// 构建更新数据
	updateData := g.Map{}

	// 只更新后台管理字段
	if req.AliasName != "" {
		updateData["alias_name"] = req.AliasName
	}
	if req.Hot >= 0 {
		updateData["hot"] = req.Hot
	}
	if req.Recommend >= 0 {
		updateData["recommend"] = req.Recommend
	}
	if req.IsNew >= 0 {
		updateData["is_new"] = req.IsNew
	}
	if req.Status >= 0 {
		updateData["status"] = req.Status
	}
	if req.Sort >= 0 {
		updateData["sort"] = req.Sort
	}

	// 如果没有要更新的字段
	if len(updateData) == 0 {
		return &v1.UpdateGameListItemRes{
			Success: false,
			Message: "没有要更新的字段",
		}, nil
	}

	// 执行更新
	_, err = dao.GameList.Ctx(ctx).
		Where("id = ? AND site_id = ?", req.Id, siteId).
		Data(updateData).
		Update()
	if err != nil {
		return &v1.UpdateGameListItemRes{
			Success: false,
			Message: fmt.Sprintf("更新游戏失败: %v", err),
		}, nil
	}

	return &v1.UpdateGameListItemRes{
		Success: true,
		Message: "更新成功",
	}, nil
}

// GetGamePlatforms 获取游戏平台列表
func (s *sGame) GetGamePlatforms(ctx context.Context, req *v1.GetGamePlatformsReq) (*v1.GetGamePlatformsRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "game.GetGamePlatforms", trace.WithAttributes(
		attribute.String("method", "GetGamePlatforms"),
		attribute.Int("site_id", int(req.SiteId)),
		attribute.Int("type", int(req.Type)),
	))
	defer span.End()

	// 构建查询条件
	query := dao.SiteGame.Ctx(ctx).Where("status = ?", 1) // 只查询启用的游戏

	// 如果指定了站点ID，添加站点过滤
	if req.SiteId > 0 {
		query = query.Where("site_id = ?", req.SiteId)
	}

	// 如果指定了type，添加type过滤
	if req.Type > 0 {
		query = query.Where("type = ?", req.Type)
	}

	// 按sort降序排序
	query = query.Order("sort DESC")

	// 查询游戏列表
	var siteGames []entity.SiteGame
	err := query.Scan(&siteGames)
	if err != nil {
		return nil, fmt.Errorf("查询游戏平台失败: %v", err)
	}

	// 转换为响应格式
	platforms := make([]*v1.GamePlatform, 0, len(siteGames))
	for _, game := range siteGames {
		platforms = append(platforms, &v1.GamePlatform{
			GameId:    int32(game.GameId),
			Name:      game.Name,
			Platform:  game.Platform,
			Type:      int32(game.Type),
			ImageCode: game.ImageCode,
			DbName:    game.DbName,
			Sort:      int32(game.Sort),
			Status:    int32(game.Status),
		})
	}

	return &v1.GetGamePlatformsRes{
		Platforms: platforms,
	}, nil
}

// GetUserBetStats 获取用户注单统计（从 ClickHouse）
func (s *sGame) GetUserBetStats(ctx context.Context, req *v1.GetUserBetStatsReq) (*v1.GetUserBetStatsRes, error) {
	if req == nil {
		return &v1.GetUserBetStatsRes{
			Success: false,
			Message: "请求参数不能为空",
		}, nil
	}

	ctx, span := tracing.StartSpan(ctx, "game.GetUserBetStats", trace.WithAttributes(
		attribute.String("method", "GetUserBetStats"),
		attribute.Int("site_id", int(req.SiteId)),
		attribute.Int64("user_id", req.UserId),
	))
	defer span.End()

	// 添加详细的参数日志
	g.Log().Infof(ctx, "GetUserBetStats 接收参数: site_id=%d, site_code=%s, user_id=%d, start=%s, end=%s, platform=%s",
		req.SiteId, req.SiteCode, req.UserId, req.StartTime, req.EndTime, req.Platform)

	// 使用getter方法获取值并打印
	g.Log().Infof(ctx, "GetUserBetStats 使用getter: site_id=%d, site_code=%s, user_id=%d, start=%s, end=%s, platform=%s",
		req.GetSiteId(), req.GetSiteCode(), req.GetUserId(), req.GetStartTime(), req.GetEndTime(), req.GetPlatform())

	// 参数验证
	if req.GetSiteId() <= 0 {
		return &v1.GetUserBetStatsRes{
			Success: false,
			Message: fmt.Sprintf("站点ID不能为空或小于等于0，接收到的值: %d", req.GetSiteId()),
		}, nil
	}
	if req.GetSiteCode() == "" {
		return &v1.GetUserBetStatsRes{
			Success: false,
			Message: fmt.Sprintf("站点代码不能为空，接收到的值: '%s'", req.GetSiteCode()),
		}, nil
	}
	if req.GetUserId() <= 0 {
		return &v1.GetUserBetStatsRes{
			Success: false,
			Message: fmt.Sprintf("用户ID不能为空或小于等于0，接收到的值: %d", req.GetUserId()),
		}, nil
	}
	if req.GetStartTime() == "" || req.GetEndTime() == "" {
		return &v1.GetUserBetStatsRes{
			Success: false,
			Message: fmt.Sprintf("开始时间和结束时间不能为空，接收到的值: start='%s', end='%s'", req.GetStartTime(), req.GetEndTime()),
		}, nil
	}

	// 获取 ClickHouse 连接
	chManager := registry.GetClickHouseManager()
	conn := chManager.GetConn()
	if conn == nil {
		return &v1.GetUserBetStatsRes{
			Success: false,
			Message: "ClickHouse 连接未初始化",
		}, nil
	}

	// 先查询总记录数（调试用）
	testQuery := "SELECT count() FROM game_analytics.bet_records WHERE site_id = ? AND site_code = ? AND user_id = ?"
	testRows, err := chManager.Query(ctx, testQuery, req.GetSiteId(), req.GetSiteCode(), req.GetUserId())
	if err != nil {
		g.Log().Errorf(ctx, "测试查询失败: %v", err)
	} else {
		var totalCount uint64
		if testRows.Next() {
			testRows.Scan(&totalCount)
			g.Log().Infof(ctx, "用户 %d 在 ClickHouse 中的总记录数: %d", req.GetUserId(), totalCount)
		}
		testRows.Close()
	}

	// 查询时间范围内的记录数（使用字符串比较，避免时区问题）
	testQuery2 := "SELECT count(), min(bet_time), max(bet_time) FROM game_analytics.bet_records WHERE site_id = ? AND site_code = ? AND user_id = ?"
	testRows2, err := chManager.Query(ctx, testQuery2, req.GetSiteId(), req.GetSiteCode(), req.GetUserId())
	if err != nil {
		g.Log().Errorf(ctx, "时间范围查询失败: %v", err)
	} else {
		var timeRangeCount uint64
		var minTime, maxTime time.Time
		if testRows2.Next() {
			err = testRows2.Scan(&timeRangeCount, &minTime, &maxTime)
			if err != nil {
				g.Log().Errorf(ctx, "Scan失败: %v", err)
			} else {
				g.Log().Infof(ctx, "用户 %d 的记录数: %d, 最早时间: %s, 最晚时间: %s",
					req.GetUserId(), timeRangeCount, minTime.Format("2006-01-02 15:04:05"), maxTime.Format("2006-01-02 15:04:05"))
			}
		}
		testRows2.Close()
	}

	// 构建查询 SQL
	query := `
		SELECT 
			user_id,
			anyLast(username) as username,
			count() as bet_count,
			toFloat64(sum(bet_amount)) as bet_amount,
			toFloat64(sum(win_amount)) as win_amount,
			toFloat64(sum(valid_amount)) as valid_amount,
			toFloat64(sum(net_amount)) as net_amount
		FROM game_analytics.bet_records
		WHERE site_id = ?
		  AND site_code = ?
		  AND user_id = ?
		  AND toString(toTimeZone(bet_time, 'Asia/Shanghai')) >= ?
		  AND toString(toTimeZone(bet_time, 'Asia/Shanghai')) <= ?
	`

	args := []interface{}{
		req.GetSiteId(),
		req.GetSiteCode(),
		req.GetUserId(),
		req.GetStartTime(),
		req.GetEndTime(),
	}

	// 如果指定了平台，添加平台过滤
	if req.GetPlatform() != "" {
		query += " AND platform = ?"
		args = append(args, req.GetPlatform())
	}

	query += " GROUP BY user_id"

	// 将时间转换为Asia/Shanghai时区后转为字符串比较
	g.Log().Infof(ctx, "执行查询，时区转换后字符串比较")
	g.Log().Infof(ctx, "查询参数: site_id=%v, site_code=%v, user_id=%v, start=%v, end=%v, platform=%v",
		args[0], args[1], args[2], args[3], args[4], req.GetPlatform())

	// 执行查询
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		g.Log().Errorf(ctx, "查询用户注单统计失败: %v", err)
		return &v1.GetUserBetStatsRes{
			Success: false,
			Message: fmt.Sprintf("查询失败: %v", err),
		}, nil
	}
	defer rows.Close()

	// 解析结果
	stats := &v1.UserBetStats{
		StartTime: req.GetStartTime(),
		EndTime:   req.GetEndTime(),
	}

	if rows.Next() {
		var userId uint64
		var betCount uint64
		var betAmount, winAmount, validAmount, netAmount float64
		err = rows.Scan(
			&userId,
			&stats.Username,
			&betCount,
			&betAmount,
			&winAmount,
			&validAmount,
			&netAmount,
		)
		if err != nil {
			g.Log().Errorf(ctx, "解析查询结果失败: %v", err)
			return &v1.GetUserBetStatsRes{
				Success: false,
				Message: fmt.Sprintf("解析结果失败: %v", err),
			}, nil
		}

		// 转换类型
		stats.UserId = int64(userId)
		stats.BetCount = int64(betCount)

		// 转换金额字段
		stats.BetAmount = fmt.Sprintf("%.2f", betAmount)
		stats.WinAmount = fmt.Sprintf("%.2f", winAmount)
		stats.ValidAmount = fmt.Sprintf("%.2f", validAmount)
		stats.NetAmount = fmt.Sprintf("%.2f", netAmount)
	} else {
		// 没有数据，返回空统计
		stats.UserId = req.GetUserId()
		stats.BetCount = 0
		stats.BetAmount = "0.00"
		stats.WinAmount = "0.00"
		stats.ValidAmount = "0.00"
		stats.NetAmount = "0.00"
	}

	// 查询按平台统计
	platformStats, err := s.queryPlatformStats(ctx, chManager, req)
	if err != nil {
		g.Log().Warningf(ctx, "查询平台统计失败: %v", err)
	} else {
		stats.PlatformStats = platformStats
	}

	g.Log().Infof(ctx, "GetUserBetStats 成功: user_id=%d, bet_count=%d, valid_amount=%s, net_amount=%s",
		stats.UserId, stats.BetCount, stats.ValidAmount, stats.NetAmount)

	return &v1.GetUserBetStatsRes{
		Success: true,
		Message: "查询成功",
		Stats:   stats,
	}, nil
}

// GetUserBetSummary 获取用户时间范围注单汇总（MySQL）
func (s *sGame) GetUserBetSummary(ctx context.Context, req *v1.GetUserBetSummaryReq) (*v1.GetUserBetSummaryRes, error) {
	siteId, _ := middleware.GetSiteIdFromContext(ctx)
	siteCode, _ := middleware.GetSiteCodeFromContext(ctx)
	userId, _ := middleware.GetUserIdFromContext(ctx)

	ctx, span := tracing.StartSpan(ctx, "game.GetUserBetSummary", trace.WithAttributes(
		attribute.String("method", "GetUserBetSummary"),
		attribute.Int("site_id", int(siteId)),
		attribute.String("site_code", siteCode),
		attribute.Int("user_id", int(userId)),
		attribute.String("start_time", req.GetStartTime()),
		attribute.String("end_time", req.GetEndTime()),
		attribute.String("platform", req.GetPlatform()),
	))
	defer span.End()

	summaryResult, err := dao.BetRecords.QueryUserBetSummary(
		ctx,
		siteId,
		siteCode,
		userId,
		req.GetStartTime(),
		req.GetEndTime(),
		req.GetPlatform(),
	)
	if err != nil {
		g.Log().Errorf(ctx, "查询用户注单汇总失败: %v", err)
		return &v1.GetUserBetSummaryRes{
			Success: false,
			Message: fmt.Sprintf("查询失败: %v", err),
		}, nil
	}

	summary := domain.BuildUserBetSummary(
		req.GetStartTime(),
		req.GetEndTime(),
		summaryResult.BetCount,
		summaryResult.BetAmount,
		summaryResult.WinAmount,
		summaryResult.ValidAmount,
		summaryResult.NetAmount,
	)

	return &v1.GetUserBetSummaryRes{
		Success: true,
		Message: "查询成功",
		Summary: summary,
	}, nil
}

// GetUserBetIncrementBatch 批量获取用户注单增量（MySQL，基于用户游标）
func (s *sGame) GetUserBetIncrementBatch(ctx context.Context, req *v1.GetUserBetIncrementBatchReq) (*v1.GetUserBetIncrementBatchRes, error) {
	if req == nil {
		return &v1.GetUserBetIncrementBatchRes{
			Success: false,
			Message: "请求参数不能为空",
		}, nil
	}

	siteCode := strings.TrimSpace(req.GetSiteCode())
	ctx, span := tracing.StartSpan(ctx, "game.GetUserBetIncrementBatch", trace.WithAttributes(
		attribute.String("method", "GetUserBetIncrementBatch"),
		attribute.Int("site_id", int(req.GetSiteId())),
		attribute.String("site_code", siteCode),
		attribute.Int("cursor_count", len(req.GetCursors())),
	))
	defer span.End()

	if req.GetSiteId() <= 0 {
		return &v1.GetUserBetIncrementBatchRes{
			Success: false,
			Message: "site_id 缺失或无效",
		}, nil
	}
	if siteCode == "" {
		return &v1.GetUserBetIncrementBatchRes{
			Success: false,
			Message: "site_code 缺失或无效",
		}, nil
	}
	if len(req.GetCursors()) == 0 {
		return &v1.GetUserBetIncrementBatchRes{
			Success: true,
			Message: "查询成功",
			List:    []*v1.UserBetIncrementItem{},
		}, nil
	}

	// 去重并保留每个用户最大的last_bet_id，避免重复用户游标影响统计。
	cursorMap := make(map[uint64]uint64, len(req.GetCursors()))
	for _, cursor := range req.GetCursors() {
		if cursor == nil || cursor.GetUserId() <= 0 {
			continue
		}
		userId := uint64(cursor.GetUserId())
		lastBetId := cursor.GetLastBetId()
		if exist, ok := cursorMap[userId]; !ok || lastBetId > exist {
			cursorMap[userId] = lastBetId
		}
	}
	if len(cursorMap) == 0 {
		return &v1.GetUserBetIncrementBatchRes{
			Success: true,
			Message: "查询成功",
			List:    []*v1.UserBetIncrementItem{},
		}, nil
	}

	cursorList := make([]dao.UserBetCursor, 0, len(cursorMap))
	for userId, lastBetId := range cursorMap {
		cursorList = append(cursorList, dao.UserBetCursor{
			UserId:    userId,
			LastBetId: lastBetId,
		})
	}

	const chunkSize = 500
	rows := make([]dao.UserBetIncrementResult, 0, len(cursorList))
	for i := 0; i < len(cursorList); i += chunkSize {
		end := i + chunkSize
		if end > len(cursorList) {
			end = len(cursorList)
		}
		part, err := dao.BetRecords.QueryUserBetIncrementBatch(
			ctx,
			uint(req.GetSiteId()),
			siteCode,
			cursorList[i:end],
		)
		if err != nil {
			g.Log().Errorf(ctx, "查询用户注单增量失败: %v", err)
			return &v1.GetUserBetIncrementBatchRes{
				Success: false,
				Message: fmt.Sprintf("查询失败: %v", err),
			}, nil
		}
		rows = append(rows, part...)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].UserId < rows[j].UserId
	})

	list := make([]*v1.UserBetIncrementItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, &v1.UserBetIncrementItem{
			UserId:        int64(row.UserId),
			ValidAmount:   fmt.Sprintf("%.4f", row.ValidAmount),
			MaxBetId:      row.MaxBetId,
			MaxSettleTime: row.MaxSettleTime,
			BetCount:      row.BetCount,
		})
	}

	return &v1.GetUserBetIncrementBatchRes{
		Success: true,
		Message: "查询成功",
		List:    list,
	}, nil
}

// queryPlatformStats 查询按平台统计
func (s *sGame) queryPlatformStats(ctx context.Context, chManager *registry.ClickHouseManager, req *v1.GetUserBetStatsReq) ([]*v1.PlatformStats, error) {
	query := `
		SELECT 
			platform,
			count() as bet_count,
			toFloat64(sum(bet_amount)) as bet_amount,
			toFloat64(sum(win_amount)) as win_amount,
			toFloat64(sum(valid_amount)) as valid_amount,
			toFloat64(sum(net_amount)) as net_amount
		FROM game_analytics.bet_records
		WHERE site_id = ?
		  AND site_code = ?
		  AND user_id = ?
		  AND toString(toTimeZone(bet_time, 'Asia/Shanghai')) >= ?
		  AND toString(toTimeZone(bet_time, 'Asia/Shanghai')) <= ?
	`

	// 注意：将ClickHouse的UTC时间转换为Asia/Shanghai时区后转为字符串比较

	args := []interface{}{
		req.GetSiteId(),
		req.GetSiteCode(),
		req.GetUserId(),
		req.GetStartTime(),
		req.GetEndTime(),
	}

	if req.GetPlatform() != "" {
		query += " AND platform = ?"
		args = append(args, req.GetPlatform())
	}

	query += " GROUP BY platform ORDER BY valid_amount DESC"

	// 使用 ClickHouseManager 的 Query 方法
	rows, err := chManager.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var platformStats []*v1.PlatformStats
	for rows.Next() {
		var platform string
		var betCount uint64
		var betAmount, winAmount, validAmount, netAmount float64

		err = rows.Scan(&platform, &betCount, &betAmount, &winAmount, &validAmount, &netAmount)
		if err != nil {
			return nil, err
		}

		platformStats = append(platformStats, &v1.PlatformStats{
			Platform:    platform,
			BetCount:    int64(betCount),
			BetAmount:   fmt.Sprintf("%.2f", betAmount),
			WinAmount:   fmt.Sprintf("%.2f", winAmount),
			ValidAmount: fmt.Sprintf("%.2f", validAmount),
			NetAmount:   fmt.Sprintf("%.2f", netAmount),
		})
	}

	return platformStats, nil
}

// GetAgentDailyBetStats 获取代理日统计（供结算）
func (s *sGame) GetAgentDailyBetStats(ctx context.Context, req *v1.GetAgentDailyBetStatsReq) (*v1.GetAgentDailyBetStatsRes, error) {
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return nil, fmt.Errorf("未获取站点ID")
	}
	if req.StartDate == "" || req.EndDate == "" {
		return nil, fmt.Errorf("开始日期和结束日期不能为空")
	}
	if req.StartDate > req.EndDate {
		return nil, fmt.Errorf("开始日期不能大于结束日期")
	}

	query := dao.BetRecords.Ctx(ctx).
		Where("site_id", siteId).
		Where("DATE(bet_time) >= ?", req.StartDate).
		Where("DATE(bet_time) <= ?", req.EndDate)
	if req.AgentId > 0 {
		query = query.Where("agent_id", req.AgentId)
	}

	var rows []struct {
		StatDate       string  `json:"stat_date"`
		AgentId        int64   `json:"agent_id"`
		BetAmount      float64 `json:"bet_amount"`
		ValidBetAmount float64 `json:"valid_bet_amount"`
		WinLoss        float64 `json:"win_loss"`
		MemberCount    int64   `json:"member_count"`
	}
	err := query.Fields(`
		DATE(bet_time) AS stat_date,
		agent_id,
		COALESCE(SUM(bet_amount),0) AS bet_amount,
		COALESCE(SUM(valid_amount),0) AS valid_bet_amount,
		COALESCE(SUM(net_amount),0) AS win_loss,
		COUNT(DISTINCT user_id) AS member_count
	`).Group("DATE(bet_time), agent_id").Order("stat_date ASC, agent_id ASC").Scan(&rows)
	if err != nil {
		return nil, err
	}

	res := &v1.GetAgentDailyBetStatsRes{List: make([]*v1.AgentDailyBetStatItem, 0, len(rows))}
	for _, row := range rows {
		res.List = append(res.List, &v1.AgentDailyBetStatItem{
			StatDate:       row.StatDate,
			AgentId:        row.AgentId,
			BetAmount:      row.BetAmount,
			ValidBetAmount: row.ValidBetAmount,
			WinLoss:        row.WinLoss,
			MemberCount:    row.MemberCount,
		})
	}
	return res, nil
}

// GetSyncTasks 获取同步任务列表
func (s *sGame) GetSyncTasks(ctx context.Context, req *v1.GetSyncTasksReq) (*v1.GetSyncTasksRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "game.GetSyncTasks", trace.WithAttributes(
		attribute.String("method", "GetSyncTasks"),
		attribute.Int("site_id", int(req.SiteId)),
		attribute.String("site_code", req.SiteCode),
		attribute.String("platform", req.Platform),
		attribute.Int("task_type", int(req.TaskType)),
		attribute.Int("status", int(req.Status)),
		attribute.Int("page", int(req.Page)),
		attribute.Int("size", int(req.Size)),
	))
	defer span.End()

	// 参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	if req.Size > 100 {
		req.Size = 100 // 限制最大页面大小
	}

	// 构建查询条件
	query := dao.SyncTasks.Ctx(ctx)

	// 站点ID过滤（0表示不过滤）
	if req.SiteId > 0 {
		query = query.Where("site_id", req.SiteId)
	}

	// 站点代码过滤
	if req.SiteCode != "" {
		query = query.Where("site_code", req.SiteCode)
	}

	// 平台过滤
	if req.Platform != "" {
		query = query.Where("platform", req.Platform)
	}

	// 任务类型过滤（0表示不过滤）
	if req.TaskType > 0 {
		query = query.Where("task_type", req.TaskType)
	}

	// 状态过滤（0表示不过滤）
	if req.Status > 0 {
		query = query.Where("status", req.Status)
	}

	// 获取总记录数
	total, err := query.Count()
	if err != nil {
		return nil, fmt.Errorf("获取任务总数失败: %v", err)
	}

	// 计算分页信息
	offset := (req.Page - 1) * req.Size

	// 获取分页数据 - 使用 Fields 指定字段并转换时间为字符串
	records, err := query.
		Fields(`
			id, site_id, site_code, platform, task_type,
			DATE_FORMAT(start_time, '%Y-%m-%d %H:%i:%s') as start_time,
			DATE_FORMAT(end_time, '%Y-%m-%d %H:%i:%s') as end_time,
			mysql_total, mysql_inserted, mysql_duplicated,
			mq_sent, mq_failed, mq_skipped,
			ch_synced, ch_failed,
			status, error_msg,
			DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
			DATE_FORMAT(updated_at, '%Y-%m-%d %H:%i:%s') as updated_at,
			DATE_FORMAT(completed_at, '%Y-%m-%d %H:%i:%s') as completed_at
		`).
		Order("id DESC").
		Limit(int(req.Size)).
		Offset(int(offset)).
		All()
	if err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %v", err)
	}

	// 转换为响应格式
	list := make([]*v1.SyncTask, 0, len(records))
	for _, record := range records {
		syncTask := &v1.SyncTask{
			Id:              record["id"].Int64(),
			SiteId:          record["site_id"].Int32(),
			SiteCode:        record["site_code"].String(),
			Platform:        record["platform"].String(),
			TaskType:        record["task_type"].Int32(),
			StartTime:       record["start_time"].String(),
			EndTime:         record["end_time"].String(),
			MysqlTotal:      record["mysql_total"].Int32(),
			MysqlInserted:   record["mysql_inserted"].Int32(),
			MysqlDuplicated: record["mysql_duplicated"].Int32(),
			MqSent:          record["mq_sent"].Int32(),
			MqFailed:        record["mq_failed"].Int32(),
			MqSkipped:       record["mq_skipped"].Int32(),
			ChSynced:        record["ch_synced"].Int32(),
			ChFailed:        record["ch_failed"].Int32(),
			Status:          record["status"].Int32(),
			ErrorMsg:        record["error_msg"].String(),
			CreatedAt:       record["created_at"].String(),
			UpdatedAt:       record["updated_at"].String(),
			CompletedAt:     record["completed_at"].String(),
		}
		list = append(list, syncTask)
	}

	return &v1.GetSyncTasksRes{
		List:  list,
		Total: int32(total),
		Page:  req.Page,
		Size:  req.Size,
	}, nil
}
