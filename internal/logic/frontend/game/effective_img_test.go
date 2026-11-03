package game

import "testing"

// H5 列表：有自有封面时优先返回自有封面，否则回退三方图片。
func TestFrontendEffectiveImg(t *testing.T) {
	cases := []struct{ cover, third, want string }{
		{"https://c/a.png", "https://t/b.png", "https://c/a.png"},
		{"", "https://t/b.png", "https://t/b.png"},
		{" \t", "https://t/b.png", "https://t/b.png"},
	}
	for _, c := range cases {
		if got := effectiveImg(c.cover, c.third); got != c.want {
			t.Errorf("effectiveImg(%q,%q)=%q want %q", c.cover, c.third, got, c.want)
		}
	}
}
