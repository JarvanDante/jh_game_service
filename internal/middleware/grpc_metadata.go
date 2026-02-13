package middleware

import (
	"context"
	"strconv"
	"strings"

	"google.golang.org/grpc/metadata"
)

//GetAdminIdFromGRPCMetadata
/**
 * @desc：从 gRPC metadata 中获取管理员ID
 * @param ctx
 * @return uint
 * @return bool
 * @author : Carson
 */
func GetAdminIdFromGRPCMetadata(ctx context.Context) (uint, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, false
	}

	adminIds := md.Get("admin_id")
	if len(adminIds) == 0 {
		return 0, false
	}

	adminId, err := strconv.ParseUint(adminIds[0], 10, 32)
	if err != nil {
		return 0, false
	}

	return uint(adminId), true
}

//GetSiteIdFromGRPCMetadata
/**
 * @desc：从 gRPC metadata 中获取商户ID
 * @param ctx
 * @return uint
 * @return bool
 * @author : Carson
 */
func GetSiteIdFromGRPCMetadata(ctx context.Context) (uint, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, false
	}

	siteIDs := md.Get("site_id")
	if len(siteIDs) == 0 {
		return 0, false
	}

	siteIDStr := strings.TrimSpace(siteIDs[0])
	if siteIDStr == "" {
		return 0, false
	}

	siteID, err := strconv.ParseInt(siteIDStr, 10, 64)
	if err != nil || siteID <= 0 {
		return 0, false
	}

	return uint(siteID), true
}

//GetSiteCodeFromGRPCMetadata
/**
 * @desc：从 gRPC metadata 中获取商户CODE
 * @param ctx
 * @return string
 * @return bool
 * @author : Carson
 */
func GetSiteCodeFromGRPCMetadata(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}

	siteCodes := md.Get("site_code")
	if len(siteCodes) == 0 {
		return "", false
	}

	siteCode := strings.TrimSpace(siteCodes[0])
	if siteCode == "" {
		return "", false
	}

	return siteCode, true
}

//GetUserIdFromGRPCMetadata
/**
 * @desc：从 gRPC metadata 中获取用户ID
 * @param ctx
 * @return uint
 * @return bool
 * @author : Carson
 */
func GetUserIdFromGRPCMetadata(ctx context.Context) (uint, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, false
	}

	userIds := md.Get("user_id")
	if len(userIds) == 0 {
		return 0, false
	}

	userId, err := strconv.ParseUint(userIds[0], 10, 32)
	if err != nil {
		return 0, false
	}

	return uint(userId), true
}

//GetCurrencyFromContext
/**
 * @desc：从上下文中获取currency（用于gRPC服务端）
 * @param ctx
 * @return string
 * @author : Carson
 */
func GetCurrencyFromContext(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if currencies := md.Get("currency"); len(currencies) > 0 {
			return strings.TrimSpace(currencies[0])
		}
	}
	return ""
}

// GetClientIPFromGRPCMetadata 从 gRPC metadata 中获取客户端IP
func GetClientIPFromGRPCMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "127.0.0.1"
	}

	ips := md.Get("client_ip")
	if len(ips) == 0 {
		return "127.0.0.1"
	}

	return ips[0]
}
