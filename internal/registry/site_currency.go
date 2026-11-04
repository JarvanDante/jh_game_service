package registry

import "fmt"

// ResolveSiteCurrency returns the site's (legacy) currency value used on balances
// and game provider calls. Priority: currency passed in the request/metadata,
// then the published Consul site config (by site code, or by site id).
// It never falls back to a hard-coded currency.
func ResolveSiteCurrency(current, siteCode string, siteID int64) (string, error) {
	return resolveSiteCurrency(current, siteCode, siteID, GetDBConfigManager())
}

type siteConfigLookup interface {
	GetSiteConfig(siteCode string) (*SiteConfig, error)
	GetSiteCodeBySiteID(siteID int64) string
}

func resolveSiteCurrency(current, siteCode string, siteID int64, m siteConfigLookup) (string, error) {
	if current != "" {
		return current, nil
	}
	if m == nil {
		return "", fmt.Errorf("站点配置未初始化，无法确定币种")
	}
	if siteCode == "" && siteID > 0 {
		siteCode = m.GetSiteCodeBySiteID(siteID)
	}
	if siteCode == "" {
		return "", fmt.Errorf("无法确定站点，无法确定币种: site_id=%d", siteID)
	}
	cfg, err := m.GetSiteConfig(siteCode)
	if err != nil {
		return "", err
	}
	if cfg == nil || cfg.Base.Currency == "" {
		return "", fmt.Errorf("站点未配置币种: site=%s", siteCode)
	}
	return cfg.Base.Currency, nil
}
