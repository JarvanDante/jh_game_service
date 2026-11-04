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

// coverURLMatchesAllowedPrefix 按解析后的 scheme+host（及可选 path 前缀）匹配，
// 避免 https://cdn.example.com@evil.com / https://cdn.example.com.evil.com 等原始字符串前缀绕过。
// 若配置项无法解析为带 scheme+host 的 URL，则要求该配置以 "/" 结尾并以规范化 URL 字符串做前缀匹配。
func coverURLMatchesAllowedPrefix(u *url.URL, prefix string) bool {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || u == nil {
		return false
	}
	p, err := url.Parse(prefix)
	if err == nil && p.Scheme != "" && p.Host != "" {
		if !strings.EqualFold(u.Scheme, p.Scheme) {
			return false
		}
		if !strings.EqualFold(u.Host, p.Host) {
			return false
		}
		prefPath := p.EscapedPath()
		if prefPath == "" {
			prefPath = "/"
		}
		if prefPath != "/" {
			if !strings.HasSuffix(prefPath, "/") {
				prefPath += "/"
			}
			path := u.EscapedPath()
			if path == "" {
				path = "/"
			}
			trim := strings.TrimSuffix(prefPath, "/")
			return path == trim || strings.HasPrefix(path, prefPath)
		}
		return true
	}
	// 无法按 URL 解析时：必须显式以 "/" 结尾，且对规范化 URL 做前缀匹配
	if !strings.HasSuffix(prefix, "/") {
		return false
	}
	canon := u.String()
	return strings.HasPrefix(canon, prefix)
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
			if coverURLMatchesAllowedPrefix(u, p) {
				return nil
			}
		}
		return fmt.Errorf("封面地址不在允许的存储域名范围内")
	}
	return nil
}
