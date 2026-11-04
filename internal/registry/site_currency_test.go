package registry

import (
	"fmt"
	"testing"
)

type fakeLookup map[string]*SiteConfig

func (f fakeLookup) GetSiteConfig(code string) (*SiteConfig, error) {
	if c, ok := f[code]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("not found")
}

func (f fakeLookup) GetSiteCodeBySiteID(id int64) string {
	for code, c := range f {
		if int64(c.Base.SiteID) == id {
			return code
		}
	}
	return ""
}

func TestResolveSiteCurrency(t *testing.T) {
	jh := &SiteConfig{}
	jh.Base.SiteID, jh.Base.Currency = 1, "MMK(K)"
	empty := &SiteConfig{}
	empty.Base.SiteID = 2
	m := fakeLookup{"jh": jh, "ym": empty}

	if c, _ := resolveSiteCurrency("USD", "jh", 1, m); c != "USD" {
		t.Errorf("explicit currency must win, got %s", c)
	}
	if c, err := resolveSiteCurrency("", "jh", 0, m); err != nil || c != "MMK(K)" {
		t.Errorf("by code: %s %v", c, err)
	}
	if c, err := resolveSiteCurrency("", "", 1, m); err != nil || c != "MMK(K)" {
		t.Errorf("by id: %s %v", c, err)
	}
	for _, tc := range []struct {
		code string
		id   int64
	}{{"ym", 0}, {"", 0}, {"", 9}, {"zz", 0}} {
		if c, err := resolveSiteCurrency("", tc.code, tc.id, m); err == nil {
			t.Errorf("%v should fail, got %s", tc, c)
		}
	}
	if _, err := resolveSiteCurrency("", "jh", 0, nil); err == nil {
		t.Error("nil manager should fail")
	}
}
