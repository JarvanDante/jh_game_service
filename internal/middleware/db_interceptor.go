package middleware

import (
	"context"
	"strconv"
	"strings"

	"jh_game_service/internal/registry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// DBInterceptor 数据库拦截器 - 完全无侵入的动态数据库路由
type DBInterceptor struct {
	dbManager *registry.DBConfigManager
}

// NewDBInterceptor 创建数据库拦截器
func NewDBInterceptor() *DBInterceptor {
	return &DBInterceptor{
		dbManager: registry.GetDBConfigManager(),
	}
}

// UnaryServerInterceptor 一元RPC拦截器
func (interceptor *DBInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 创建带有动态数据库的上下文
		newCtx := interceptor.createDynamicDBContext(ctx)

		// 调用原始处理器
		return handler(newCtx, req)
	}
}

// StreamServerInterceptor 流式RPC拦截器
func (interceptor *DBInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 创建带有动态数据库的上下文
		newCtx := interceptor.createDynamicDBContext(ss.Context())

		// 包装ServerStream
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		// 调用原始处理器
		return handler(srv, wrappedStream)
	}
}

// createDynamicDBContext 创建带有动态数据库的上下文
func (interceptor *DBInterceptor) createDynamicDBContext(ctx context.Context) context.Context {
	// 获取站点代码
	siteCode := interceptor.getSiteCodeFromContext(ctx)
	if siteCode == "" {
		// 没有站点信息，返回原始上下文
		return ctx
	}

	// 将站点代码设置到上下文中，供DAO层使用
	return context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
}

// getSiteCodeFromContext 从上下文获取站点代码
func (interceptor *DBInterceptor) getSiteCodeFromContext(ctx context.Context) string {
	// 首先尝试从上下文中获取site_code
	if siteCode := ctx.Value("site_code"); siteCode != nil {
		if siteCodeStr, ok := siteCode.(string); ok {
			return siteCodeStr
		}
	}

	// 如果没有site_code，尝试从gRPC metadata获取site_id
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if siteIDs := md.Get("site_id"); len(siteIDs) > 0 {
			if siteIDStr := strings.TrimSpace(siteIDs[0]); siteIDStr != "" {
				// 转换为int64
				if siteID, err := strconv.ParseInt(siteIDStr, 10, 64); err == nil && siteID > 0 {
					// 根据site_id查找对应的site_code
					return interceptor.dbManager.GetSiteCodeBySiteID(siteID)
				}
			}
		}
	}

	return ""
}

// wrappedServerStream 包装的ServerStream
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context 返回包装的上下文
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
