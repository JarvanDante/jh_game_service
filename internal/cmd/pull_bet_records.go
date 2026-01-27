package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"jh_game_service/internal/logic/frontend/game/platform"
	"jh_game_service/internal/model/entity"
	"jh_game_service/internal/mq"
	"jh_game_service/internal/registry"
)

var (
	// PullBetRecords 拉取注单命令
	PullBetRecords = gcmd.Command{
		Name:  "pull",
		Usage: "pull <platform> [startTime] [endTime]",
		Brief: "拉取指定平台的注单数据",
		Arguments: []gcmd.Argument{
			{
				Name:  "platform",
				Short: "p",
				Brief: "游戏平台: ONE, IG_JILI, all",
				IsArg: true,
			},
			{
				Name:  "startTime",
				Short: "s",
				Brief: "开始时间（可选），格式: 2006-01-02 15:04:05",
				IsArg: true,
			},
			{
				Name:  "endTime",
				Short: "e",
				Brief: "结束时间（可选），格式: 2006-01-02 15:04:05",
				IsArg: true,
			},
		},
		Description: `
拉取指定游戏平台的注单数据，保存到MySQL并同步到ClickHouse

支持的平台:
  - ONE       ONE游戏平台
  - IG_JILI   IG_JILI游戏平台
  - all       所有平台

示例:
  ./jh_game_service pull ONE                                    # 拉取最近10分钟注单
  ./jh_game_service pull ONE "2026-01-24 00:00:00" "2026-01-24 23:59:59"  # 补单指定时间范围
  ./jh_game_service pull all                                    # 拉取所有平台
`,
		Strict: true,
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 获取平台参数（GetArg索引：0=程序名, 1=pull, 2=ONE）
			platformName := parser.GetArg(2).String()
			if platformName == "" {
				fmt.Println("错误: 请指定要拉取的平台")
				fmt.Println("用法: ./jh_game_service pull <platform> [startTime] [endTime]")
				fmt.Println("支持的平台: ONE, IG_JILI, all")
				return fmt.Errorf("缺少平台参数")
			}

			// 获取可选的时间参数
			startTime := parser.GetArg(3).String()
			endTime := parser.GetArg(4).String()

			fmt.Printf("=== 开始拉取注单 ===\n")
			fmt.Printf("平台: %s\n", platformName)
			if startTime != "" && endTime != "" {
				fmt.Printf("时间范围: %s ~ %s (补单模式)\n", startTime, endTime)
			} else {
				fmt.Printf("时间: %s (正常拉单)\n", time.Now().Format("2006-01-02 15:04:05"))
			}
			fmt.Println("====================")

			// 初始化Consul
			fmt.Println("初始化Consul客户端...")
			consulClient := registry.InitConsul()

			// 初始化数据库配置管理器
			fmt.Println("初始化数据库配置管理器...")
			dbManager := registry.GetDBConfigManager()
			if err := dbManager.Init(consulClient); err != nil {
				return fmt.Errorf("初始化数据库配置管理器失败: %v", err)
			}
			defer dbManager.Stop()

			// 初始化 RabbitMQ（用于发送同步消息）
			fmt.Println("初始化RabbitMQ连接...")
			if err := mq.InitRabbitMQ(ctx); err != nil {
				fmt.Printf("RabbitMQ连接初始化失败: %v（将跳过ClickHouse同步）\n", err)
			} else {
				fmt.Println("RabbitMQ连接初始化成功")
				defer mq.CloseRabbitMQ()
			}

			// 初始化ClickHouse
			fmt.Println("初始化ClickHouse连接...")
			chHost := g.Cfg().MustGet(ctx, "clickhouse.host").String()
			chPort := g.Cfg().MustGet(ctx, "clickhouse.port").Int()
			chDatabase := g.Cfg().MustGet(ctx, "clickhouse.database").String()
			chUsername := g.Cfg().MustGet(ctx, "clickhouse.username").String()
			chPassword := g.Cfg().MustGet(ctx, "clickhouse.password").String()

			chManager := registry.GetClickHouseManager()
			if err := chManager.Init(chHost, chPort, chDatabase, chUsername, chPassword); err != nil {
				fmt.Printf("ClickHouse连接初始化失败: %v\n", err)
				// 不中断，继续保存MySQL
			} else {
				fmt.Println("ClickHouse连接初始化成功")
				defer chManager.Close()
			}

			// 执行拉取
			if platformName == "all" {
				return pullAllPlatforms(ctx, startTime, endTime)
			}
			return pullPlatform(ctx, platformName, startTime, endTime)
		},
	}
)

//pullAllPlatforms
/**
 * @desc：拉取所有平台(获取所有平台)
 * @param ctx
 * @param startTime
 * @param endTime
 * @return error
 * @author : Carson
 */
func pullAllPlatforms(ctx context.Context, startTime, endTime string) error {
	platforms := []string{"ONE", "IG_JILI"}

	for _, p := range platforms {
		fmt.Printf("\n--- 拉取平台: %s ---\n", p)
		if err := pullPlatform(ctx, p, startTime, endTime); err != nil {
			fmt.Printf("拉取平台 %s 失败: %v\n", p, err)
			continue
		}
	}

	fmt.Println("\n=== 所有平台拉取完成 ===")
	return nil
}

//pullPlatform
/**
 * @desc：拉取指定平台(单个)
 * @param ctx
 * @param platformName
 * @param startTime
 * @param endTime
 * @return error
 * @author : Carson
 */
func pullPlatform(ctx context.Context, platformName, startTime, endTime string) error {
	fmt.Printf("\n平台==> %s \n", platformName)
	dbManager := registry.GetDBConfigManager()
	siteCodes := dbManager.GetAllSiteCodes()

	fmt.Printf("站点数量: %d\n", len(siteCodes))

	totalRecords := 0
	for _, siteCode := range siteCodes {
		count, err := pullSitePlatform(ctx, siteCode, platformName, startTime, endTime)
		if err != nil {
			fmt.Printf("拉取站点 %s 失败: %v\n", siteCode, err)
			continue
		}
		totalRecords += count
	}

	fmt.Printf("\n平台 %s 拉取完成，共 %d 条记录\n", platformName, totalRecords)
	return nil
}

//pullSitePlatform
/**
 * @desc：拉取单个站点的指定平台
 * @param ctx
 * @param siteCode
 * @param platformName
 * @param startTime
 * @param endTime
 * @return int
 * @return error
 * @author : Carson
 */
func pullSitePlatform(ctx context.Context, siteCode string, platformName, startTime, endTime string) (int, error) {
	// 获取站点配置
	dbManager := registry.GetDBConfigManager()
	siteConfig, err := dbManager.GetSiteConfig(siteCode)

	if err != nil {
		return 0, fmt.Errorf("获取站点配置失败: %v", err)
	}
	totalCount := 0
	// 遍历站点的游戏，找到匹配的平台
	for _, gameConfig := range siteConfig.Game {
		if gameConfig.Platform != platformName {
			continue
		}

		fmt.Printf("\n站点: %s, 游戏: %s\n", siteCode, gameConfig.Name)

		count, err := pullGameBetRecords(ctx, siteCode, siteConfig.Base.SiteID, &gameConfig, startTime, endTime)
		if err != nil {
			fmt.Printf("  拉取失败: %v\n", err)
			continue
		}

		totalCount += count
		fmt.Printf("  拉取成功: %d 条记录\n", count)
	}

	return totalCount, nil
}

//pullGameBetRecords
/**
 * @desc：拉取单个游戏的注单
 * @param ctx
 * @param siteCode
 * @param siteId
 * @param gameConfig
 * @param startTime
 * @param endTime
 * @return int
 * @return error
 * @author : Carson
 */
func pullGameBetRecords(ctx context.Context, siteCode string, siteId int, gameConfig *registry.GameConfig, startTime, endTime string) (int, error) {
	// 创建游戏平台实例
	gamePlatform, err := platform.GetFactory().Create(gameConfig.Platform)
	if err != nil {
		return 0, fmt.Errorf("创建游戏平台实例失败: %v", err)
	}

	// 如果指定了时间范围，使用指定的时间（补单模式）
	if startTime != "" && endTime != "" {
		fmt.Printf("  补单模式: %s ~ %s\n", startTime, endTime)

		// 构建拉单参数
		params := &platform.PullRecordsParams{
			SiteId:   siteId,
			SiteCode: siteCode,
			SiteGame: entity.SiteGame{
				GameId:   gameConfig.GameId,
				Type:     gameConfig.Type,
				Name:     gameConfig.Name,
				Platform: gameConfig.Platform,
				DbName:   gameConfig.DbName,
				ApiId:    gameConfig.ApiId,
				ApiKey:   gameConfig.ApiKey,
				ApiUrl:   gameConfig.ApiUrl,
			},
			StartTime: startTime,
			EndTime:   endTime,
		}

		// 调用拉单方法
		records, err := gamePlatform.PullBetRecords(ctx, params)
		if err != nil {
			return 0, fmt.Errorf("拉取注单失败: %v", err)
		}

		return len(records), nil
	}

	// 正常拉单模式：从配置读取时间参数
	pullMinutes := g.Cfg().MustGet(ctx, "betRecordsPuller.pullMinutes", 10).Int()
	apiDelay := g.Cfg().MustGet(ctx, "betRecordsPuller.apiDelay", 5).Int()

	// 计算时间范围
	now := time.Now()
	endTimeCalc := now.Add(-time.Duration(apiDelay) * time.Minute)
	startTimeCalc := endTimeCalc.Add(-time.Duration(pullMinutes) * time.Minute)

	// 构建拉单参数
	params := &platform.PullRecordsParams{
		SiteId:   siteId,
		SiteCode: siteCode,
		SiteGame: entity.SiteGame{
			GameId:   gameConfig.GameId,
			Type:     gameConfig.Type,
			Name:     gameConfig.Name,
			Platform: gameConfig.Platform,
			DbName:   gameConfig.DbName,
			ApiId:    gameConfig.ApiId,
			ApiKey:   gameConfig.ApiKey,
			ApiUrl:   gameConfig.ApiUrl,
		},
		StartTime: startTimeCalc.Format("2006-01-02 15:04:05"),
		EndTime:   endTimeCalc.Format("2006-01-02 15:04:05"),
	}

	fmt.Printf("  时间范围: %s ~ %s\n", params.StartTime, params.EndTime)

	// 调用拉单方法（内部已处理批量入库MySQL）
	records, err := gamePlatform.PullBetRecords(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("拉取注单失败: %v", err)
	}

	// 返回记录数
	return len(records), nil
}
