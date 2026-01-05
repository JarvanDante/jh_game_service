package middleware

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/stats"
)

// TraceStatsHandler 实现 gRPC stats.Handler 接口，用于追踪和统计
type TraceStatsHandler struct{}

// NewTraceStatsHandler 创建新的追踪统计处理器
func NewTraceStatsHandler() *TraceStatsHandler {
	return &TraceStatsHandler{}
}

// TagRPC 在 RPC 开始时调用
func (h *TraceStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	tracer := otel.Tracer("jh_game_service")
	ctx, span := tracer.Start(ctx, info.FullMethodName,
		trace.WithAttributes(
			attribute.String("rpc.service", "jh_game_service"),
			attribute.String("rpc.method", info.FullMethodName),
		),
	)

	// 将 span 存储到上下文中
	return context.WithValue(ctx, "grpc_span", span)
}

// HandleRPC 处理 RPC 统计信息
func (h *TraceStatsHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	span, ok := ctx.Value("grpc_span").(trace.Span)
	if !ok {
		return
	}

	switch st := s.(type) {
	case *stats.Begin:
		// RPC 开始
		span.SetAttributes(
			attribute.String("rpc.phase", "begin"),
			attribute.Bool("rpc.client", st.Client),
		)
		g.Log().Debugf(ctx, "gRPC call begin: client=%v", st.Client)

	case *stats.InPayload:
		// 接收到输入数据
		span.SetAttributes(
			attribute.String("rpc.phase", "in_payload"),
			attribute.Int("rpc.payload.length", st.Length),
		)

	case *stats.OutPayload:
		// 发送输出数据
		span.SetAttributes(
			attribute.String("rpc.phase", "out_payload"),
			attribute.Int("rpc.payload.length", st.Length),
		)

	case *stats.End:
		// RPC 结束
		span.SetAttributes(
			attribute.String("rpc.phase", "end"),
			attribute.String("rpc.end_time", st.EndTime.Format(time.RFC3339)),
		)

		if st.Error != nil {
			span.SetAttributes(
				attribute.Bool("rpc.error", true),
				attribute.String("rpc.error_message", st.Error.Error()),
			)
			g.Log().Errorf(ctx, "gRPC call error: %v", st.Error)
		} else {
			span.SetAttributes(attribute.Bool("rpc.success", true))
			g.Log().Debugf(ctx, "gRPC call success")
		}

		span.End()
	}
}

// TagConn 在连接开始时调用
func (h *TraceStatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	g.Log().Debugf(ctx, "gRPC connection tagged: %+v", info)
	return ctx
}

// HandleConn 处理连接统计信息
func (h *TraceStatsHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	switch st := s.(type) {
	case *stats.ConnBegin:
		g.Log().Debugf(ctx, "gRPC connection begin: client=%v", st.Client)
	case *stats.ConnEnd:
		g.Log().Debugf(ctx, "gRPC connection end: client=%v", st.Client)
	}
}
