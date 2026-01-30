package grpc_client

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	balanceV1 "jh_game_service/api/grpc/balance/v1"
	"jh_game_service/internal/registry"
)

// GetBalanceClient 获取 Balance 服务的 gRPC 客户端
func GetBalanceClient(ctx context.Context) (balanceV1.BalanceClient, *grpc.ClientConn, error) {
	// 从 Consul 获取 balance 服务地址
	addr, err := registry.GetServiceAddr("jh_balance_service")
	if err != nil {
		return nil, nil, fmt.Errorf("从Consul获取balance服务地址失败: %v", err)
	}

	g.Log().Infof(ctx, "balance服务地址: %s", addr)

	// 创建 gRPC 连接
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("无法连接到balance服务: %v", err)
	}

	return balanceV1.NewBalanceClient(conn), conn, nil
}
