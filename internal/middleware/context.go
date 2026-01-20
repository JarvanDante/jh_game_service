package middleware

import (
	"context"
	"fmt"
	"google.golang.org/grpc/metadata"
	"strings"
)

const (
	SiteIdContextKey   = "site_id"
	SiteCodeContextKey = "site_code"
	AdminIdContextKey  = "admin_id"
	UserIdContextKey   = "user_id"
	ClientIPContextKey = "client_ip"
)

// GetAdminIdFromContext 从上下文中获取管理员ID
// 优先从 gRPC metadata 获取，如果没有再从 context 获取
func GetAdminIdFromContext(ctx context.Context) (uint, bool) {
	// 首先尝试从 gRPC metadata 获取
	if adminId, ok := GetAdminIdFromGRPCMetadata(ctx); ok {
		return adminId, true
	}

	// 如果 gRPC metadata 中没有，尝试从 context 获取
	if adminId, ok := ctx.Value(SiteIdContextKey).(uint); ok {
		return adminId, true
	}
	return 0, false
}

//GetSiteIdFromContext
/**
 * @desc：获取siteId -优先从 gRPC metadata 获取，如果没有再从 context 获取
 * @param ctx
 * @return uint
 * @return bool
 * @author : Carson
 */
func GetSiteIdFromContext(ctx context.Context) (uint, bool) {
	// 首先尝试从 gRPC metadata 获取
	if siteId, ok := GetSiteIdFromGRPCMetadata(ctx); ok {
		return siteId, true
	}

	// 如果 gRPC metadata 中没有，尝试从 context 获取
	if siteId, ok := ctx.Value(AdminIdContextKey).(uint); ok {
		return siteId, true
	}
	return 0, false
}

//GetUserIdFromContext
/**
 * @desc：获取userId -优先从 gRPC metadata 获取，如果没有再从 context 获取
 * @param ctx
 * @return uint
 * @return bool
 * @author : Carson
 */
func GetUserIdFromContext(ctx context.Context) (uint, bool) {
	// 首先尝试从 gRPC metadata 获取
	if userId, ok := GetUserIdFromGRPCMetadata(ctx); ok {
		return userId, true
	}

	// 如果 gRPC metadata 中没有，尝试从 context 获取
	if userId, ok := ctx.Value(UserIdContextKey).(uint); ok {
		return userId, true
	}
	return 0, false
}

//GetSiteCodeFromContext
/**
 * @desc：获取siteCode -优先从 gRPC metadata 获取，如果没有再从 context 获取
 * @param ctx
 * @return uint
 * @return bool
 * @author : Carson
 */
func GetSiteCodeFromContext(ctx context.Context) (string, bool) {
	// 首先尝试从 gRPC metadata 获取
	if siteCode, ok := GetSiteCodeFromGRPCMetadata(ctx); ok {
		return siteCode, true
	}

	// 如果 gRPC metadata 中没有，尝试从 context 获取
	if siteCode, ok := ctx.Value(SiteCodeContextKey).(string); ok {
		return siteCode, true
	}
	return "", false
}

// SetAdminIdToContext 将管理员ID设置到上下文中
func SetAdminIdToContext(ctx context.Context, adminId uint) context.Context {
	return context.WithValue(ctx, AdminIdContextKey, adminId)
}

//GetUserInfoFromContext
/**
 * @desc：从上下文获取用户信息
 * @param ctx
 * @return userId
 * @return siteId
 * @return err
 * @author : Carson
 */
func GetUserInfoFromContext(ctx context.Context) (userId uint, siteId int, err error) {
	// 尝试从gRPC metadata获取
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// 获取user_id
		if userIds := md.Get("user_id"); len(userIds) > 0 {
			var userIdInt int
			_, err := fmt.Sscanf(userIds[0], "%d", &userIdInt)
			if err == nil && userIdInt > 0 {
				userId = uint(userIdInt)
			}
		}

		// 获取site_id
		if siteIds := md.Get("site_id"); len(siteIds) > 0 {
			_, err := fmt.Sscanf(siteIds[0], "%d", &siteId)
			if err == nil && siteId > 0 {
				// siteId已设置
			}
		}
	}

	// 如果metadata中没有，尝试从context.Value获取
	if userId == 0 {
		if uid := ctx.Value("user_id"); uid != nil {
			switch v := uid.(type) {
			case uint:
				userId = v
			case int:
				userId = uint(v)
			case int64:
				userId = uint(v)
			}
		}
	}

	if siteId == 0 {
		if sid := ctx.Value("site_id"); sid != nil {
			switch v := sid.(type) {
			case int:
				siteId = v
			case int64:
				siteId = int(v)
			case uint:
				siteId = int(v)
			}
		}
	}

	if userId == 0 {
		return 0, 0, fmt.Errorf("用户未登录")
	}

	if siteId == 0 {
		return 0, 0, fmt.Errorf("站点信息缺失")
	}

	return userId, siteId, nil
}

//GetUsernameFromContext
/**
 * @desc：从上下文获取用户名
 * @param ctx
 * @return string
 * @author : Carson
 */
func GetUsernameFromContext(ctx context.Context) string {
	// 尝试从gRPC metadata获取
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if usernames := md.Get("username"); len(usernames) > 0 {
			return usernames[0]
		}
	}

	// 尝试从context.Value获取
	if username := ctx.Value("username"); username != nil {
		if usernameStr, ok := username.(string); ok {
			return usernameStr
		}
	}

	return ""
}

//GetDomainFromContext
/**
 * @desc：从上下文中获取domain（用于gRPC服务端）
 * @param ctx
 * @return string
 * @author : Carson
 */
func GetDomainFromContext(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if domains := md.Get("domain"); len(domains) > 0 {
			return strings.TrimSpace(domains[0])
		}
	}
	return ""
}

//GetClientIPFromContext
/**
 * @desc：从上下文中获取客户端IP
 * @param ctx
 * @return string
 * @author : Carson
 */
func GetClientIPFromContext(ctx context.Context) string {
	// 首先尝试从 gRPC metadata 获取
	if ip := GetClientIPFromGRPCMetadata(ctx); ip != "127.0.0.1" {
		return ip
	}

	// 如果 gRPC metadata 中没有，尝试从 context 获取
	if ip, ok := ctx.Value(ClientIPContextKey).(string); ok {
		return ip
	}
	return "127.0.0.1"
}

// SetClientIPToContext 将客户端IP设置到上下文中
func SetClientIPToContext(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ClientIPContextKey, ip)
}
