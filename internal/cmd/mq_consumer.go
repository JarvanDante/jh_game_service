package cmd

import (
	"context"
	"jh_game_service/internal/mq"
	"jh_game_service/internal/registry"
	"os"
	"os/signal"
	"syscall"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	MqConsumer = &gcmd.Command{
		Name:  "mq-consumer",
		Usage: "mq-consumer",
		Brief: "启动 Game MQ Consumer，监听用户同步消息",
		Description: `
启动 Game MQ Consumer，监听来自 App 服务的用户同步消息，
并将用户信息同步到 user_mapping 表。

示例：
  ./jh_game_service mq-consumer
`,
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Info(ctx, "Game MQ Consumer 启动中...")

			// 初始化 Consul 客户端
			consulClient := registry.InitConsul()

			// 初始化数据库配置管理器
			dbManager := registry.GetDBConfigManager()
			if err := dbManager.Init(consulClient); err != nil {
				g.Log().Fatalf(ctx, "Failed to init db config manager: %v", err)
			}
			defer dbManager.Stop()

			// 初始化 RabbitMQ
			if err := mq.InitRabbitMQ(ctx); err != nil {
				g.Log().Fatalf(ctx, "Failed to init RabbitMQ: %v", err)
			}
			defer mq.CloseRabbitMQ()

			// 启动消费者
			if err := mq.StartSyncUserConsumer(ctx); err != nil {
				g.Log().Fatalf(ctx, "Failed to start consumer: %v", err)
			}

			// 等待中断信号
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			g.Log().Info(ctx, "Game MQ Consumer 停止")
			return nil
		},
	}
)

func init() {
	// 注册命令到主命令
	Main.AddCommand(MqConsumer)
}
