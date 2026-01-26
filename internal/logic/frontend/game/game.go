package game

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc/metadata"
	v1 "jh_game_service/api/frontend/game/v1"
	"jh_game_service/internal/dao"
	"jh_game_service/internal/logic/frontend/game/platform"
	_ "jh_game_service/internal/logic/frontend/game/platform/ig_jili" // 自动注册IG_JILI平台
	_ "jh_game_service/internal/logic/frontend/game/platform/one"     // 自动注册ONE平台
	"jh_game_service/internal/middleware"
	"jh_game_service/internal/model/do"
	"jh_game_service/internal/model/entity"
	"jh_game_service/internal/service/frontend"
	"strings"
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
