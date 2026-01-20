package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// TraceInterceptor gRPC拦截器，用于OpenTelemetry trace context处理
func TraceInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// OpenTelemetry的gRPC拦截器已经处理了trace context的传播
	// 这里我们只需要记录一下接收到的trace信息（可选）

	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		traceID := span.SpanContext().TraceID().String()
		g.Log().Debugf(ctx, "处理gRPC请求: %s, TraceID: %s", info.FullMethod, traceID)
	}

	// 调用实际的处理函数
	return handler(ctx, req)
}

// GetTraceIDFromContext 从上下文中获取OpenTelemetry的TraceID
func GetTraceIDFromContext(ctx context.Context) string {
	// 只使用OpenTelemetry的TraceID
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}

	return ""
}

// LogWithTrace 带traceId的日志记录 - JSON格式
func LogWithTrace(ctx context.Context, level string, message string, args ...interface{}) {
	traceID := GetTraceIDFromContext(ctx)

	// 构建结构化日志数据
	logData := g.Map{
		"service":  "admin_service",
		"trace_id": traceID,
		"msg":      fmt.Sprintf(message, args...),
	}

	// 根据级别记录日志
	switch strings.ToLower(level) {
	case "debug":
		g.Log().Debug(ctx, logData)
	case "info":
		g.Log().Info(ctx, logData)
	case "warn", "warning":
		g.Log().Warning(ctx, logData)
	case "error":
		g.Log().Error(ctx, logData)
	case "fatal":
		g.Log().Fatal(ctx, logData)
	default:
		g.Log().Info(ctx, logData)
	}
}

// LogWithTraceAndFields 带traceId和额外字段的日志记录 - JSON格式
func LogWithTraceAndFields(ctx context.Context, level string, message string, fields g.Map, args ...interface{}) {
	traceID := GetTraceIDFromContext(ctx)

	// 构建结构化日志数据
	logData := g.Map{
		"service":  "admin_service",
		"trace_id": traceID,
		"msg":      fmt.Sprintf(message, args...),
	}

	// 添加额外字段
	for k, v := range fields {
		logData[k] = v
	}

	// 根据级别记录日志
	switch strings.ToLower(level) {
	case "debug":
		g.Log().Debug(ctx, logData)
	case "info":
		g.Log().Info(ctx, logData)
	case "warn", "warning":
		g.Log().Warning(ctx, logData)
	case "error":
		g.Log().Error(ctx, logData)
	case "fatal":
		g.Log().Fatal(ctx, logData)
	default:
		g.Log().Info(ctx, logData)
	}
}
