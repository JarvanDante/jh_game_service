package game

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc/metadata"
	v1 "jh_game_service/api/frontend/game/v1"
	"jh_game_service/internal/dao"
	"jh_game_service/internal/logic/frontend/game/platform"
	_ "jh_game_service/internal/logic/frontend/game/platform/one" // 自动注册ONE平台
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
