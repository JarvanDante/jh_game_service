package game

import (
	"strings"
	"testing"
)

func TestValidateCoverURL(t *testing.T) {
	ok := []string{
		"https://jh-minio.panalow.cc/uploads/a/2026/11/x.png",
		"http://localhost:19000/uploads/x.png",
		"HTTPS://example.com/a.jpg",
		"https://h/" + strings.Repeat("a", 490),
	}
	for _, u := range ok {
		if err := validateCoverURL(u, nil); err != nil {
			t.Errorf("%q should pass: %v", u, err)
		}
	}
	bad := []string{
		"javascript:alert(1)",
		"ftp://example.com/a.png",
		"data:image/png;base64,AAAA",
		"//example.com/a.png",
		"/uploads/a.png",
		"http:///a.png",
		"https://example.com/" + strings.Repeat("a", 500),
		"http://exa mple.com/%zz",
	}
	for _, u := range bad {
		if err := validateCoverURL(u, nil); err == nil {
			t.Errorf("%q should be rejected", u)
		}
	}
	prefixes := []string{"https://jh-minio.panalow.cc/"}
	if err := validateCoverURL("https://jh-minio.panalow.cc/uploads/a.png", prefixes); err != nil {
		t.Errorf("prefix match should pass: %v", err)
	}
	if err := validateCoverURL("https://evil.com/a.png", prefixes); err == nil {
		t.Error("non-matching prefix should be rejected")
	}
}

func TestValidateCoverURL_PrefixBypass(t *testing.T) {
	prefixes := []string{"https://cdn.example.com/"}
	// 用户信息 / 子域名绕过：原始字符串前缀会误判，scheme+host 匹配应拒绝
	bypass := []string{
		"https://cdn.example.com@evil.com/a.png",
		"https://cdn.example.com.evil.com/a.png",
		"https://evil.com/cdn.example.com/a.png",
		"http://cdn.example.com/a.png", // scheme 不同
	}
	for _, u := range bypass {
		if err := validateCoverURL(u, prefixes); err == nil {
			t.Errorf("%q should be rejected by host-aware prefix match", u)
		}
	}
	if err := validateCoverURL("https://cdn.example.com/uploads/a.png", prefixes); err != nil {
		t.Errorf("same host should pass: %v", err)
	}
	// 带 path 的白名单
	pathPrefixes := []string{"https://cdn.example.com/uploads/"}
	if err := validateCoverURL("https://cdn.example.com/uploads/a.png", pathPrefixes); err != nil {
		t.Errorf("path prefix should pass: %v", err)
	}
	if err := validateCoverURL("https://cdn.example.com/other/a.png", pathPrefixes); err == nil {
		t.Error("path outside whitelist should be rejected")
	}
	// 无法解析为 URL 的配置项必须以 / 结尾
	slashRequired := []string{"not-a-url-prefix"}
	if err := validateCoverURL("not-a-url-prefix/x.png", slashRequired); err == nil {
		t.Error("non-URL prefix without trailing slash should not match")
	}
}
