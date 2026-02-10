package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gtime"

	"jh_game_service/internal/model/entity"
	"jh_game_service/internal/mq"
	"jh_game_service/internal/registry"
)

var (
	OutboxRelay = &gcmd.Command{
		Name:  "outbox-relay",
		Usage: "outbox-relay",
		Brief: "启动Outbox中继，将mq_outbox消息投递到MQ",
		Description: `
启动Outbox中继，轮询各站点数据库的 mq_outbox 表，
将待发送消息投递到MQ，并更新发送状态。

示例：
  ./jh_game_service outbox-relay
`,
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Info(ctx, "Outbox Relay 启动中...")

			// 初始化 Consul 客户端
			consulClient := registry.InitConsul()

			// 初始化数据库配置管理器
			dbManager := registry.GetDBConfigManager()
			if err := dbManager.Init(consulClient); err != nil {
				return fmt.Errorf("初始化数据库配置管理器失败: %v", err)
			}
			defer dbManager.Stop()

			// 初始化 RabbitMQ
			if err := mq.InitRabbitMQ(ctx); err != nil {
				return fmt.Errorf("RabbitMQ连接初始化失败: %v", err)
			}
			defer mq.CloseRabbitMQ()

			siteCodes := dbManager.GetAllSiteCodes()
			if len(siteCodes) == 0 {
				return fmt.Errorf("未获取到站点配置")
			}

			g.Log().Infof(ctx, "Outbox Relay 已启动，站点数=%d", len(siteCodes))

			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			// 监听退出信号
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			for {
				select {
				case <-ticker.C:
					for _, siteCode := range siteCodes {
						groupName := fmt.Sprintf("site_%s", siteCode)
						db := g.DB(groupName)
						if db == nil {
							g.Log().Warningf(ctx, "数据库连接不存在: %s", groupName)
							continue
						}

						// 单站点循环拉取，直到不足一批
						for {
							processed, err := processOutboxBatch(ctx, db, siteCode)
							if err != nil {
								g.Log().Errorf(ctx, "处理Outbox失败: site=%s, err=%v", siteCode, err)
								break
							}
							if processed < outboxBatchSize {
								break
							}
						}
					}
				case <-sigChan:
					g.Log().Info(ctx, "Outbox Relay 停止")
					return nil
				}
			}
		},
	}
)

const (
	outboxBatchSize   = 200
	outboxMaxDelaySec = 300
)

var outboxMissing sync.Map

func processOutboxBatch(ctx context.Context, db gdb.DB, siteCode string) (int, error) {
	if _, ok := outboxMissing.Load(siteCode); ok {
		return 0, nil
	}

	now := gtime.Now()
	var items []entity.MqOutbox
	err := db.Model("mq_outbox").
		WhereIn("status", []int{0, 2}).
		Where("next_retry_at IS NULL OR next_retry_at <= ?", now).
		Order("id ASC").
		Limit(outboxBatchSize).
		Scan(&items)
	if err != nil {
		if isTableNotFoundErr(err) {
			outboxMissing.Store(siteCode, struct{}{})
			g.Log().Warningf(ctx, "站点 %s 缺少mq_outbox表，已跳过", siteCode)
			return 0, nil
		}
		return 0, err
	}

	if len(items) == 0 {
		return 0, nil
	}

	for _, item := range items {
		if item.Exchange == "" || item.RoutingKey == "" {
			_ = markOutboxFailed(ctx, db, item.Id, item.RetryCount+1, "exchange或routing_key为空")
			continue
		}

		if err := mq.PublishRaw(ctx, item.Exchange, item.RoutingKey, []byte(item.Payload)); err != nil {
			_ = markOutboxFailed(ctx, db, item.Id, item.RetryCount+1, err.Error())
			continue
		}

		if _, err := db.Model("mq_outbox").
			Where("id", item.Id).
			Data(g.Map{
				"status":     1,
				"updated_at": now,
			}).Update(); err != nil {
			g.Log().Warningf(ctx, "更新Outbox状态失败: id=%d, err=%v", item.Id, err)
		}
	}

	return len(items), nil
}

func isTableNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "doesn't exist") || strings.Contains(msg, "error 1146") || strings.Contains(msg, "no such table")
}

func markOutboxFailed(ctx context.Context, db gdb.DB, id uint64, retryCount int, lastErr string) error {
	delay := time.Duration(retryCount*5) * time.Second
	if delay > outboxMaxDelaySec*time.Second {
		delay = outboxMaxDelaySec * time.Second
	}
	nextRetryAt := gtime.New(time.Now().Add(delay))

	if len(lastErr) > 255 {
		lastErr = lastErr[:255]
	}

	_, err := db.Model("mq_outbox").
		Where("id", id).
		Data(g.Map{
			"status":        2,
			"retry_count":   retryCount,
			"next_retry_at": nextRetryAt,
			"last_error":    lastErr,
			"updated_at":    gtime.Now(),
		}).Update()
	return err
}
