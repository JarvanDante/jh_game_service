package cmd

import (
	"context"
	"jh_game_service/internal/mq"
	"jh_game_service/internal/registry"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	ClickHouseSyncConsumer = &gcmd.Command{
		Name:  "clickhouse-sync",
		Usage: "clickhouse-sync",
		Brief: "启动 ClickHouse 同步消费者，监听注单与分析数据同步消息",
		Description: `
启动 ClickHouse 同步消费者，监听来自拉单服务的注单同步消息，
以及来自 Outbox 的用户/资金日志同步消息，
并将数据写入 ClickHouse 的 game_analytics 相关表。

示例：
  ./jh_game_service clickhouse-sync
`,
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Info(ctx, "ClickHouse 同步消费者启动中...")

			// 初始化 ClickHouse
			g.Log().Info(ctx, "初始化 ClickHouse 连接...")
			chHost := g.Cfg().MustGet(ctx, "clickhouse.host").String()
			chPort := g.Cfg().MustGet(ctx, "clickhouse.port").Int()
			chDatabase := g.Cfg().MustGet(ctx, "clickhouse.database").String()
			chUsername := g.Cfg().MustGet(ctx, "clickhouse.username").String()
			chPassword := g.Cfg().MustGet(ctx, "clickhouse.password").String()

			chManager := registry.GetClickHouseManager()
			if err := chManager.Init(chHost, chPort, chDatabase, chUsername, chPassword); err != nil {
				g.Log().Fatalf(ctx, "ClickHouse 连接初始化失败: %v", err)
			}
			defer chManager.Close()
			g.Log().Info(ctx, "ClickHouse 连接初始化成功")

			// 初始化 RabbitMQ
			if err := mq.InitRabbitMQ(ctx); err != nil {
				g.Log().Fatalf(ctx, "Failed to init RabbitMQ: %v", err)
			}
			defer mq.CloseRabbitMQ()

			startConsumers := func() error {
				if err := mq.StartSyncBetToClickHouseConsumer(ctx); err != nil {
					return err
				}
				if err := mq.StartAnalyticsSyncConsumers(ctx); err != nil {
					return err
				}
				g.Log().Info(ctx, "ClickHouse 消费者已启动")
				return nil
			}

			// 首次启动消费者
			if err := startConsumers(); err != nil {
				g.Log().Fatalf(ctx, "Failed to start ClickHouse sync consumer: %v", err)
			}

			// 连接健康巡检：MQ断线后自动重连并恢复消费
			healthTicker := time.NewTicker(5 * time.Second)
			defer healthTicker.Stop()

			// 等待中断信号
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			for {
				select {
				case <-healthTicker.C:
					if mq.IsRabbitMQHealthy() {
						continue
					}
					g.Log().Warning(ctx, "检测到MQ连接异常，开始重连并恢复消费者...")
					if err := mq.ReconnectRabbitMQ(ctx); err != nil {
						g.Log().Errorf(ctx, "MQ重连失败: %v", err)
						continue
					}
					if err := startConsumers(); err != nil {
						g.Log().Errorf(ctx, "重连后恢复消费者失败: %v", err)
						continue
					}
					g.Log().Info(ctx, "MQ重连并恢复消费者成功")
				case <-sigChan:
					g.Log().Info(ctx, "ClickHouse 同步消费者停止")
					return nil
				}
			}
		},
	}
)
