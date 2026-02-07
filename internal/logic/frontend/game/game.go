package game

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc/metadata"
	v1 "jh_game_service/api/frontend/game/v1"
	balanceV1 "jh_game_service/api/grpc/balance/v1"
	"jh_game_service/internal/dao"
	"jh_game_service/internal/domain"
	grpc_client "jh_game_service/internal/grpc_client"
	"jh_game_service/internal/logic/frontend/game/platform"
	_ "jh_game_service/internal/logic/frontend/game/platform/ig_jili" // 自动注册IG_JILI平台
	_ "jh_game_service/internal/logic/frontend/game/platform/one"     // 自动注册ONE平台
	"jh_game_service/internal/middleware"
	"jh_game_service/internal/model/do"
	"jh_game_service/internal/model/entity"
	"jh_game_service/internal/service/frontend"
	"strings"
	"time"
)

type (
	sGame struct{}
)

func init() {
	frontend.RegisterGame(&sGame{})
	// 注入获取游戏配置的函数，避免循环导入
	platform.GetGameConfigFunc = GetGameConfigById
}

//RedirectToGame
/**
 * @desc：进入游戏
 * @param ctx
 * @param req
 * @return *v1.RedirectToGameRes
 * @return error
 * @author : Carson
 */
func (s *sGame) RedirectToGame(ctx context.Context, req *v1.RedirectToGameReq) (*v1.RedirectToGameRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	// 1. 获取站点和用户信息
	siteCode, ok := middleware.GetSiteCodeFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("获取站点信息失败")
	}

	userId, siteId, err := middleware.GetUserInfoFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败")
	}

	username := middleware.GetUsernameFromContext(ctx)
	if username == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}

	currency := middleware.GetCurrencyFromContext(ctx)
	if currency == "" {
		currency = "MMK(K)"
	}

	domain := middleware.GetDomainFromContext(ctx)
	if domain == "" {
		domain = "https://example.com"
	}
	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "https://" + domain
	}

	clientIp := middleware.GetClientIPFromContext(ctx)

	// 2. 查询游戏信息
	var siteGame entity.SiteGame
	err = dao.SiteGame.Ctx(ctx).
		Where(do.SiteGame{
			SiteId: int(siteId),
			GameId: int(req.GameId),
		}).
		Scan(&siteGame)
	if err != nil {
		return nil, fmt.Errorf("查询游戏信息失败: %v", err)
	}
	if siteGame.Id == 0 {
		return nil, fmt.Errorf("游戏不存在")
	}

	// 3. 检查游戏状态
	if siteGame.Status != 1 {
		return nil, fmt.Errorf("%s游戏关闭维护", siteGame.Name)
	}

	// 4. 通过工厂创建游戏平台实例
	gamePlatform, err := platform.GetFactory().Create(siteGame.Platform)
	if err != nil {
		return nil, err
	}

	// 5. 调用平台的GetGameUrl方法
	resultUrl, err := gamePlatform.GetGameUrl(ctx, &platform.GameParams{
		SiteId:   siteId,
		UserId:   userId,
		Username: username,
		SiteCode: siteCode,
		Currency: currency,
		Domain:   domain,
		ClientIP: clientIp,
		SiteGame: siteGame,
		Request:  req,
	})

	if err != nil {
		return nil, fmt.Errorf("获取游戏链接失败: %v", err)
	}

	// 6. 返回结果
	gameType, _ := resultUrl["type"].(string)
	gameUrl, _ := resultUrl["url"].(string)

	return &v1.RedirectToGameRes{
		Type: gameType,
		Url:  gameUrl,
	}, nil
}

//getClientIPFromContext
/**
 * @desc：从上下文获取客户端IP
 * @param ctx
 * @return string
 * @author : Carson
 */
func (s *sGame) getClientIPFromContext(ctx context.Context) string {
	// 尝试从gRPC metadata获取
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ips := md.Get("client_ip"); len(ips) > 0 {
			return ips[0]
		}
		if ips := md.Get("x-real-ip"); len(ips) > 0 {
			return ips[0]
		}
		if ips := md.Get("x-forwarded-for"); len(ips) > 0 {
			return ips[0]
		}
	}

	// 尝试从context.Value获取
	if ip := ctx.Value("client_ip"); ip != nil {
		if ipStr, ok := ip.(string); ok {
			return ipStr
		}
	}

	return ""
}

// checkIPRestriction 检查IP限制
// TODO: 实现IP限制检查逻辑
func (s *sGame) checkIPRestriction(ctx context.Context, gameId int32, key string, clientIp string) error {
	// 1. 通过game_id和key查询GameList,获取vendor
	// 2. 通过platform_id和vendor查询game_restrict_country表
	// 3. 如果找到限制配置且country_list不为空
	// 4. 通过IP获取国家代码
	// 5. 检查IP所属国家是否在黑名单中
	// 6. 如果在黑名单中，返回错误

	g.Log().Debugf(ctx, "IP限制检查功能待实现: clientIp=%s", clientIp)
	return nil
}

//GetGameCategories
/**
 * @desc：获取游戏种类
 * @param ctx
 * @param req
 * @return *v1.GetGameCategoriesRes
 * @return error
 * @author : Carson
 */
func (s *sGame) GetGameCategories(ctx context.Context, req *v1.GetGameCategoriesReq) (*v1.GetGameCategoriesRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	// 从上下文获取站点信息（不需要用户登录）
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return nil, fmt.Errorf("获取站点信息失败")
	}

	// 构建查询条件
	query := dao.SiteGame.Ctx(ctx).
		Where(do.SiteGame{
			SiteId: int(siteId),
			Status: 1, // 只查询状态为1的游戏
		})

	// 如果指定了type，添加type过滤
	if req.Type > 0 {
		query = query.Where(do.SiteGame{
			Type: int(req.Type),
		})
	}

	// 按sort降序排序
	query = query.Order("sort DESC")

	// 查询游戏列表
	var siteGames []entity.SiteGame
	err := query.Scan(&siteGames)
	if err != nil {
		return nil, fmt.Errorf("查询游戏种类失败: %v", err)
	}

	// 转换为响应格式
	categories := make([]*v1.GameCategory, 0, len(siteGames))
	for _, game := range siteGames {
		categories = append(categories, &v1.GameCategory{
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

	return &v1.GetGameCategoriesRes{
		Categories: categories,
	}, nil
}

//GetGameList
/**
 * @desc：获取游戏列表
 * @param ctx
 * @param req
 * @return *v1.GetGameListRes
 * @return error
 * @author : Carson
 */
func (s *sGame) GetGameList(ctx context.Context, req *v1.GetGameListReq) (*v1.GetGameListRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	// 从上下文获取站点信息（不需要用户登录）
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return nil, fmt.Errorf("获取站点信息失败")
	}

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
	query := dao.GameList.Ctx(ctx).Where("site_id = ?", siteId)

	// 平台过滤
	if req.Platform != "" {
		query = query.Where("platform = ?", req.Platform)
	}

	// 热门过滤
	if req.Hot > 0 {
		query = query.Where("hot = ?", req.Hot)
	}

	// 新游戏过滤
	if req.IsNew > 0 {
		query = query.Where("is_new = ?", req.IsNew)
	}

	// 推荐过滤
	if req.Recommend > 0 {
		query = query.Where("recommend = ?", req.Recommend)
	}

	// 固定过滤状态为1（启用）
	query = query.Where("status = ?", 1)

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
	list := make([]*v1.GameItem, 0, len(gameList))
	for _, game := range gameList {
		gameItem := &v1.GameItem{
			Id:        int32(game.Id),
			SiteId:    int32(game.SiteId),
			Platform:  game.Platform,
			Vendor:    game.Vendor,
			Code:      game.Code,
			Name:      game.Name,
			Style:     game.Style,
			Img:       game.Img,
			Img1:      game.Img1,
			Hot:       int32(game.Hot),
			Recommend: int32(game.Recommend),
			IsNew:     int32(game.IsNew),
			Status:    int32(game.Status),
			Sort:      int32(game.Sort),
			GameId:    0, // 默认值
		}

		// 通过 platform 和 site_id 查询 site_game 表获取 game_id
		var siteGame entity.SiteGame
		err = dao.SiteGame.Ctx(ctx).
			Where(do.SiteGame{
				SiteId:   int(siteId),
				Platform: game.Platform,
			}).
			Scan(&siteGame)
		if err == nil && siteGame.Id > 0 {
			gameItem.GameId = int32(siteGame.GameId)
		}

		list = append(list, gameItem)
	}

	return &v1.GetGameListRes{
		List:  list,
		Total: int32(total),
	}, nil
}

//GetGameBalance
/**
 * @desc：获取游戏余额
 * @param ctx
 * @param req
 * @return *v1.GetGameBalanceRes
 * @return error
 * @author : Carson
 */
func (s *sGame) GetGameBalance(ctx context.Context, req *v1.GetGameBalanceReq) (*v1.GetGameBalanceRes, error) {
	if req == nil || req.GameId <= 0 {
		return &v1.GetGameBalanceRes{
			Balance: 0,
			Success: false,
			Message: "游戏ID无效",
		}, nil
	}

	// 1. 获取站点和用户信息
	siteCode, ok := middleware.GetSiteCodeFromContext(ctx)
	if !ok {
		return &v1.GetGameBalanceRes{
			Balance: 0,
			Success: false,
			Message: "获取站点信息失败",
		}, nil
	}

	userId, siteId, err := middleware.GetUserInfoFromContext(ctx)
	if err != nil {
		return &v1.GetGameBalanceRes{
			Balance: 0,
			Success: false,
			Message: "获取用户信息失败",
		}, nil
	}

	username := middleware.GetUsernameFromContext(ctx)
	if username == "" {
		return &v1.GetGameBalanceRes{
			Balance: 0,
			Success: false,
			Message: "用户名不能为空",
		}, nil
	}

	currency := middleware.GetCurrencyFromContext(ctx)
	if currency == "" {
		currency = "MMK(K)"
	}

	g.Log().Infof(ctx, "查询游戏余额: site_id=%d, user_id=%d, username=%s, game_id=%d",
		siteId, userId, username, req.GameId)

	// 2. 查询游戏配置
	gameConfig, err := GetGameConfigById(ctx, siteId, int(req.GameId))
	if err != nil {
		g.Log().Errorf(ctx, "查询游戏配置失败: %v", err)
		return &v1.GetGameBalanceRes{
			Balance: 0,
			Success: false,
			Message: fmt.Sprintf("查询游戏配置失败: %v", err),
		}, nil
	}

	g.Log().Infof(ctx, "游戏配置: GameId=%d, Platform=%s, Name=%s",
		gameConfig.GameId, gameConfig.Platform, gameConfig.Name)

	// 3. 使用工厂创建游戏平台实例
	gamePlatform, err := platform.GetFactory().Create(gameConfig.Platform)
	if err != nil {
		g.Log().Errorf(ctx, "创建游戏平台实例失败: %v", err)
		return &v1.GetGameBalanceRes{
			Balance: 0,
			Success: false,
			Message: fmt.Sprintf("创建游戏平台实例失败: %v", err),
		}, nil
	}

	g.Log().Infof(ctx, "成功创建游戏平台实例: %s", gameConfig.Platform)

	// 4. 构建游戏参数
	params := &platform.GameParams{
		SiteId:   siteId,
		UserId:   userId,
		Username: username,
		SiteCode: siteCode,
		Currency: currency,
		SiteGame: *gameConfig,
	}

	g.Log().Infof(ctx, "游戏参数: GameId=%d, ApiUrl=%s", params.SiteGame.GameId, params.SiteGame.ApiUrl)

	// 5. 查询游戏余额
	balance, err := gamePlatform.GetBalance(ctx, params)
	if err != nil {
		g.Log().Errorf(ctx, "查询游戏余额失败: %v", err)
		return &v1.GetGameBalanceRes{
			Balance: 0,
			Success: false,
			Message: fmt.Sprintf("查询游戏余额失败: %v", err),
		}, nil
	}

	g.Log().Infof(ctx, "查询游戏余额成功: game_id=%d, balance=%.4f", req.GameId, balance)

	return &v1.GetGameBalanceRes{
		Balance: balance,
		Success: true,
		Message: "查询成功",
	}, nil
}

// RecallGameBalance 一键回收游戏余额
/**
 * @desc：一键回收三方游戏余额到中心钱包
 * @param ctx
 * @param req
 * @return *v1.RecallGameBalanceRes
 * @return error
 * @author : AI Assistant
 */
func (s *sGame) RecallGameBalance(ctx context.Context, req *v1.RecallGameBalanceReq) (*v1.RecallGameBalanceRes, error) {
	// 1. 获取站点和用户信息
	siteCode, ok := middleware.GetSiteCodeFromContext(ctx)
	if !ok {
		return &v1.RecallGameBalanceRes{
			Success: false,
			Message: "获取站点信息失败",
		}, nil
	}

	userId, siteId, err := middleware.GetUserInfoFromContext(ctx)
	if err != nil {
		return &v1.RecallGameBalanceRes{
			Success: false,
			Message: "获取用户信息失败",
		}, nil
	}

	username := middleware.GetUsernameFromContext(ctx)
	if username == "" {
		return &v1.RecallGameBalanceRes{
			Success: false,
			Message: "用户名不能为空",
		}, nil
	}

	currency := middleware.GetCurrencyFromContext(ctx)
	if currency == "" {
		currency = "MMK(K)"
	}

	g.Log().Infof(ctx, "一键回收游戏余额: site_id=%d, user_id=%d, username=%s", siteId, userId, username)

	// 2. 调用 balance 服务获取 lastGameId
	balanceClient, conn, err := grpc_client.GetBalanceClient(ctx)
	if err != nil {
		g.Log().Errorf(ctx, "连接balance服务失败: %v", err)
		return &v1.RecallGameBalanceRes{
			Success: false,
			Message: "连接余额服务失败",
		}, nil
	}
	defer conn.Close()

	// 添加 metadata
	md := metadata.New(map[string]string{
		"site_id":   fmt.Sprintf("%d", siteId),
		"user_id":   fmt.Sprintf("%d", userId),
		"username":  username,
		"site_code": siteCode,
		"currency":  currency,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// 获取最后进入的游戏ID
	lastGameRes, err := balanceClient.GetLastBalanceLogGame(ctx, &balanceV1.GetLastBalanceLogGameReq{
		SiteId: int32(siteId),
		UserId: int32(userId),
	})

	if err != nil {
		g.Log().Errorf(ctx, "查询最后游戏ID失败: %v", err)
		return &v1.RecallGameBalanceRes{
			Success: false,
			Message: "查询游戏状态失败",
		}, nil
	}

	lastGameId := int(lastGameRes.GameId)
	g.Log().Infof(ctx, "最后游戏ID: %d", lastGameId)

	// 3. 如果 lastGameId <= 0，说明用户没有在游戏中
	if lastGameId <= 0 {
		g.Log().Infof(ctx, "用户未在游戏中，无需回收")
		return &v1.RecallGameBalanceRes{
			Success:        true,
			Message:        "无需回收，您当前未在游戏中",
			RecalledAmount: 0,
			GameId:         0,
			GameName:       "",
		}, nil
	}

	// 4. 查询游戏配置
	gameConfig, err := GetGameConfigById(ctx, siteId, lastGameId)
	if err != nil {
		g.Log().Errorf(ctx, "查询游戏配置失败: %v", err)
		return &v1.RecallGameBalanceRes{
			Success: false,
			Message: "查询游戏配置失败",
		}, nil
	}

	g.Log().Infof(ctx, "游戏配置: GameId=%d, Platform=%s, Name=%s", gameConfig.GameId, gameConfig.Platform, gameConfig.Name)

	// 5. 创建游戏平台实例
	gamePlatform, err := platform.GetFactory().Create(gameConfig.Platform)
	if err != nil {
		g.Log().Errorf(ctx, "创建游戏平台实例失败: %v", err)
		return &v1.RecallGameBalanceRes{
			Success: false,
			Message: "创建游戏平台实例失败",
		}, nil
	}

	// 6. 构建游戏参数
	params := &platform.GameParams{
		SiteId:   siteId,
		UserId:   userId,
		Username: username,
		SiteCode: siteCode,
		Currency: currency,
		SiteGame: *gameConfig,
	}

	// 7. 查询游戏余额
	gameBalance, err := gamePlatform.GetBalance(ctx, params)
	if err != nil {
		g.Log().Errorf(ctx, "查询游戏余额失败: %v", err)
		return &v1.RecallGameBalanceRes{
			Success: false,
			Message: "查询游戏余额失败",
		}, nil
	}

	g.Log().Infof(ctx, "游戏余额: %.2f", gameBalance)

	// 8. 如果游戏余额 <= 0 或者太小（< 0.1），无需回收
	if gameBalance <= 0 || gameBalance < 0.1 {
		g.Log().Infof(ctx, "游戏余额太小(%.4f)，无需回收", gameBalance)
		return &v1.RecallGameBalanceRes{
			Success:        true,
			Message:        "回收成功",
			RecalledAmount: 0,
			GameId:         int32(lastGameId),
			GameName:       gameConfig.Name,
		}, nil
	}

	// 9. 从游戏转出余额到中心钱包
	transferParams := &platform.TransferParams{
		SiteId:   siteId,
		UserId:   userId,
		Username: username,
		SiteCode: siteCode,
		Currency: currency,
		Amount:   gameBalance,
		SiteGame: *gameConfig,
	}

	err = gamePlatform.TransferOut(ctx, transferParams)
	if err != nil {
		g.Log().Errorf(ctx, "转出游戏余额失败: %v", err)
		return &v1.RecallGameBalanceRes{
			Success: false,
			Message: fmt.Sprintf("回收失败: %v", err),
		}, nil
	}

	g.Log().Infof(ctx, "一键回收成功: game_id=%d, game_name=%s, amount=%.3f", lastGameId, gameConfig.Name, gameBalance)

	return &v1.RecallGameBalanceRes{
		Success:        true,
		Message:        "回收成功",
		RecalledAmount: gameBalance,
		GameId:         int32(lastGameId),
		GameName:       gameConfig.Name,
	}, nil
}

// GetUserBetHistory 获取用户历史投注列表（时间倒序）
func (s *sGame) GetUserBetHistory(ctx context.Context, req *v1.GetUserBetHistoryReq) (*v1.GetUserBetHistoryRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if req.GetStartTime() == "" || req.GetEndTime() == "" {
		return nil, fmt.Errorf("开始时间和结束时间不能为空")
	}

	siteCode, ok := middleware.GetSiteCodeFromContext(ctx)
	if !ok || siteCode == "" {
		return nil, fmt.Errorf("获取站点信息失败")
	}

	userId, siteId, err := middleware.GetUserInfoFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败")
	}

	list, total, err := dao.BetRecords.QueryUserBetHistory(
		ctx,
		uint(siteId),
		siteCode,
		userId,
		req.GetStartTime(),
		req.GetEndTime(),
		int(req.GetPage()),
		int(req.GetSize()),
	)
	if err != nil {
		return nil, fmt.Errorf("查询历史投注失败: %v", err)
	}

	return domain.BuildUserBetHistoryRes(req.GetStartTime(), req.GetEndTime(), list, total), nil
}

// GetWashCodeQuota 获取当前用户可洗码额度
func (s *sGame) GetWashCodeQuota(ctx context.Context, req *v1.GetWashCodeQuotaReq) (*v1.GetWashCodeQuotaRes, error) {
	if req == nil {
		req = &v1.GetWashCodeQuotaReq{}
	}

	siteCode, ok := middleware.GetSiteCodeFromContext(ctx)
	if !ok || siteCode == "" {
		return &v1.GetWashCodeQuotaRes{
			Success: false,
			Message: "获取站点信息失败",
		}, nil
	}

	userId, siteId, err := middleware.GetUserInfoFromContext(ctx)
	if err != nil {
		return &v1.GetWashCodeQuotaRes{
			Success: false,
			Message: "获取用户信息失败",
		}, nil
	}

	lastBetId := req.GetLastBetId()
	lastRebateTime := ""
	if lastBetId == 0 {
		lastRebateTime, err = s.getLastRebateTimeFromBalance(ctx, siteId, userId, siteCode)
		if err != nil {
			g.Log().Errorf(ctx, "获取最近洗码时间失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
			return &v1.GetWashCodeQuotaRes{
				Success: false,
				Message: "获取最近洗码时间失败",
			}, nil
		}
	}

	quota, err := dao.BetRecords.QueryUserWashCodeQuota(
		ctx,
		uint(siteId),
		siteCode,
		uint64(userId),
		lastBetId,
		lastRebateTime,
	)
	if err != nil {
		g.Log().Errorf(ctx, "查询可洗码额度失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		return &v1.GetWashCodeQuotaRes{
			Success: false,
			Message: "查询可洗码额度失败",
		}, nil
	}

	return &v1.GetWashCodeQuotaRes{
		Success:        true,
		Message:        "查询成功",
		ValidBetAmount: quota.ValidAmount,
		MaxBetId:       quota.MaxBetId,
		LastRebateTime: lastRebateTime,
	}, nil
}

func (s *sGame) getLastRebateTimeFromBalance(ctx context.Context, siteId int, userId uint, siteCode string) (string, error) {
	balanceClient, conn, err := grpc_client.GetBalanceClient(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	mdMap := map[string]string{
		"site_id": fmt.Sprintf("%d", siteId),
		"user_id": fmt.Sprintf("%d", userId),
	}
	if siteCode != "" {
		mdMap["site_code"] = siteCode
	}
	if username := middleware.GetUsernameFromContext(ctx); username != "" {
		mdMap["username"] = username
	}
	if currency := middleware.GetCurrencyFromContext(ctx); currency != "" {
		mdMap["currency"] = currency
	}

	ctxWithMeta := metadata.NewOutgoingContext(ctx, metadata.New(mdMap))
	res, err := balanceClient.GetBalanceLog(ctxWithMeta, &balanceV1.GetBalanceLogReq{
		Start:  "2000-01-01 00:00:00",
		End:    time.Now().Format("2006-01-02 15:04:05"),
		Type:   4, // 前台类型: 4=返水
		Status: 3, // 成功
		Page:   1,
		Size:   1,
	})
	if err != nil {
		return "", err
	}
	if res == nil || len(res.List) == 0 {
		return "", nil
	}

	return strings.TrimSpace(res.List[0].GetTime()), nil
}
