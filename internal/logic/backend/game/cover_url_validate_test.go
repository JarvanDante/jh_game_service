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
