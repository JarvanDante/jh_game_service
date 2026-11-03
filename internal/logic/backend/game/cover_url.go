package game

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// maxCoverURLLen 与 game_list.cover_url 列长度保持一致
const maxCoverURLLen = 500

// coverURLAllowedPrefixes 可选配置 game.coverUrlAllowedPrefixes（字符串数组）；
// 未配置时不做前缀限制，只校验协议/host/长度。
func coverURLAllowedPrefixes(ctx context.Context) []string {
	v := g.Cfg().MustGet(ctx, "game.coverUrlAllowedPrefixes")
	if v == nil || v.IsEmpty() {
		return nil
	}
	var out []string
	for _, p := range v.Strings() {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// validateCoverURL 校验自有封面地址：仅 http/https、host 非空、长度 <= 500，可选前缀白名单
func validateCoverURL(raw string, allowedPrefixes []string) error {
	if len(raw) > maxCoverURLLen {
		return fmt.Errorf("封面地址长度不能超过%d个字符", maxCoverURLLen)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("封面地址格式不正确")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("封面地址仅支持 http/https")
	}
	if u.Hostname() == "" {
		return fmt.Errorf("封面地址缺少域名")
	}
	if len(allowedPrefixes) > 0 {
		for _, p := range allowedPrefixes {
			if strings.HasPrefix(raw, p) {
				return nil
			}
		}
		return fmt.Errorf("封面地址不在允许的存储域名范围内")
	}
	return nil
}
