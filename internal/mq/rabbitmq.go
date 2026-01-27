package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"jh_game_service/internal/dao"
	"jh_game_service/internal/middleware"
	"jh_game_service/internal/model/do"
	"jh_game_service/internal/util"

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
