package cmd

import (
	"context"
	"fmt"
	"jh_game_service/internal/controller/backend/game"
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
			registry.InitConsul()

			// 注册服务到 Consul
			fmt.Println("注册服务到Consul...")
			if err := registry.RegisterService(); err != nil {
				fmt.Printf("Consul注册失败: %v\n", err)
				g.Log().Fatalf(ctx, "register service failed: %v", err)
			}

			fmt.Println("Consul服务注册成功")

			c := grpcx.Server.NewConfig()
			c.Options = append(c.Options, []grpc.ServerOption{
				// 使用 StatsHandler 替代 Interceptor 进行追踪和统计
				grpc.StatsHandler(middleware.NewTraceStatsHandler()),
				// 保留验证拦截器
				grpcx.Server.ChainUnary(
					grpcx.Server.UnaryValidate,
				),
			}...)
			s := grpcx.Server.New(c)
			game.Register(s)

			fmt.Println("gRPC服务器启动中...")

			// 强制刷新输出
			os.Stdout.Sync()

			s.Run()
			return nil
		},
	}
)
