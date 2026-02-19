package agent_settlement

import (
	"context"
	v1 "jh_game_service/api/grpc/agent_settlement/v1"
	"jh_game_service/internal/service/backend"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	v1.UnimplementedAgentSettlementServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterAgentSettlementServer(s.Server, &Controller{})
}

func (*Controller) GetAgentDailyBetStats(ctx context.Context, req *v1.GetAgentDailyBetStatsReq) (*v1.GetAgentDailyBetStatsRes, error) {
	return backend.AgentSettlement().GetAgentDailyBetStats(ctx, req)
}
