package cmd

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"jh_game_service/internal/logic/frontend/game/platform"
	"jh_game_service/internal/model/entity"
	"jh_game_service/internal/registry"
)

var (
	SyncGameSource = gcmd.Command{
		Name:  "sync",
		Usage: "sync <PLATFORM>",
		Brief: "同步游戏平台的游戏列表和图片资源",
		Arguments: []gcmd.Argument{
			{
				Name:  "platform",
				Short: "p",
				Brief: "游戏平台: ONE, IG_JILI",
				IsArg: true,
			},
		},
		Description: `
同步游戏平台的游戏列表和图片资源到数据库

示例:
  ./jh_game_service sync ONE
`,
		Strict: true,
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 1. 获取平台参数（GetArg索引：0=程序名, 1=sync, 2=ONE）
			platformName := parser.GetArg(2).String()
			if platformName == "" {
				return fmt.Errorf("请指定游戏平台，例如: ./jh_game_service sync ONE")
			}

			g.Log().Infof(ctx, "开始同步游戏平台: %s", platformName)

			// 2. 初始化Consul
			fmt.Println("初始化Consul客户端...")
			consulClient := registry.InitConsul()

			// 3. 初始化数据库配置管理器
			fmt.Println("初始化数据库配置管理器...")
			dbManager := registry.GetDBConfigManager()
			if err := dbManager.Init(consulClient); err != nil {
				return fmt.Errorf("初始化数据库配置管理器失败: %v", err)
			}
			defer dbManager.Stop()

			// 4. 查询游戏配置（使用默认站点）
			siteId := 1 // 默认站点ID
			siteCode := "jh"

			var siteGame *entity.SiteGame
			err = g.DB().Model("site_game").
				Where("site_id", siteId).
				Where("platform", platformName).
				Where("status", 1).
				Scan(&siteGame)

			if err != nil {
				return fmt.Errorf("查询游戏配置失败: %v", err)
			}

			if siteGame == nil {
				return fmt.Errorf("未找到平台 %s 的游戏配置", platformName)
			}

			g.Log().Infof(ctx, "游戏配置: GameId=%d, Platform=%s, ApiUrl=%s",
				siteGame.GameId, siteGame.Platform, siteGame.ApiUrl)

			// 3. 创建平台实例
			gamePlatform, err := platform.GetFactory().Create(platformName)
			if err != nil {
				return fmt.Errorf("创建游戏平台实例失败: %v", err)
			}

			// 4. 执行同步
			params := &platform.GetGameSourceParams{
				SiteId:   siteId,
				SiteCode: siteCode,
				GameId:   siteGame.GameId,
				SiteGame: *siteGame,
			}

			err = gamePlatform.GetGameSource(ctx, params)
			if err != nil {
				return fmt.Errorf("同步游戏资源失败: %v", err)
			}

			g.Log().Infof(ctx, "游戏资源同步完成")
			return nil
		},
	}
)
