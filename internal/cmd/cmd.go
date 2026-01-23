package cmd

import (
	"context"
	"fmt"
	"jh_game_service/internal/controller/backend/game"
	frontendGame "jh_game_service/internal/controller/frontend/game"
	"os"
	"time"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"google.golang.org/grpc"

	"jh_game_service/internal/middleware"
	"jh_game_service/internal/registry"
	"jh_game_service/internal/tracing"
)

var (
	// Main is the main command.
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start grpc server of jh_game_service",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 强制输出到stdout确保能看到
			fmt.Println("=== jh_game_service 启动 ===")
			fmt.Printf("启动时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
			fmt.Printf("进程ID: %d\n", os.Getpid())
			fmt.Println("==============================")

			// 初始化Jaeger追踪
			cleanup, err := tracing.InitJaeger()
			if err != nil {
				fmt.Printf("初始化Jaeger失败: %v\n", err)
				g.Log().Errorf(ctx, "初始化Jaeger失败: %v", err)
			} else {
				fmt.Println("Jaeger追踪初始化成功")
				g.Log().Info(ctx, "Jaeger追踪初始化成功")
				defer cleanup()
			}

			// 初始化 Consul 客户端
			fmt.Println("初始化Consul客户端...")
			consulClient := registry.InitConsul()

			// 注册服务到 Consul
			fmt.Println("注册服务到Consul...")
			if err := registry.RegisterService(); err != nil {
				fmt.Printf("Consul注册失败: %v\n", err)
				g.Log().Fatalf(ctx, "register service failed: %v", err)
			}

			fmt.Println("Consul服务注册成功")

			// 初始化数据库配置管理器
			fmt.Println("初始化Game服务数据库配置管理器...")
			dbManager := registry.GetDBConfigManager()
			if err := dbManager.Init(consulClient); err != nil {
				fmt.Printf("Game服务数据库配置管理器初始化失败: %v\n", err)
				g.Log().Fatalf(ctx, "init game db config manager failed: %v", err)
			}
			fmt.Println("Game服务数据库配置管理器初始化成功")

			// 初始化ClickHouse连接
			fmt.Println("初始化ClickHouse连接...")
			chHost := g.Cfg().MustGet(ctx, "clickhouse.host").String()
			chPort := g.Cfg().MustGet(ctx, "clickhouse.port").Int()
			chDatabase := g.Cfg().MustGet(ctx, "clickhouse.database").String()
			chUsername := g.Cfg().MustGet(ctx, "clickhouse.username").String()
			chPassword := g.Cfg().MustGet(ctx, "clickhouse.password").String()

			chManager := registry.GetClickHouseManager()
			if err := chManager.Init(chHost, chPort, chDatabase, chUsername, chPassword); err != nil {
				fmt.Printf("ClickHouse连接初始化失败: %v\n", err)
				g.Log().Errorf(ctx, "ClickHouse连接初始化失败: %v", err)
				// 不中断服务，继续运行
			} else {
				fmt.Println("ClickHouse连接初始化成功")
				defer chManager.Close()
			}

			// 确保在退出时停止数据库配置管理器
			defer dbManager.Stop()

			// 注单拉取使用K8s CronJob方式执行：
			// ./jh_game_service pull ONE
			// ./jh_game_service pull IG_JILI

			// 创建数据库拦截器
			dbInterceptor := middleware.NewDBInterceptor()

			c := grpcx.Server.NewConfig()
			c.Options = append(c.Options, []grpc.ServerOption{
				// 使用 StatsHandler 替代 Interceptor 进行追踪和统计
				grpc.StatsHandler(middleware.NewTraceStatsHandler()),
				// 添加数据库拦截器和验证拦截器
				grpcx.Server.ChainUnary(
					dbInterceptor.UnaryServerInterceptor(),
					grpcx.Server.UnaryValidate,
				),
				grpcx.Server.ChainStream(
					dbInterceptor.StreamServerInterceptor(),
				),
			}...)
			s := grpcx.Server.New(c)
			game.Register(s)
			frontendGame.Register(s)

			fmt.Println("gRPC服务器启动中...")

			// 强制刷新输出
			os.Stdout.Sync()

			s.Run()
			return nil
		},
	}
)
