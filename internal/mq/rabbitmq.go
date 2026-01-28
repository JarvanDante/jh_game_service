package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"jh_game_service/internal/dao"
	"jh_game_service/internal/middleware"
	"jh_game_service/internal/model/do"
	"jh_game_service/internal/model/entity"
	"jh_game_service/internal/util"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	conn    *amqp.Connection
	channel *amqp.Channel
)

// SyncUserToGameMessage 同步用户到游戏服务消息
type SyncUserToGameMessage struct {
	SiteID   int32  `json:"site_id"`
	SiteCode string `json:"site_code"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

// SyncBetToClickHouseMessage 同步注单到 ClickHouse 消息
type SyncBetToClickHouseMessage struct {
	SiteID      uint32  `json:"site_id"`
	SiteCode    string  `json:"site_code"`
	UserID      uint64  `json:"user_id"`
	Username    string  `json:"username"`
	Platform    string  `json:"platform"`
	GameID      uint32  `json:"game_id"`
	GameName    string  `json:"game_name"`
	GameType    string  `json:"game_type"`
	OrderNo     string  `json:"order_no"`
	RoundID     string  `json:"round_id"`
	BetAmount   float64 `json:"bet_amount"`
	WinAmount   float64 `json:"win_amount"`
	ValidAmount float64 `json:"valid_amount"`
	NetAmount   float64 `json:"net_amount"`
	BetTime     string  `json:"bet_time"`
	SettleTime  string  `json:"settle_time"`
	Status      uint8   `json:"status"`
}

// InitRabbitMQ 初始化RabbitMQ连接
func InitRabbitMQ(ctx context.Context) error {
	url := g.Cfg().MustGet(ctx, "rabbitmq.url").String()
	if url == "" {
		g.Log().Warning(ctx, "RabbitMQ URL未配置，跳过初始化")
		return fmt.Errorf("RabbitMQ URL未配置")
	}

	var err error
	conn, err = amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("连接RabbitMQ失败: %v", err)
	}

	channel, err = conn.Channel()
	if err != nil {
		return fmt.Errorf("创建RabbitMQ channel失败: %v", err)
	}

	g.Log().Info(ctx, "RabbitMQ初始化成功")
	return nil
}

// CloseRabbitMQ 关闭RabbitMQ连接
func CloseRabbitMQ() {
	if channel != nil {
		channel.Close()
	}
	if conn != nil {
		conn.Close()
	}
}

// StartSyncUserConsumer 启动同步用户消费者
func StartSyncUserConsumer(ctx context.Context) error {
	if channel == nil {
		return fmt.Errorf("RabbitMQ channel未初始化")
	}

	// 声明 exchange
	err := channel.ExchangeDeclare(
		"user.created", // exchange name
		"topic",        // exchange type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("声明exchange失败: %v", err)
	}

	// 声明队列
	queue, err := channel.QueueDeclare(
		"game.sync_user", // queue name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return fmt.Errorf("声明队列失败: %v", err)
	}

	// 绑定队列到 exchange
	err = channel.QueueBind(
		queue.Name,          // queue name
		"user.sync_to_game", // routing key
		"user.created",      // exchange
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return fmt.Errorf("绑定队列失败: %v", err)
	}

	// 设置 QoS
	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("设置QoS失败: %v", err)
	}

	// 开始消费消息
	msgs, err := channel.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("开始消费失败: %v", err)
	}

	g.Log().Info(ctx, "Game MQ Consumer 已启动，等待消息...")

	// 处理消息
	go func() {
		for msg := range msgs {
			if err := handleSyncUserMessage(ctx, msg); err != nil {
				g.Log().Errorf(ctx, "处理消息失败: %v", err)
				msg.Nack(false, true) // 重新入队
			} else {
				msg.Ack(false) // 确认消息
			}
		}
	}()

	return nil
}

// handleSyncUserMessage 处理同步用户消息
func handleSyncUserMessage(ctx context.Context, msg amqp.Delivery) error {
	var syncMsg SyncUserToGameMessage
	if err := json.Unmarshal(msg.Body, &syncMsg); err != nil {
		g.Log().Errorf(ctx, "解析消息失败: %v, body: %s", err, string(msg.Body))
		return err
	}

	g.Log().Infof(ctx, "收到同步用户消息: site_id=%d, site_code=%s, user_id=%d, username=%s",
		syncMsg.SiteID, syncMsg.SiteCode, syncMsg.UserID, syncMsg.Username)

	// 创建带站点信息的 context
	dbGroup := fmt.Sprintf("site_%s", syncMsg.SiteCode)
	siteCtx := context.WithValue(ctx, "dynamic_db_group", dbGroup)
	siteCtx = context.WithValue(siteCtx, middleware.SiteCodeContextKey, syncMsg.SiteCode)

	// 检查用户是否已存在
	count, err := dao.UserMapping.Ctx(siteCtx).
		Where("site_id", syncMsg.SiteID).
		Where("user_id", syncMsg.UserID).
		Count()
	if err != nil {
		return fmt.Errorf("查询用户映射失败: %w", err)
	}

	if count > 0 {
		g.Log().Infof(siteCtx, "用户映射已存在，跳过: user_id=%d", syncMsg.UserID)
		return nil
	}

	buildUsername := util.BuildUsername(syncMsg.Username, syncMsg.SiteCode)
	// 插入用户映射
	userMapping := &do.UserMapping{
		SiteId:    syncMsg.SiteID,
		SiteCode:  syncMsg.SiteCode,
		UserId:    syncMsg.UserID,
		Username:  buildUsername,
		CreatedAt: gtime.Now(),
		UpdatedAt: gtime.Now(),
	}

	_, err = dao.UserMapping.Ctx(siteCtx).Insert(userMapping)
	if err != nil {
		return fmt.Errorf("插入用户映射失败: %w", err)
	}

	g.Log().Infof(siteCtx, "用户映射创建成功: site_id=%d, user_id=%d, username=%s",
		syncMsg.SiteID, syncMsg.UserID, syncMsg.Username)

	return nil
}

// PublishBetToClickHouse 发布注单同步到 ClickHouse 的消息
func PublishBetToClickHouse(ctx context.Context, msg *SyncBetToClickHouseMessage) error {
	if channel == nil {
		return fmt.Errorf("RabbitMQ channel未初始化")
	}

	// 声明 exchange（幂等操作）
	err := channel.ExchangeDeclare(
		"bet.sync", // exchange name
		"topic",    // exchange type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return fmt.Errorf("声明exchange失败: %v", err)
	}

	// 序列化消息
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	// 发布消息
	err = channel.Publish(
		"bet.sync",               // exchange
		"bet.sync_to_clickhouse", // routing key
		false,                    // mandatory
		false,                    // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // 持久化消息
		},
	)
	if err != nil {
		return fmt.Errorf("发布消息失败: %v", err)
	}

	return nil
}

// StartSyncBetToClickHouseConsumer 启动同步注单到 ClickHouse 消费者
func StartSyncBetToClickHouseConsumer(ctx context.Context) error {
	if channel == nil {
		return fmt.Errorf("RabbitMQ channel未初始化")
	}

	// 声明 exchange
	err := channel.ExchangeDeclare(
		"bet.sync", // exchange name
		"topic",    // exchange type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return fmt.Errorf("声明exchange失败: %v", err)
	}

	// 声明队列
	queue, err := channel.QueueDeclare(
		"clickhouse.sync_bet", // queue name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		return fmt.Errorf("声明队列失败: %v", err)
	}

	// 绑定队列到 exchange
	err = channel.QueueBind(
		queue.Name,               // queue name
		"bet.sync_to_clickhouse", // routing key
		"bet.sync",               // exchange
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		return fmt.Errorf("绑定队列失败: %v", err)
	}

	// 设置 QoS - 预取更多消息用于批量处理
	err = channel.Qos(
		5000,  // prefetch count (批量处理5000条)
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("设置QoS失败: %v", err)
	}

	// 开始消费消息
	msgs, err := channel.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("开始消费失败: %v", err)
	}

	g.Log().Info(ctx, "ClickHouse 同步消费者已启动（批量模式），等待消息...")

	// 批量处理消息
	go batchProcessSyncBetMessages(ctx, msgs)

	return nil
}

// batchProcessSyncBetMessages 批量处理同步注单消息
func batchProcessSyncBetMessages(ctx context.Context, msgs <-chan amqp.Delivery) {
	const (
		batchSize    = 5000            // 批量大小：5000条
		flushTimeout = 5 * time.Second // 刷新超时时间：5秒
	)

	var batch []*entity.ClickHouseBetRecord
	var batchMsgs []amqp.Delivery
	ticker := time.NewTicker(flushTimeout)
	defer ticker.Stop()

	flushBatch := func() {
		if len(batch) == 0 {
			return
		}

		g.Log().Infof(ctx, "开始批量插入ClickHouse: %d条记录", len(batch))

		// 批量插入到 ClickHouse
		if err := util.BatchInsertBetRecordsToCH(ctx, batch); err != nil {
			g.Log().Errorf(ctx, "批量插入ClickHouse失败: %v", err)
			// 批量Nack，重新入队
			for _, msg := range batchMsgs {
				msg.Nack(false, true)
			}
		} else {
			g.Log().Infof(ctx, "批量插入ClickHouse成功: %d条记录", len(batch))
			// 批量Ack
			for _, msg := range batchMsgs {
				msg.Ack(false)
			}
		}

		// 清空批次
		batch = batch[:0]
		batchMsgs = batchMsgs[:0]
	}

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				// 通道关闭，刷新剩余数据
				flushBatch()
				g.Log().Info(ctx, "消息通道已关闭，批量处理器退出")
				return
			}

			// 解析消息
			record, err := parseSyncBetMessage(ctx, msg)
			if err != nil {
				g.Log().Errorf(ctx, "解析消息失败: %v, body: %s", err, string(msg.Body))
				msg.Nack(false, false) // 丢弃无效消息
				continue
			}

			// 添加到批次
			batch = append(batch, record)
			batchMsgs = append(batchMsgs, msg)

			// 达到批量大小，立即刷新
			if len(batch) >= batchSize {
				flushBatch()
				ticker.Reset(flushTimeout) // 重置定时器
			}

		case <-ticker.C:
			// 定时刷新，避免消息积压太久
			if len(batch) > 0 {
				g.Log().Infof(ctx, "定时刷新批次: %d条记录", len(batch))
				flushBatch()
			}
		}
	}
}

// parseSyncBetMessage 解析同步注单消息
func parseSyncBetMessage(ctx context.Context, msg amqp.Delivery) (*entity.ClickHouseBetRecord, error) {
	var syncMsg SyncBetToClickHouseMessage
	if err := json.Unmarshal(msg.Body, &syncMsg); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	// 转换为 ClickHouse 记录
	record := &entity.ClickHouseBetRecord{
		SiteID:      syncMsg.SiteID,
		SiteCode:    syncMsg.SiteCode,
		UserID:      syncMsg.UserID,
		Username:    syncMsg.Username,
		Platform:    syncMsg.Platform,
		GameID:      syncMsg.GameID,
		GameName:    syncMsg.GameName,
		GameType:    syncMsg.GameType,
		OrderNo:     syncMsg.OrderNo,
		RoundID:     syncMsg.RoundID,
		BetAmount:   syncMsg.BetAmount,
		WinAmount:   syncMsg.WinAmount,
		ValidAmount: syncMsg.ValidAmount,
		NetAmount:   syncMsg.NetAmount,
		Status:      syncMsg.Status,
		UpdatedAt:   time.Now(), // 设置更新时间用于去重
	}

	// 解析时间
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if syncMsg.BetTime != "" {
		betTime, err := time.ParseInLocation("2006-01-02 15:04:05", syncMsg.BetTime, loc)
		if err == nil {
			record.BetTime = betTime
		}
	}
	if syncMsg.SettleTime != "" && syncMsg.SettleTime != "0000-00-00 00:00:00" {
		settleTime, err := time.ParseInLocation("2006-01-02 15:04:05", syncMsg.SettleTime, loc)
		if err == nil {
			record.SettleTime = &settleTime
		}
	}

	return record, nil
}

// handleSyncBetToClickHouseMessage 处理同步注单到 ClickHouse 消息（已废弃，使用批量处理）
// 保留此函数以便回滚
func handleSyncBetToClickHouseMessage(ctx context.Context, msg amqp.Delivery) error {
	var syncMsg SyncBetToClickHouseMessage
	if err := json.Unmarshal(msg.Body, &syncMsg); err != nil {
		g.Log().Errorf(ctx, "解析注单同步消息失败: %v, body: %s", err, string(msg.Body))
		return err
	}

	g.Log().Infof(ctx, "收到注单同步消息: site=%s, order=%s, user_id=%d, platform=%s",
		syncMsg.SiteCode, syncMsg.OrderNo, syncMsg.UserID, syncMsg.Platform)

	// 转换为 ClickHouse 记录
	record := &entity.ClickHouseBetRecord{
		SiteID:      syncMsg.SiteID,
		SiteCode:    syncMsg.SiteCode,
		UserID:      syncMsg.UserID,
		Username:    syncMsg.Username,
		Platform:    syncMsg.Platform,
		GameID:      syncMsg.GameID,
		GameName:    syncMsg.GameName,
		GameType:    syncMsg.GameType,
		OrderNo:     syncMsg.OrderNo,
		RoundID:     syncMsg.RoundID,
		BetAmount:   syncMsg.BetAmount,
		WinAmount:   syncMsg.WinAmount,
		ValidAmount: syncMsg.ValidAmount,
		NetAmount:   syncMsg.NetAmount,
		Status:      syncMsg.Status,
		UpdatedAt:   time.Now(), // 设置更新时间用于去重
	}

	// 解析时间
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if syncMsg.BetTime != "" {
		betTime, err := time.ParseInLocation("2006-01-02 15:04:05", syncMsg.BetTime, loc)
		if err == nil {
			record.BetTime = betTime
		}
	}
	if syncMsg.SettleTime != "" && syncMsg.SettleTime != "0000-00-00 00:00:00" {
		settleTime, err := time.ParseInLocation("2006-01-02 15:04:05", syncMsg.SettleTime, loc)
		if err == nil {
			record.SettleTime = &settleTime
		}
	}

	// 插入到 ClickHouse
	if err := util.InsertBetRecordToCH(ctx, record); err != nil {
		return fmt.Errorf("插入ClickHouse失败: %w", err)
	}

	g.Log().Infof(ctx, "注单同步到ClickHouse成功: site=%s, order=%s", syncMsg.SiteCode, syncMsg.OrderNo)
	return nil
}
